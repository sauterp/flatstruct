// Package flatstruct provides functionality to flatten any JSON encodable Go value into a table and unflatten again. This can be useful to export data from a Go application in the form of a spreadsheet such as Excel or CSV.
package flatstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// FlattenSlice flattens a slice. Each slice element will result in one additional row.
// TODO support slices of slices.
func FlattenSlice(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	sValue := reflect.ValueOf(s)

	if sValue.Len() > 0 {
		headers = append(headers, headerBase)
		var newHeaders []string
		for i := 0; i < sValue.Len(); i++ {
			var newRows [][]string
			var err error
			newHeaders, newRows, err = Flatten(headerBase, sValue.Index(i).Interface())
			if err != nil {
				return nil, nil, err
			}
			iJSON, err := json.Marshal(i)
			if err != nil {
				return nil, nil, err
			}
			for r := 0; r < len(newRows); r++ {
				newRows[r] = append([]string{string(iJSON)}, newRows[r]...)
			}
			rows = append(rows, newRows...)
		}
		headers = append(headers, newHeaders...)
	}

	return headers, rows, nil
}

// FillAndAppend ensures that both tables have the same number of rows, by filling up the one with fewer rows. After that it will concatenate them horizontally.
func FillAndAppend(headers, newHeaders []string, rows, newRows [][]string) ([]string, [][]string) {
	rowsLen := len(rows)
	newRowsLen := len(newRows)
	var maxLen int
	var tableToFill *[][]string
	if rowsLen > newRowsLen {
		maxLen = rowsLen
		tableToFill = &newRows
	} else {
		maxLen = newRowsLen
		tableToFill = &rows
	}
	lenTableToFill := len(*tableToFill)
	emptyRowsToAdd := maxLen - lenTableToFill
	var emptyRowLen int
	if lenTableToFill == 0 {
		emptyRowLen = 0
	} else {
		emptyRowLen = len((*tableToFill)[0])
	}
	for r := 0; r < emptyRowsToAdd; r++ {
		*tableToFill = append(*tableToFill, make([]string, emptyRowLen))
	}
	for r := 0; r < maxLen; r++ {
		rows[r] = append(rows[r], newRows[r]...)
	}

	headers = append(headers, newHeaders...)

	return headers, rows
}

// FlattenStruct flattens a struct value. Each struct field will result in one additional column unless that field again is a struct value.
func FlattenStruct(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	sValue := reflect.ValueOf(s)
	sType := reflect.TypeOf(s)
	if sValue.Kind() == reflect.Struct &&
		sType != reflect.TypeOf(time.Time{}) {
		nFields := sType.NumField()
		for i := 0; i < nFields; i++ {
			fieldVal := sValue.Field(i)

			// TODO Use strings.Split to ignore tag options
			// https://stackoverflow.com/questions/55879028/golang-get-structs-field-name-by-json-tag
			tag := sType.Field(i).Tag.Get("json")

			var newHeaders []string
			var newRows [][]string
			if fieldVal.Kind() == reflect.Slice {
				// column for slice indices
				newHeaderBase := fmt.Sprintf("%s.[]%s", headerBase, tag)

				newHeaders, newRows, err = FlattenSlice(newHeaderBase, fieldVal.Interface())
			} else {
				newheaderBase := fmt.Sprintf("%s.%s", headerBase, tag)

				newHeaders, newRows, err = FlattenStruct(newheaderBase, fieldVal.Interface())
			}
			if err != nil {
				return nil, nil, err
			}

			headers, rows = FillAndAppend(headers, newHeaders, rows, newRows)
		}
	} else {
		headers, rows, err = FlattenDefault(headerBase, s)
	}

	return headers, rows, nil
}

// FlattenDefault encodes s as a JSON value.
func FlattenDefault(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	sValue := reflect.ValueOf(s)
	headers = []string{headerBase}

	var bytes []byte
	var enc interface{}
	if (sValue == reflect.Value{}) || (sValue.Kind() == reflect.Interface && (sValue.IsNil() || sValue.IsZero())) {
		enc = nil
	} else {
		enc = sValue.Interface()
	}
	bytes, err = json.Marshal(enc)
	if err != nil {
		return nil, nil, err
	}
	rows = [][]string{{string(bytes)}}

	return headers, rows, nil
}

// Flatten flattens an arbitrary json encodable struct value to a table, where slice elements allocate rows and struct fields allocate columns.
// TODO Order headers by nesting depth
func Flatten(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	// TODO error if headerBase starts with number or is not valid Go identifier
	sValue := reflect.ValueOf(s)
	//	sType := reflect.TypeOf(s)
	switch sValue.Kind() {
	case reflect.Slice:
		// TODO handle case where slice of slice of slice ... is the first header
		headers, rows, err = FlattenSlice("[]"+headerBase, s)

	case reflect.Struct:
		headers, rows, err = FlattenStruct(headerBase, s)

	default:
		headers, rows, err = FlattenDefault(headerBase, s)
	}
	if err != nil {
		return nil, nil, err
	}

	return headers, rows, nil
}

