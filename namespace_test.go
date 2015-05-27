/*
 * Copyright 2014 URX
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
	"reflect"
	"testing"
)

func TestSimpleNamespace(t *testing.T) {

	constructInterpreter := func(filename string, name, salt string, inputs map[string]interface{}) *Interpreter {
		js := readTest(filename)
		return &Interpreter{Name: name,
			Salt:      salt,
			Inputs:    inputs,
			Overrides: make(map[string]interface{}),
			Outputs:   make(map[string]interface{}),
			Code:      js,
		}
	}

	inputs := make(map[string]interface{})
	inputs["userid"] = "test-id"

	e1 := constructInterpreter("test/simple_ops.json", "simple_ops", "simple_ops_salt", inputs)
	e2 := constructInterpreter("test/random_ops.json", "simple_ops", "simple_ops_salt", inputs)
	e3 := constructInterpreter("test/simple.json", "simple_ops", "simple_ops_salt", inputs)

	n := NewSimpleNamespace("simple_namespace", 100, "userid", inputs)
	n.AddExperiment("simple ops", e1, 10)
	n.AddExperiment("random ops", e2, 10)
	n.AddExperiment("simple", e3, 80)

	x := n.availableSegments

	seg := n.getSegment()
	if seg != 92 {
		t.Errorf("Incorrect allocation (%v) for test-id. Expected 92.", seg)
	}

	interpreter := n.Run()
	output, exists := interpreter.Get("output")

	if !exists || output != "test" {
		t.Errorf("Namespace run was not successful out:[%+v]\n", interpreter)
	}

	n.RemoveExperiment("random ops")
	n.AddExperiment("random ops", e2, 10)
	y := n.availableSegments

	if reflect.DeepEqual(x, y) == false {
		t.Errorf("Removing and re-adding experiment to a namespace resulted in mismatched allocations. X: %v, Y: %v\n", x, y)
	}

	unitstr := generateString()
	inputs["userid"] = unitstr
	n = NewSimpleNamespace("simple_namespace", 100, "userid", inputs)
	n.AddExperiment("simple ops", e1, 10)
	n.AddExperiment("random ops", e2, 10)
	n.AddExperiment("simple", e3, 80)
	n.RemoveExperiment("random ops")
	n.RemoveExperiment("simple ops")
	n.RemoveExperiment("simple")
	if len(n.availableSegments) != 100 {
		t.Errorf("Expected all segments to be available. Actual %d\n", len(n.availableSegments))
	}
}
