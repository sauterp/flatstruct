package flatstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
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

// Flatten TODO
func Flatten(heaaderBase string, s interface{}) (headers []string, rows [][]string, err error) {
	sValue := reflect.ValueOf(s)
	sType := reflect.TypeOf(s)
	switch sValue.Kind() {
	case reflect.Slice:
		// column for slice indices
		headers = append(headers, "0")
		var newHeaders []string
		for i := 0; i < sValue.Len(); i++ {
			var newRows [][]string
			var err error
			newHeaders, newRows, err = Flatten(heaaderBase, sValue.Index(i).Interface())
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

	case reflect.Struct:
		nFields := sType.NumField()
		for i := 0; i < nFields; i++ {
			tag := sType.Field(i).Tag.Get("json")
			newheaaderBase := fmt.Sprintf("%s.%s", heaaderBase, tag)

			fieldVal := sValue.Field(i)
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

	default:
		headers = []string{heaaderBase}

		bytes, err := json.Marshal(sValue.Interface())
		if err != nil {
			// TODO
		}
		rows = [][]string{{string(bytes)}}

	}

	return headers, rows, nil
}
