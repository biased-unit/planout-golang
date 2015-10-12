package planout

import (
	"testing"
	"io/ioutil"
	"encoding/json"
)

type Inner struct {
	Value string
}

type Outer struct {
	Inner *Inner
}

type NestedStruct struct {
	Outer *Outer
}

func TestNestedIndex(t *testing.T) {
	data, err := ioutil.ReadFile("test/nested_index.json")
	if err != nil {
		t.Fatal(err)
	}
	var js map[string]interface{}
	err = json.Unmarshal(data, &js)
	if (err != nil) {
		t.Fatal(err)
	}

	inputs := make(map[string]interface{})
	inputs["s"] = &NestedStruct{
		Outer: &Outer{
			Inner: &Inner{
				Value: "foo",
			},
		},
	}

	exp := &Interpreter{
		Name: "nested_test",
		Salt: "salt123",
		Evaluated: false,
		Inputs: inputs,
		Outputs: make(map[string]interface{}),
		Code: js,
	}

	if _, ok := exp.Run(); !ok {
		t.Fatal("Failed to run experiment")
	}

	if (exp.Outputs["out"] != "foo") {
		t.Fail()
	}
}

type StructWithArray struct {
	Array []*int
}

func TestArrayInStruct(t *testing.T) {
	data, err := ioutil.ReadFile("test/array_field_test.json")
	if err != nil {
		t.Fatal(err)
	}
	var js map[string]interface{}
	err = json.Unmarshal(data, &js)
	if (err != nil) {
		t.Fatal(err)
	}

	inputs := make(map[string]interface{})
	i := 123
	inputs["s"] = &StructWithArray{
		Array: []*int{&i},
	}

	exp := &Interpreter{
		Name: "test_array_field",
		Salt: "blasdfalks",
		Inputs: inputs,
		Outputs: make(map[string]interface{}),
		Code: js,
	}

	if _, ok := exp.Run(); !ok {
		t.Fatal("Experiment run failed")
	}

	if elem := *(exp.Outputs["element"].(*int)); elem != 123 {
		t.Fail()
	}
}

type StructWithMap struct {
	Map map[string]int64
}

func TestMapField(t *testing.T) {
	data, err := ioutil.ReadFile("test/map_index_test.json")
	if err != nil {
		t.Fatal(err)
	}
	var js map[string]interface{}
	err = json.Unmarshal(data, &js)
	if (err != nil) {
		t.Fatal(err)
	}

	inputs := make(map[string]interface{})
	mapField := make(map[string]int64)
	mapField["key"] = 42
	inputs["s"] = &StructWithMap{
		Map: mapField,
	}

	exp := &Interpreter{
		Name: "test_map_index",
		Salt: "asdfkhjaslkdfjh",
		Inputs: inputs,
		Outputs: make(map[string]interface{}),
		Code: js,
	}

	if _, ok := exp.Run(); !ok {
		t.Fatal("Experiment run failed")
	}

	if elem := exp.Outputs["element"]; elem != int64(42) {
		t.Fail()
	}

	if exp.Outputs["empty"] != nil {
		t.Fail()
	}
}

type StructWithNilField struct {
	None interface{}
}

func TestStructWithNilField(t *testing.T) {
	data, err := ioutil.ReadFile("test/struct_with_nil_field.json")
	if err != nil {
		t.Fatal(err)
	}
	var js map[string]interface{}
	err = json.Unmarshal(data, &js)
	if (err != nil) {
		t.Fatal(err)
	}

	inputs := make(map[string]interface{})
	inputs["struct"] = &StructWithNilField{}

	exp := &Interpreter{
		Name: "struct with nil field",
		Salt: "safasdf",
		Inputs: inputs,
		Outputs: make(map[string]interface{}),
		Code: js,
	}

	if _, ok := exp.Run(); !ok {
		t.Fatal("Experiment run failed")
	}

	if exp.Outputs["nil"] != nil {
		t.Fail()
	}
}