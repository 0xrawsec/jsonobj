package jsonobj

import (
	"encoding/json"
	"testing"

	"github.com/0xrawsec/toast"
)

type MyStruct struct {
	Field1 string
	Field2 int
	EmbeddedStruct
}

type Dummy struct {
	A string
	B int64
}

type EmbeddedStruct struct {
	Field3      float64
	IntSlice    []int
	StructSlice []Dummy
	M           map[string]int
}

func jsonStr(v any) string {
	if b, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

func TestObject(t *testing.T) {
	t.Parallel()
	tt := toast.FromT(t)

	o := New()

	tt.Assert(o.IsEmpty())

	o.SetField("test", 42)
	o.SetField("toast", 42)
	tt.Assert(o.GetField("test").(int) == 42)

	_, err := json.Marshal(&o)
	tt.CheckErr(err)

	// triggers index out of range bug
	o.SetField("toast", 44)

	// the field is already there so SetField will modify it
	o.SetField("test", 43)
	tt.Assert(o.HasField("test"))
	tt.Assert(o.GetField("test").(int) == 43)
	tt.Assert(o.GetField("toast").(int) == 44)

	tt.Assert(!o.HasField("unknown"))
}

func TestNestedObject(t *testing.T) {
	t.Parallel()
	tt := toast.FromT(t)

	s := MyStruct{
		Field1: "test",
		Field2: 42,
		EmbeddedStruct: EmbeddedStruct{Field3: 42, IntSlice: []int{42}, StructSlice: []Dummy{{
			A: "Ola",
			B: 4242,
		}},
			M: map[string]int{"test": 42},
		},
	}

	o := FromStruct(s)
	b, err := json.Marshal(&o)
	tt.CheckErr(err)
	tt.Log(string(b))

	js, err := json.Marshal(&s)
	tt.CheckErr(err)
	tt.Log(string(js))
	// json object should be equal with json serialization
	tt.Assert(string(b) == string(js))

	new := MyStruct{}
	tt.CheckErr(json.Unmarshal(b, &new))
	// serialization after deserialization should not change anything
	tt.Assert(jsonStr(new) == jsonStr(s))
}

func TestNestedObjectLowercase(t *testing.T) {
	t.Parallel()
	tt := toast.FromT(t)

	s := MyStruct{
		Field1: "test",
		Field2: 42,
		EmbeddedStruct: EmbeddedStruct{Field3: 42, IntSlice: []int{42}, StructSlice: []Dummy{{
			A: "Ola",
			B: 4242,
		}},
			M: map[string]int{"test": 42},
		},
	}

	o := FromStructWithOptions(s, Options{LowerCase})
	b, err := json.Marshal(&o)
	tt.CheckErr(err)
	tt.Log(string(b))

	new := MyStruct{}
	tt.CheckErr(json.Unmarshal(b, &new))
	// serialization after deserialization should not change anything
	tt.Assert(jsonStr(new) == jsonStr(s))
}

func TestConvertSlice(t *testing.T) {
	t.Parallel()
	tt := toast.FromT(t)

	s := []interface{}{
		[]int{42},
		map[string]string{"toast": "test"},
	}

	o := New()

	tt.Log(jsonStr(o.ConvertSlice(s)))
}

func TestFromMap(t *testing.T) {
	t.Parallel()
	tt := toast.FromT(t)

	m := map[string]any{
		"B": "test",
		"A": 42,
		"C": []int{1, 2, 3}}

	o := FromMap(m)

	// key order is not guaranteed
	s := jsonStr(o)
	tt.Log(s)
	new := new(map[string]any)
	json.Unmarshal([]byte(s), &new)

	tt.Assert(jsonStr(m) == jsonStr(new))
}

func TestSnakeCase(t *testing.T) {
	t.Parallel()

	tt := toast.FromT(t)

	tt.Assert(camelToSnake("TestTest") == "test_test", "Unexpected snake case")
	tt.Assert(camelToSnake("TestTEST") == "test_test", "Unexpected snake case")
	tt.Assert(camelToSnake("OneTWOThree") == "one_two_three", "Unexpected snake case")
	tt.Assert(camelToSnake("One2Three") == "one_2_three", "Unexpected snake case")
	tt.Assert(camelToSnake("One23") == "one_23", "Unexpected snake case")
	tt.Assert(camelToSnake("1Step2Step") == "1_step_2_step", "Unexpected snake case")
	tt.Assert(camelToSnake("123") == "123", "Unexpected snake case")
}
