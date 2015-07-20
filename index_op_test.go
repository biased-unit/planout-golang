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