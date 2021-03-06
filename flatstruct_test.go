package flatstruct

import (
	"reflect"
	"testing"

	"github.com/sauterp/flatstruct/util"
)

type AB struct {
	A string `json:"a"`
	B string `json:"b"`
}

// FlatUnflatTest TODO
func flatUnflatTest(t *testing.T, structured interface{}, flattened [][]string, headerBase, flatTestName, flatErr, unflatTestName, unflatErr string) {
	t.Run(flatTestName, func(t *testing.T) {
		myHeaders, myRows, err := FlattenBegin(headerBase, structured)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)

		if !util.CheckEq(t, flattened, myFlattened) {
			t.Errorf(flatErr)
		}
	})

	t.Run(unflatTestName, func(t *testing.T) {
		// create a pointer to an empty instance of structured to be filled
		myUnflattened := reflect.New(reflect.TypeOf(structured))
		myHeaderBase, err := Unflatten(flattened, myUnflattened.Interface())
		//myHeaderBase, err := Unflatten(flattened, &emptyStructured)
		if err != nil {
			// TODO
		}

		if !util.CheckObjEq(t, structured, myUnflattened.Elem().Interface()) {
			t.Errorf(unflatErr)
		}
		util.CheckBaseHeader(t, headerBase, myHeaderBase)
	})
}

// TestEmptyStruct TODO
func TestEmptyStruct(t *testing.T) {
	type Empty struct{}
	flattened := [][]string{{}}
	headerBase := "myBase"

	t.Run("FlattenBegin empty", func(t *testing.T) {
		var s Empty
		myHeaders, myRows, err := FlattenBegin(headerBase, s)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)
		if !util.CheckEq(t, flattened, myFlattened) {
			t.Errorf("Should be flattened empty")
		}
	})

	t.Run("Unflatten empty", func(t *testing.T) {
		var s Empty
		myHeaderBase, err := Unflatten(flattened, &s)
		if err != nil {
			// TODO
		}
		var sExpect Empty
		if !util.CheckObjEq(t, sExpect, s) {
			t.Errorf("Should be unflattened empty")
		}
		util.CheckBaseHeader(t, "", myHeaderBase)
	})
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
	headerBase := "myBase"
	flatUnflatTest(t, structured, flattened, headerBase, "Flatten flat stuct value", "Should be flattened flat struct value", "Flatten flat stuct value", "Should be unflattened flat struct value")
}

func TestStructWithEmptySlice(t *testing.T) {
	type WithEmptySlice struct {
		A          string `json:"a"`
		EmptySlice []int  `json:"empty_slice"`
	}
	structured := WithEmptySlice{
		A:          "aval",
		EmptySlice: nil,
	}
	flattened := [][]string{
		{"myBase.a"},
		{`"aval"`},
	}
	headerBase := "myBase"
	flatUnflatTest(t, structured, flattened, headerBase, "Flatten stuct value with emtpy slice", "Should be flattened struct value with emtpy slice", "Flatten stuct value with emtpy slice", "Should be unflattened struct value with emtpy slice")
}

func TestNestedStruct(t *testing.T) {
	type ABCD struct {
		AB AB     `json:"ab"`
		C  string `json:"c"`
		D  string `json:"d"`
	}
	t.Run("Simple nested struct", func(t *testing.T) {
		structured := ABCD{
			AB: AB{
				A: "aval",
				B: "bval",
			},
			C: "cval",
			D: "dval",
		}
		flattened := [][]string{
			{"myBase.ab.a", "myBase.ab.b", "myBase.c", "myBase.d"},
			{`"aval"`, `"bval"`, `"cval"`, `"dval"`},
		}
		headerBase := "myBase"
		flatUnflatTest(t, structured, flattened, headerBase, "Flatten flat stuct value with nested struct value with nested struct value", "Should be flattened flat struct value", "Flatten flat stuct value with nested struct value", "Should be unflattened flat struct value with nested struct value")
	})
	t.Run("More complicated nested struct", func(t *testing.T) {
		type GH struct {
			G int  `json:"g"`
			H bool `json:"h"`
		}
		type FGH struct {
			F  string `json:"f"`
			GH GH     `json:"gh"`
		}
		type ABCDEFGH struct {
			E    string `json:"e"`
			ABCD ABCD   `json:"abcd"`
			FGH  FGH    `json:"fgh"`
		}
		structured := ABCDEFGH{
			FGH: FGH{
				F: "fval",
				GH: GH{
					G: 987,
					H: true,
				},
			},
			ABCD: ABCD{
				AB: AB{
					A: "aval",
					B: "bval",
				},
				C: "cval",
				D: "dval",
			},
		}
		flattened := [][]string{
			{"myBase.e", "myBase.abcd.ab.a", "myBase.abcd.ab.b", "myBase.abcd.c", "myBase.abcd.d", "myBase.fgh.f", "myBase.fgh.gh.g", "myBase.fgh.gh.h"},
			{"\"\"", "\"aval\"", "\"bval\"", "\"cval\"", "\"dval\"", "\"fval\"", "987", "true"},
		}
		headerBase := "myBase"
		flatUnflatTest(t, structured, flattened, headerBase, "Flatten flat struct value ABCDEFG", "Should be flattened flat ABCDEFG", "Unflatten flat ABCDEFG", "Should be unflattened flat ABCDEFG")
	})
}

