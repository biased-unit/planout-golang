/*
 * Copyright 2015 URX
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package planout

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func runExperiment(rawCode []byte) (*Interpreter, bool) {

	code := make(map[string]interface{})
	json.Unmarshal(rawCode, &code)

	// fmt.Printf("Code: %v\n", code)

	expt := &Interpreter{
		Salt:      "test_salt",
		Evaluated: false,
		Inputs:    map[string]interface{}{},
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
		Code:      code,
	}

	_, ok := expt.Run()

	return expt, ok
}

func runConfig(config []byte) (*Interpreter, bool) {
	setX := []byte(`{"op": "set", "var": "x", "value":`)
	end := []byte(`}`)

	rawCode := append(setX, config...)
	rawCode = append(rawCode, end...)

	return runExperiment(rawCode)
}

func TestCoreOps(t *testing.T) {

	// Test SET
	expt, _ := runExperiment([]byte(`{"op": "set", "value": "x_val", "var": "x"}`))
	x, _ := expt.Get("x")
	if compare(x, "x_val") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	// Test SEQ
	expt, _ = runExperiment([]byte(`
	 	{"op": "seq",
	 	"seq": [ {"op": "set", "value": "x_val", "var": "x"},
	 		 {"op": "set", "value": "y_val", "var": "y"} ]}`))
	x, _ = expt.Get("x")
	if compare(x, "x_val") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}
	y, _ := expt.Get("y")
	if compare(y, "y_val") != 0 {
		t.Errorf("Variable y. Expected y_val. Actual %v\n", y)
	}

	// Test Array
	expt, _ = runExperiment([]byte(`
	 	{"op": "set", "var": "x", "value": {"op": "array", "values": [4, 5, "a"]}}`))
	x, _ = expt.Get("x")

	// Test Dictionary
	expt, _ = runExperiment([]byte(`
	{"op": "set", "var": "x", "value": {"op": "map", "a": 2, "b": "foo", "c": [0, 1, 2]}}`))
	x, _ = expt.Get("x")
	if x == nil {
		t.Errorf("Variable x. Expected a map. Actual nil.\n")
	} else {
		x_map, ok := x.(map[string]interface{})
		if !ok {
			t.Errorf("Variable x. Expected of type map. Actual %v\n", reflect.TypeOf(x))
		}
		foo, ok := x_map["b"]
		if !ok {
			t.Errorf("Variable x['b']. Expected 'foo'. Does not exists.\n")
		}
		if foo != "foo" {
			t.Errorf("Variable x['b']. Expected 'foo'. Actual %v\n", foo)
		}
	}

	// Test empty dictionary
	expt, _ = runExperiment([]byte(`
	{"op": "set", "var": "x", "value": {"op": "map"}}`))
	x, _ = expt.Get("x")
	if x == nil {
		t.Errorf("Variable x. Expected a map. Actual nil.\n")
	} else {
		x_map, ok := x.(map[string]interface{})
		if !ok {
			t.Errorf("Variable x. Expected of type map. Actual %v\n", reflect.TypeOf(x))
		}
		if len(x_map) != 0 {
			t.Errorf("Variable x. Expected empty map. Actual %v\n", x_map)
		}
	}

	// Test Condition
	expt, _ = runExperiment([]byte(`
	 	{"op": "cond",
	 	"cond": [ {"if": 0, "then": {"op": "set", "var": "x", "value": "x_0"}},
	 		  {"if": 1, "then": {"op": "set", "var": "x", "value": "x_1"}}]}`))
	x, _ = expt.Get("x")
	if compare(x, "x_1") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	expt, _ = runExperiment([]byte(`
	 	{"op": "cond",
	 	"cond": [ {"if": 1, "then": {"op": "set", "var": "x", "value": "x_0"}},
	 		  {"if": 0, "then": {"op": "set", "var": "x", "value": "x_1"}}]}`))
	x, _ = expt.Get("x")
	if compare(x, "x_0") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	// Test GET
	expt, _ = runExperiment([]byte(`
	 	{"op": "seq",
	 	"seq": [{"op": "set", "var": "x", "value": "x_val"},
	 		{"op": "set", "var": "y", "value": {"op": "get", "var": "x"}}]}`))
	x, _ = expt.Get("x")
	if compare(x, "x_val") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	y, _ = expt.Get("y")
	if compare(y, "x_val") != 0 {
		t.Errorf("Variable y. Expected x_val. Actual %v\n", y)
	}

	// Test Index
	expt, _ = runConfig([]byte(` {"op": "index", "index": 0, "base": [10, 20, 30]}`))
	x, _ = expt.Get("x")
	if compare(x, 10) != 0 {
		t.Errorf("Variable x. Expected 10. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": 2, "base": [10, 20, 30]}`))
	x, _ = expt.Get("x")
	if compare(x, 30) != 0 {
		t.Errorf("Variable x. Expected 30. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": "a", "base": {"a": 42, "b": 43}}`))
	x, _ = expt.Get("x")
	if compare(x, 42) != 0 {
		t.Errorf("Variable x. Expected 42. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": 6, "base": [10, 20, 30]}`))
	x, _ = expt.Get("x")
	if x != nil {
		t.Errorf("Variable x. Expected nil. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": "c", "base": {"a": 42, "b": 43}}`))
	x, _ = expt.Get("x")
	if x != nil {
		t.Errorf("Variable x. Expected nil. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": 2, "base": {"op": "array", "values": [10, 20, 30]}}`))
	x, _ = expt.Get("x")
	if compare(x, 30) != 0 {
		t.Errorf("Variable x. Expected 30. Actual %v\n", x)
	}

	// Test Coalesce
	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": {"op": "array", "values": [100, 200, 300]}}`))
	x, _ = expt.Get("x")
	lhs := fmt.Sprintf("%v", x)
	rhs := fmt.Sprintf("%v", []interface{}{100, 200, 300})
	if compare(lhs, rhs) != 0 {
		t.Errorf("Variable x. Expected [100, 200, 300]. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [100, 200, 300, null]}`))
	x, _ = expt.Get("x")
	lhs = fmt.Sprintf("%v", x)
	rhs = fmt.Sprintf("%v", []interface{}{100, 200, 300})
	if compare(lhs, rhs) != 0 {
		t.Errorf("Variable x. Expected [100, 200, 300]. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [null]}`))
	x, _ = expt.Get("x")
	if len(x.([]interface{})) != 0 {
		t.Errorf("Variable x. Expected []. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [null, 42, null]}`))
	x, _ = expt.Get("x")
	lhs = fmt.Sprintf("%v", x)
	rhs = fmt.Sprintf("%v", []interface{}{42})
	if compare(lhs, rhs) != 0 {
		t.Errorf("Variable x. Expected [42]. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [null, null, 43]}`))
	x, _ = expt.Get("x")
	lhs = fmt.Sprintf("%v", x)
	rhs = fmt.Sprintf("%v", []interface{}{43})
	if compare(lhs, rhs) != 0 {
		t.Errorf("Variable x. Expected [43]. Actual %v\n", x)
	}

	// Test Length
	expt, _ = runConfig([]byte(`{"op": "length", "values": {"op": "array", "values": [1,2,3,4,5]}}`))
	x, _ = expt.Get("x")
	if compare(x, 5) != 0 {
		t.Errorf("Variable x. Expected 5. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "length", "values": [1,2,3,4,5]}`))
	x, _ = expt.Get("x")
	if compare(x, 5) != 0 {
		t.Errorf("Variable x. Expected 5. Actual %v\n", x)
	}

	// arr = [111, 222, 333];
	// x = length(arr);
	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"arr","value":{"op":"array","values":[111,222,333]}},
						{"op":"set","var":"x","value":{"values":[{"op":"get","var":"arr"}],"op":"length"}}]}`))
	x, _ = expt.Get("x")
	if compare(x, 3) != 0 {
		t.Errorf("Variable x. Expected 3. Actual %v\n", x)
	}

	// a = 111;
	// b = 222;
	// c = [a, b];
	// x = length(c);
	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"a","value":111},
						{"op":"set","var":"b","value":222},
						{"op":"set","var":"c","value":{"op":"array","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]}},
						{"op":"set","var":"x","value":{"values":[{"op":"get","var":"c"}],"op":"length"}}]}`))
	x, _ = expt.Get("x")
	if compare(x, 2) != 0 {
		t.Errorf("Variable x. Expected 2. Actual %v\n", x)
	}

	// a = 111;
	// b = 222;
	// x = length([a, b]);
	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"a","value":111},
						{"op":"set","var":"b","value":222},
						{"op":"set","var":"x", "value":{"op":"length","values":[{"op":"array","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]}]}}]}`))
	x, _ = expt.Get("x")
	if compare(x, 2) != 0 {
		t.Errorf("Variable x. Expected 2. Actual %v\n", x)
	}

	// a = 1111;
	// x = length([a, 3333]);
	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"a","value":1111},
						{"op":"set","var":"x","value":{"values":[{"op":"array","values":[{"op":"get","var":"a"},3333]}],"op":"length"}}]}`))
	x, _ = expt.Get("x")

	// x = length([111, 222]);
	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"x","value":{"values":[{"op":"array","values":[111,222]}],"op":"length"}}]}`))
	x, _ = expt.Get("x")
	if compare(x, 2) != 0 {
		t.Errorf("Variable x. Expected 2. Actual %v\n", x)
	}

	// x = length([]);
	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"x","value":{"values":[{"op":"array","values":[]}],"op":"length"}}]}`))
	x, _ = expt.Get("x")
	if compare(x, 0) != 0 {
		t.Errorf("Variable x. Expected 2. Actual %v\n", x)
	}

	// Test NOT operator
	expt, _ = runConfig([]byte(`{"op": "not", "value": 0}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected True. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "not", "value": false}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected True. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "not", "value": 1}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "not", "value": true}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	// Test OR operator
	expt, _ = runConfig([]byte(`{"op": "or", "values": [0, 0, 0, 0]}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "or", "values": [0, 0, 0, 1]}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "or", "values": [false, true, false]}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	// Test AND operator
	expt, _ = runConfig([]byte(`{"op": "and", "values": [1, 1, 0]}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "and", "values": [0, 0, 1]}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "and", "values": [true, true, true]}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected True. Actual %v\n", x)
	}

	// Test Commutative operators
	expt, _ = runConfig([]byte(`{"op": "min", "values": [33, 7, 18, 21, -3]}`))
	x, _ = expt.Get("x")
	if compare(x, -3) != 0 {
		t.Errorf("Variable x. Expected -3. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "max", "values": [33, 7, 18, 21, -3]}`))
	x, _ = expt.Get("x")
	if compare(x, 33) != 0 {
		t.Errorf("Variable x. Expected 33. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "sum", "values": [33, 7, 18, 21, -3]}`))
	x, _ = expt.Get("x")
	if compare(x, 76) != 0 {
		t.Errorf("Variable x. Expected 76. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "product", "values": [33, 7, 18, 21, -3]}`))
	x, _ = expt.Get("x")
	if compare(x, -261954) != 0 {
		t.Errorf("Variable x. Expected -261954. Actual %v\n", x)
	}

	// Test Binary operators
	expt, _ = runConfig([]byte(`{"op": "equals", "left": 1, "right": 2}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "equals", "left": 2, "right": 2}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected True. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": ">", "left": 1, "right": 2}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "<", "left": 1, "right": 2}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected True. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": ">=", "left": 2, "right": 2}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected True. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": ">=", "left": 1, "right": 2}`))
	x, _ = expt.Get("x")
	if compare(x, false) != 0 {
		t.Errorf("Variable x. Expected False. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "<=", "left": 1, "right": 2}`))
	x, _ = expt.Get("x")
	if compare(x, true) != 0 {
		t.Errorf("Variable x. Expected True. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "%", "left": 11, "right": 3}`))
	x, _ = expt.Get("x")
	if compare(x, 2) != 0 {
		t.Errorf("Variable x. Expected 2. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(`{"op": "/", "left": 3, "right": 4}`))
	x, _ = expt.Get("x")
	if compare(x, 0.75) != 0 {
		t.Errorf("Variable x. Expected 0.75 . Actual %v\n", x)
	}

	// Test RETURN operator
	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"x","value":2}, 
					{"op":"return","value":true}, 
					{"op":"set","var":"y","value":4}]}`))
	if !isTrue(expt.InExperiment) {
		t.Errorf("Variable x. Expected True . Actual %v\n", expt.InExperiment)
	}

	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"x","value":2}, 
					{"op":"return","value":42}, 
					{"op":"set","var":"y","value":4}]}`))
	if !isTrue(expt.InExperiment) {
		t.Errorf("Variable x. Expected True . Actual %v\n", expt.InExperiment)
	}

	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"x","value":2}, 
					{"op":"return","value":false}, 
					{"op":"set","var":"y","value":4}]}`))
	if isTrue(expt.InExperiment) {
		t.Errorf("Variable x. Expected False . Actual %v\n", expt.InExperiment)
	}

	expt, _ = runExperiment([]byte(`{"op":"seq",
					"seq":[{"op":"set","var":"x","value":2}, 
					{"op":"return","value":0}, 
					{"op":"set","var":"y","value":4}]}`))
	if isTrue(expt.InExperiment) {
		t.Errorf("Variable x. Expected False . Actual %v\n", expt.InExperiment)
	}
}