// TODO what happens if we don't find the field tag?
// TODO return error if field is not found.
func getFieldIndexByTag(t reflect.Type, tag string) (int, error) {
	for f := 0; f < t.NumField(); f++ {
		if t.Field(f).Tag.Get("json") == tag {
			return f, nil
		}
	}
	return 0, fmt.Errorf("Cannot find struct field with tag: %s", tag)
}

// Retrieve returns a pointer to a slice value. If no corresponding element exists for index, the slice will be appended with zero elements up to that index before returning the element.
func Retrieve(sliceVal reflect.Value, index int) reflect.Value {
	sliceVal = sliceVal.Elem()
	vLen := sliceVal.Len()
	if vLen-1 < index {
		necessaryLen := index - vLen + 1
		appSlice := reflect.MakeSlice(sliceVal.Type(), necessaryLen, necessaryLen)
		for i := 0; i < appSlice.Len(); i++ {
			sliceVal.Set(reflect.Append(sliceVal, appSlice.Index(i)))
		}
	}
	return sliceVal.Index(index).Addr()
}

// DescendTreeAndEncode finds the leaf in the struct value s(imagine it like a tree) pointed to by the header and the sliceIndices and encodes the rowEl in that field value.
// s should be a pointer.
func DescendTreeAndEncode(s interface{}, header string, sliceIndices []int, rowEl string) error {
	// descend the "type tree" to the leaf pointed to by this header and set its value
	currentValueNode := reflect.ValueOf(s).Elem()
	currentTypeNode := currentValueNode.Type()

	split := strings.Split(header, ".")
	// TODO support slice at base

	sliceIndex := 0
	for si := 1; si < len(split); si++ {
		fieldTag := split[si]
		if len(fieldTag) >= 2 && fieldTag[:2] == "[]" {
			fieldTag = fieldTag[2:]
			fieldIndex, err := getFieldIndexByTag(currentTypeNode, fieldTag)
			if err != nil {
				return err
			}
			currentValueNode = currentValueNode.Field(fieldIndex)
			currentValueNode = Retrieve(currentValueNode.Addr(), sliceIndices[sliceIndex])
			sliceIndex++
			currentValueNode = currentValueNode.Elem()
			currentTypeNode = currentValueNode.Type()
		} else {
			fieldIndex, err := getFieldIndexByTag(currentTypeNode, fieldTag)
			if err != nil {
				return err
			}
			currentValueNode = currentValueNode.Field(fieldIndex)
			currentTypeNode = currentValueNode.Type()
		}
	}

	//encode
	err := json.Unmarshal([]byte(rowEl), currentValueNode.Addr().Interface())
	if err != nil {
		return err
	}
	return nil
}

// CheckIsSliceIndex returns true if the header points to a slice. This is the case if the last field name in the headaer is preceded by '[]'.
func CheckIsSliceIndex(header string) bool {
	split := strings.Split(header, ".")
	lastEl := split[len(split)-1]
	if len(lastEl) >= 2 && lastEl[:2] == "[]" {
		return true
	}
	return false
}

// Unflatten unflattens the table given in f into s. The first row is assumed to contain the headers.
func Unflatten(f [][]string, s interface{}) (headerBase string, err error) {
	if len(f) < 1 || len(f[0]) < 1 {
		// TODO is this correct?
		return "", nil
	}

	headers := f[0]
	rows := f[1:]

	// Create slice index map
	sliceIndexCols := make(map[string]int, 0)
	for h := 0; h < len(headers); h++ {
		header := headers[h]
		if CheckIsSliceIndex(header) {
			sliceIndexCols[header] = h
		}
	}

	for h := 0; h < len(headers); h++ {
		// lookup index columns
		var indexCols []int
		header := headers[h]
		split := strings.Split(header, ".")
		base := split[0]
		for s := 1; s < len(split); s++ {
			if v, ok := sliceIndexCols[base]; ok {
				indexCols = append(indexCols, v)
			}
			base = base + "." + split[s]
		}

		if !CheckIsSliceIndex(header) {
			for r := 0; r < len(rows); r++ {
				// retrieve slice indices
				if rows[r][h] != "" {
					var sliceIndices []int
					for i := 0; i < len(indexCols); i++ {
						rowEl := rows[r][indexCols[i]]
						var index int
						err := json.Unmarshal([]byte(rowEl), &index)
						if err != nil {
							return "", err
						}
						sliceIndices = append(sliceIndices, index)
					}

					err := DescendTreeAndEncode(s, header, sliceIndices, rows[r][h])
					if err != nil {
						return "", err
					}
				}
			}
		}
	}

	firstHeader := headers[0]
	firstHeaderBase := strings.Split(firstHeader, ".")[0]
	if CheckIsSliceIndex(firstHeader) {
		headerBase = firstHeaderBase[2:]
	} else {
		headerBase = firstHeaderBase
	}

	return headerBase, nil
}