/* TODO support slice at base in Unflatten
func TestSlice(t *testing.T) {
	structured := []AB{
		{
			A: "1a",
			B: "1b",
		},
		{
			A: "2a",
			B: "2b",
		},
	}
	myHeaders, myRows, err := FlattenBegin("myBase", structured)
	if err != nil {
		// TODO
	}
	myFlattened := append([][]string{myHeaders}, myRows...)
	flattened := [][]string{
		{`[]myBase`, "[]myBase.a", "[]myBase.b"},
		{`0`, `"1a"`, `"1b"`},
		{`1`, `"2a"`, `"2b"`},
	}
	if !testEq(t, flattened, myFlattened) {
		t.Errorf("Should be slice of flat")
	}
	headerBase := "myBase"
	flatUnflatTest(t, structured, flattened, headerBase, "Flatten slice of flat struct values", "Should be flattened slice of AB", "Unflatten slice offlat struct values", "Should be unflattened slice of AB")
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
	myHeaders, myRows, err := FlattenBegin("myBase", s)
	if err != nil {
		// TODO
	}
	myFlattened := append([][]string{myHeaders}, myRows...)
	should := [][]string{
		{`[]myBase`, `[][]myBase`, "[][]myBase.a", "[][]myBase.b"},
		{`0`, `0`, `"11a"`, `"11b"`},
		{`0`, `1`, `"12a"`, `"12b"`},
		{`1`, `0`, `"21a"`, `"21b"`},
		{`1`, `1`, `"22a"`, `"22b"`},
	}
	if !testEq(t, should, myFlattened) {
		t.Errorf("Should be slice of slice")
	}
}
*/

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
		myHeaders, myRows, err := FlattenBegin("myBase", s)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)
		should := [][]string{
			{"myBase.[]abs", "myBase.[]abs.a", "myBase.[]abs.b", "myBase.[]cdes", "myBase.[]cdes.c", "myBase.[]cdes.d", "myBase.[]cdes.e"},
			{"0", "\"1a\"", "\"1b\"", "0", "23", "5.678", "true"},
			{"1", "\"2a\"", "\"2b\"", "1", "45", "789.123", "false"},
			{"", "", "", "2", "56", "345.2799", "false"},
		}
		if !util.CheckEq(t, should, myFlattened) {
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
		myHeaders, myRows, err := FlattenBegin("myBase", s)
		if err != nil {
			// TODO
		}
		myFlattened := append([][]string{myHeaders}, myRows...)
		should := [][]string{
			{"myBase.[]abs", "myBase.[]abs.a", "myBase.[]abs.b", "myBase.[]cdes", "myBase.[]cdes.c", "myBase.[]cdes.d", "myBase.[]cdes.e"},
			{"0", "\"1a\"", "\"1b\"", "0", "23", "5.678", "true"},
			{"1", "\"2a\"", "\"2b\"", "1", "45", "789.123", "false"},
			{"2", "\"3a\"", "\"3b\"", "", "", "", ""},
			{"3", "\"4a\"", "\"4b\"", "", "", "", ""},
		}
		if !util.CheckEq(t, should, myFlattened) {
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
