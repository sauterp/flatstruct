// Package util provides equality checks useful for writing tests for flatstruct.
package util

import (
	"reflect"
	"testing"
)

// CheckEqStrSlice checks the equality of two slices of strings.
func CheckEqStrSlice(t *testing.T, should, is []string) bool {
	if len(should) != len(is) {
		return false
	}

	for i := range should {
		if should[i] != is[i] {
			t.Errorf("\nshould[i]: %v\n\n is[i]: %v\n", should[i], is[i])
			return false
		}
	}

	return true
}

// CheckEq checks the equality of two tables(slice of slices of strings).
func CheckEq(t *testing.T, should, is [][]string) bool {
	if len(should) != len(is) {
		t.Errorf("\nlen(should) != len(is)\n\nlen(should): %d\n\nlen(is): %d\n", len(should), len(is))
		return false
	}

	for i := range should {
		if !CheckEqStrSlice(t, should[i], is[i]) {
			t.Errorf("\nshould[i]: %v\n\n is[i]: %v\n", should[i], is[i])
			return false
		}
	}

	return true
}

// CheckObjEq checks the equality of two interface values.
func CheckObjEq(t *testing.T, should, is interface{}) bool {
	if !reflect.DeepEqual(should, is) {
		t.Errorf("\nshould: %#v\n\n is: %#v", should, is)
		return false
	}

	return true
}

// CheckBaseHeader checks the equality of two strings which should be the same header.
func CheckBaseHeader(t *testing.T, should, is string) {
	if should != is {
		t.Errorf("base header mismatch:\n\nshould:\n\t%s\nis:\n\t%s\n", should, is)
	}
}
