package flatstruct

import "testing"

func testEqStrSlice(t *testing.T, should, is []string) bool {
	if len(should) != len(is) {
		return false
	}

	for i := range should {
		if should[i] != is[i] {
			return false
		}
	}

	return true
}

func testEq(t *testing.T, should, is [][]string) bool {
	if len(should) != len(is) {
		t.Errorf("\nlen(should) != len(is)\n\nshould: %v\n\nis: %v\n", should, is)
		return false
	}

	for i := range should {
		if !testEqStrSlice(t, should[i], is[i]) {
			t.Errorf("\nshould: %v\n\n is: %v\n\n mismatch at index i: %d", should, is, i)
			return false
		}
	}

	return true
}

func TestEmptyStruct(t *testing.T) {
	type Empty struct{}
	var s Empty
	is := Flatten(s)
	should := [][]string{{}, {}}
	if !testEq(t, should, is) {
		t.Errorf("Should be empty")
	}
}

func TestFlatStruct(t *testing.T) {
	type Flat struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	s := Flat{
		A: "aval",
		B: "bval",
	}
	is := Flatten(s)
	should := [][]string{
		{"a", "b"},
		{"aval", "bval"},
	}
	if !testEq(t, should, is) {
		t.Errorf("Should be flat")
	}
}
