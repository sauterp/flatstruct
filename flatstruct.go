package flatstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

/* documentation notes
Every Go identifier has to start with a letter, thus we will use numbers for special column headers in our output.
https://golang.org/ref/spec#identifier

encode maps as JSON
*/

// CompNRowsCols computes the number of rows and columns necessary to represent object s in a table.
// TODO test
func CompNRowsCols(s interface{}) (nrows, ncols int) {
	nrows = 0
	ncols = 0

	sValue := reflect.ValueOf(s)
	switch sValue.Kind() {
	case reflect.Slice:
		// one table row for each slice element
		sLen := sValue.Len()
		nrows += sLen
		// extra column for the element indices
		ncols++

		for i := 0; i < sLen; i++ {
			fnrows, fncols := CompNRowsCols(sValue.Interface().([]interface{})[i])
			if nrows < fnrows {
				nrows = fnrows
			}
			ncols += fncols
		}

	case reflect.Struct:
		for i := 0; i < sValue.NumField(); i++ {
			field := sValue.Field(i)

			fnrows, fncols := CompNRowsCols(field.Interface())
			nrows += fnrows
			ncols += fncols
		}

	default:
		nrows = 0
		ncols = 1

	}

	return nrows, ncols
}

// FlattenSlice TODO
func FlattenSlice(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	sValue := reflect.ValueOf(s)

	if sValue.Len() > 0 {
		headers = append(headers, headerBase)
		var newHeaders []string
		for i := 0; i < sValue.Len(); i++ {
			var newRows [][]string
			var err error
			newHeaders, newRows, err = FlattenBegin(headerBase, sValue.Index(i).Interface())
			if err != nil {
				// TODO
			}
			iJSON, err := json.Marshal(i)
			if err != nil {
				// TODO
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

// FlattenStruct TODO
func FlattenStruct(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	newHeaders, newRows, err := Flatten(headerBase, s)
	if err != nil {
		// TODO
	}

	headers, rows = FillAndAppend(headers, newHeaders, rows, newRows)

	return headers, rows, nil
}

// FillAndAppend TODO
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

// Flatten TODO
func Flatten(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	// TODO error if headerBase starts with number or is not valid Go identifier
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
				// TODO support slice of slice
				newHeaderBase := fmt.Sprintf("%s.[]%s", headerBase, tag)

				newHeaders, newRows, err = FlattenSlice(newHeaderBase, fieldVal.Interface())
			} else {
				newheaderBase := fmt.Sprintf("%s.%s", headerBase, tag)

				newHeaders, newRows, err = Flatten(newheaderBase, fieldVal.Interface())
			}
			if err != nil {
				// TODO
			}

			headers, rows = FillAndAppend(headers, newHeaders, rows, newRows)
		}
	} else {
		headers, rows, err = FlattenDefault(headerBase, s)
	}

	return headers, rows, nil
}

// FlattenDefault TODO
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
		// TODO
	}
	rows = [][]string{{string(bytes)}}

	return headers, rows, nil
}

// FlattenBegin TODO
// TODO Order headers by nesting depth
func FlattenBegin(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	// TODO error if headerBase starts with number or is not valid Go identifier
	sValue := reflect.ValueOf(s)
	//	sType := reflect.TypeOf(s)
	switch sValue.Kind() {
	case reflect.Slice:
		// TODO handle case where slice of slice of slice ... is the first header
		headers, rows, err = FlattenSlice("[]"+headerBase, s)

	case reflect.Struct:
		headers, rows, err = Flatten(headerBase, s)

	default:
		headers, rows, err = FlattenDefault(headerBase, s)
	}
	if err != nil {
		// TODO
	}

	return headers, rows, nil
}

// TODO what happens if we don't find the field tag?
func getFieldIndexByTag(t reflect.Type, tag string) int {
	for f := 0; f < t.NumField(); f++ {
		if t.Field(f).Tag.Get("json") == tag {
			return f
		}
	}
	return -1
}

// Retrieve TODO
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

// DescendTreeAndEncode TODO
func DescendTreeAndEncode(s interface{}, header string, sliceIndices []int, rowEl string) {
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
			fieldIndex := getFieldIndexByTag(currentTypeNode, fieldTag)
			currentValueNode = currentValueNode.Field(fieldIndex)
			currentValueNode = Retrieve(currentValueNode.Addr(), sliceIndices[sliceIndex])
			sliceIndex++
			currentValueNode = currentValueNode.Elem()
			currentTypeNode = currentValueNode.Type()
		} else {
			fieldIndex := getFieldIndexByTag(currentTypeNode, fieldTag)
			currentValueNode = currentValueNode.Field(fieldIndex)
			currentTypeNode = currentValueNode.Type()
		}
	}

	//encode
	err := json.Unmarshal([]byte(rowEl), currentValueNode.Addr().Interface())
	if err != nil {
		// TODO
		panic(err)
	}
}

// CheckIsSliceIndex TODO
func CheckIsSliceIndex(header string) bool {
	split := strings.Split(header, ".")
	lastEl := split[len(split)-1]
	if len(lastEl) >= 2 && lastEl[:2] == "[]" {
		return true
	}
	return false
}

// Unflatten TODO
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
							// TODO
							panic(err)
						}
						sliceIndices = append(sliceIndices, index)
					}

					// descend tree

					// encode value
					DescendTreeAndEncode(s, header, sliceIndices, rows[r][h])
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
