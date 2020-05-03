package flatstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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
	headers = append(headers, headerBase)

	sValue := reflect.ValueOf(s)
	var newHeaders []string
	for i := 0; i < sValue.Len(); i++ {
		var newRows [][]string
		var err error
		newHeaders, newRows, err = Flatten(headerBase, sValue.Index(i).Interface())
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
	return headers, rows, nil
}

// Flatten TODO
// TODO Order headers by nesting depth
func Flatten(headerBase string, s interface{}) (headers []string, rows [][]string, err error) {
	// TODO error if headerBase starts with number or is not valid Go identifier
	sValue := reflect.ValueOf(s)
	sType := reflect.TypeOf(s)
	switch sValue.Kind() {
	case reflect.Slice:
		newHeaders, newRows, err := FlattenSlice("[]"+headerBase, s)
		if err != nil {
			// TODO
		}
		rows = append(rows, newRows...)
		headers = append(headers, newHeaders...)

	case reflect.Struct:
		nFields := sType.NumField()
		for i := 0; i < nFields; i++ {
			fieldVal := sValue.Field(i)

			// TODO Use strings.Split to ignore tag options
			// https://stackoverflow.com/questions/55879028/golang-get-structs-field-name-by-json-tag
			tag := sType.Field(i).Tag.Get("json")

			if fieldVal.Kind() == reflect.Slice {
				// column for slice indices
				newHeaderBase := fmt.Sprintf("[]%s.%s", headerBase, tag)
				newHeaders, newRows, err := FlattenSlice(newHeaderBase, fieldVal.Interface())
				if err != nil {
					// TODO
				}
				rows = append(rows, newRows...)
				headers = append(headers, newHeaders...)
			} else {
				newheaaderBase := fmt.Sprintf("%s.%s", headerBase, tag)

				newHeaders, newRows, err := Flatten(newheaaderBase, fieldVal.Interface())
				if err != nil {
					// TODO
				}

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
			}
		}

	default:
		headers = []string{headerBase}

		bytes, err := json.Marshal(sValue.Interface())
		if err != nil {
			// TODO
		}
		rows = [][]string{{string(bytes)}}

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

// Unflatten TODO
func Unflatten(f [][]string, s interface{}) (headerBase string, err error) {
	if len(f) < 1 || len(f[0]) < 1 {
		// TODO is this correct?
		return "", nil
	}

	sValue := reflect.ValueOf(s).Elem()
	sType := reflect.TypeOf(s).Elem()

	headers := f[0]
	rows := f[1:]

	for h := 0; h < len(headers); h++ {
		// TODO support unordered slices
		header := headers[h]
		// TODO proper way to label the index in the headers
		isSliceIndex := header[:2] == "[]"
		if isSliceIndex {
			header = header[2:]
			// headers[0] is a slice index
			sliceLen := 0
			for r := 0; r < len(rows); r++ {
				// build slice
				index := rows[r][0]
				if index == "" {
					// slice is empty or has no more elemtns
					break
				} else {
					var i int
					err := json.Unmarshal([]byte(index), &i)
					if err != nil {
						// TODO
					}
					if i > sliceLen {
						sliceLen = i
					}
				}
			}
			sliceLen++ // for correct length
		}

		// check headerBase
		currentHeaderBase := strings.SplitN(header, ".", 2)[0]
		if headerBase == "" {
			headerBase = currentHeaderBase
		} else if currentHeaderBase != headerBase {
			// TODO
		}

		if !isSliceIndex {
			split := strings.Split(header, ".")
			// descend the "type tree" to the leaf pointed to by this header and set its value
			currentValueNode := sValue
			currentTypeNode := sType
			var fieldIndex int
			for s := 1; s < len(split); s++ {
				fieldTag := split[s]
				fieldIndex = getFieldIndexByTag(currentTypeNode, fieldTag)
				currentValueNode = currentValueNode.Field(fieldIndex)
				currentTypeNode = currentValueNode.Type()
			}
			if len(rows) > 1 && rows[1][h] != "" {
				// this column is part of a slice
				for r := 0; r < len(rows); r++ {
					// TODO test whether empty values are handeled correctly
					rowEl := rows[r][h]
					if rowEl != "" {
						currentValueNode = reflect.Append(currentValueNode, reflect.New(currentTypeNode.Elem()))
						err := json.Unmarshal([]byte(rowEl), currentValueNode.Addr().Index(r).Interface())
						if err != nil {
							// TODO
							panic(err)
						}
					} else {
						break
					}
				}
			} else {
				err := json.Unmarshal([]byte(rows[0][h]), currentValueNode.Addr().Interface())
				if err != nil {
					// TODO
					panic(err)
				}
			}
		}
	}

	switch sValue.Kind() {
	case reflect.Slice:

	case reflect.Struct:

	default:

	}

	return headerBase, nil
}
