package flatstruct

import (
	"fmt"
	"testing"
)

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

func printTable(table [][]string) string {
	sTable := ""
	for _, t := range table {
		sTable = sTable + "\n" + fmt.Sprintf("%#v", t)
	}
	return sTable
}

func testEq(t *testing.T, should, is [][]string) bool {
	if len(should) != len(is) {
		t.Errorf("\nlen(should) != len(is)\n\nshould: %v\n\nis: %v\n", printTable(should), printTable(is))
		return false
	}

	for i := range should {
		if !testEqStrSlice(t, should[i], is[i]) {
			t.Errorf("\nshould: %v\n\n is: %v\n\n mismatch at index i: %d", printTable(should), printTable(is), i)
			return false
		}
	}

	return true
}

func testObjEq(t *testing.T, should, is interface{}) bool {
	if should != is {
		t.Errorf("\nshould: %#v\n\n is: %#v", should, is)
		return false
	}

	return true
}

func TestEmptyStruct(t *testing.T) {
	type Empty struct{}
	flattened := [][]string{{}}

	t.Run("Flatten empty", func(t *testing.T) {
		var s Empty
		myHeaders, myRows, err := Flatten("myBase", s)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)
		if !testEq(t, flattened, myFlattened) {
			t.Errorf("Should be flattened empty")
		}
	})

	t.Run("Unflatten empty", func(t *testing.T) {
		var s Empty
		err := Unflatten(flattened, &s)
		if err != nil {
			// TODO
		}
		var sExpect Empty
		if !testObjEq(t, sExpect, s) {
			t.Errorf("Should be unflattened empty")
		}
	})
}

type AB struct {
	A string `json:"a"`
	B string `json:"b"`
}

func TestFlatStruct(t *testing.T) {
	structured := AB{
		A: "aval",
		B: "bval",
	}
	flattened := [][]string{
		{"myBase.a", "myBase.b"},
		{`"aval"`, `"bval"`},
	}
	t.Run("Flatten flat stuct value", func(t *testing.T) {
		myHeaders, myRows, err := Flatten("myBase", structured)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)

		if !testEq(t, flattened, myFlattened) {
			t.Errorf("Should be flattened flat struct value")
		}
	})

	t.Run("Flatten flat stuct value", func(t *testing.T) {
		var myUnflattened AB
		err := Unflatten(flattened, &myUnflattened)
		if err != nil {
			// TODO
		}

		if !testObjEq(t, structured, myUnflattened) {
			t.Errorf("Should be unflattened flat struct value")
		}
	})
}

func TestSlice(t *testing.T) {
	s := []AB{
		{
			A: "1a",
			B: "1b",
		},
		{
			A: "2a",
			B: "2b",
		},
	}
	myHeaders, myRows, err := Flatten("myBase", s)
	if err != nil {
		// TODO
	}
	myFlattened := append([][]string{myHeaders}, myRows...)
	should := [][]string{
		{`0`, "myBase.a", "myBase.b"},
		{`0`, `"1a"`, `"1b"`},
		{`1`, `"2a"`, `"2b"`},
	}
	if !testEq(t, should, myFlattened) {
		t.Errorf("Should be slice of flat")
	}
}

func TestSliceOfSlice(t *testing.T) {
	s := [][]AB{
		{
			{
				A: "11a",
				B: "11b",
			},
			{
				A: "12a",
				B: "12b",
			},
		},

		{
			{
				A: "21a",
				B: "21b",
			},
			{
				A: "22a",
				B: "22b",
			},
		},
	}
	myHeaders, myRows, err := Flatten("myBase", s)
	if err != nil {
		// TODO
	}
	myFlattened := append([][]string{myHeaders}, myRows...)
	should := [][]string{
		{`0`, `0`, "myBase.a", "myBase.b"},
		{`0`, `0`, `"11a"`, `"11b"`},
		{`0`, `1`, `"12a"`, `"12b"`},
		{`1`, `0`, `"21a"`, `"21b"`},
		{`1`, `1`, `"22a"`, `"22b"`},
	}
	if !testEq(t, should, myFlattened) {
		t.Errorf("Should be slice of slice")
	}
}

func TestStructWithTwoSlices(t *testing.T) {
	type CDE struct {
		C int     `json:"c"`
		D float64 `json:"d"`
		E bool    `json:"e"`
	}
	type TwoSlices struct {
		ABs  []AB  `json:"abs"`
		CDEs []CDE `json:"cdes"`
	}

	t.Run("Longer second slice", func(t *testing.T) {
		s := TwoSlices{
			ABs: []AB{
				{
					A: "1a",
					B: "1b",
				},
				{
					A: "2a",
					B: "2b",
				},
			},
			CDEs: []CDE{
				{
					C: 23,
					D: 5.678,
					E: true,
				},
				{
					C: 45,
					D: 789.123,
					E: false,
				},
				{
					C: 56,
					D: 345.2799,
					E: false,
				},
			},
		}
		myHeaders, myRows, err := Flatten("myBase", s)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)
		should := [][]string{
			{"0", "myBase.abs.a", "myBase.abs.b", "0", "myBase.cdes.c", "myBase.cdes.d", "myBase.cdes.e"},
			{"0", "\"1a\"", "\"1b\"", "0", "23", "5.678", "true"},
			{"1", "\"2a\"", "\"2b\"", "1", "45", "789.123", "false"},
			{"", "", "", "2", "56", "345.2799", "false"},
		}
		if !testEq(t, should, myFlattened) {
			t.Errorf("Should be two slices where the first one has some empty rows in the table representation")
		}
	})

	t.Run("Longer first slice", func(t *testing.T) {
		s := TwoSlices{
			ABs: []AB{
				{
					A: "1a",
					B: "1b",
				},
				{
					A: "2a",
					B: "2b",
				},
				{
					A: "3a",
					B: "3b",
				},
				{
					A: "4a",
					B: "4b",
				},
			},
			CDEs: []CDE{
				{
					C: 23,
					D: 5.678,
					E: true,
				},
				{
					C: 45,
					D: 789.123,
					E: false,
				},
			},
		}
		myHeaders, myRows, err := Flatten("myBase", s)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)
		should := [][]string{
			{"0", "myBase.abs.a", "myBase.abs.b", "0", "myBase.cdes.c", "myBase.cdes.d", "myBase.cdes.e"},
			{"0", "\"1a\"", "\"1b\"", "0", "23", "5.678", "true"},
			{"1", "\"2a\"", "\"2b\"", "1", "45", "789.123", "false"},
			{"2", "\"3a\"", "\"3b\"", "", "", "", ""},
			{"3", "\"4a\"", "\"4b\"", "", "", "", ""},
		}
		if !testEq(t, should, myFlattened) {
			t.Errorf("Should be two slices where the second one has some empty rows in the table representation")
		}
	})
}

func TestMap(t *testing.T) {
	// TODO
}

func TestInterface(t *testing.T) {
	// TODO
}
