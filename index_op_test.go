package planout

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func getInterpreter(filename string) (*Interpreter, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var js map[string]interface{}
	err = json.Unmarshal(data, &js)
	if err != nil {
		return nil, err
	}

	return &Interpreter{
		Name:      "the name",
		Salt:      "the salt",
		Evaluated: false,
		Inputs:    make(map[string]interface{}),
		Outputs:   make(map[string]interface{}),
		Code:      js,
	}, nil
}

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
	exp, err := getInterpreter("test/nested_index.json")
	if err != nil {
		t.Fatal(err)
	}

	exp.Inputs["s"] = &NestedStruct{
		Outer: &Outer{
			Inner: &Inner{
				Value: "foo",
			},
		},
	}

	if _, ok := exp.Run(); !ok {
		t.Fatal("Failed to run experiment")
	}

	if exp.Outputs["out"] != "foo" {
		t.Fail()
	}
}

type StructWithArray struct {
	Array []*int
}

func TestArrayInStruct(t *testing.T) {
	exp, err := getInterpreter("test/array_field_test.json")
	if err != nil {
		t.Fatal(err)
	}

	i := int(123)
	exp.Inputs["s"] = &StructWithArray{
		Array: []*int{&i},
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
	exp, err := getInterpreter("test/map_index_test.json")
	if err != nil {
		t.Fatal(err)
	}

	mapField := make(map[string]int64)
	mapField["key"] = 42
	exp.Inputs["s"] = &StructWithMap{
		Map: mapField,
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
	exp, err := getInterpreter("test/struct_with_nil_field.json")
	if err != nil {
		t.Fatal(err)
	}

	exp.Inputs["struct"] = &StructWithNilField{}

	if _, ok := exp.Run(); !ok {
		t.Fatal("Experiment run failed")
	}

	if exp.Outputs["nil"] != nil {
		t.Fail()
	}
}
