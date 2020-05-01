package flatstruct

import (
	"fmt"
	"reflect"
)

/* documentation notes
Every Go identifier has to start with a letter, thus we will use numbers for special column headers in our output.
https://golang.org/ref/spec#identifier

encode maps as JSON
*/

// CompNRowsCols computes the number of rows and columns necessary to represent object s in a table.
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
			nrows += fnrows
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
func Flatten(s interface{}) [][]string {
	var flat [][]string
	sv := reflect.ValueOf(s)
	fmt.Println(sv)

	flat = append(flat, make([]string, 0))
	flat = append(flat, make([]string, 0))
	sf := reflect.TypeOf(s)
	for i := 0; i < sf.NumField(); i++ {
		val := reflect.ValueOf(s).Field(i)
		t := sf.Field(i).Tag.Get("json")
		fmt.Println("t ", t)
		flat[0] = append(flat[0], t)
		flat[1] = append(flat[1], val.String())
	}
	return flat
}
