package flatstruct

import (
	"fmt"
	"reflect"
)

/* documentation notes
Every Go identifier has to start with a letter, thus we will use numbers for special column headers in our output.
https://golang.org/ref/spec#identifier
*/

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
