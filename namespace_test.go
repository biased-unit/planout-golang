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
	js1 := readTest("test/simple_ops.json")
	js2 := readTest("test/random_ops.json")
	js3 := readTest("test/simple.json")

	inputs := make(map[string]interface{})
	inputs["userid"] = "test-id"

	n := NewSimpleNamespace("simple_namespace", 100, "userid", inputs)
	n.AddExperiment("simple ops", js1, 10)
	n.AddExperiment("random ops", js2, 10)
	n.AddExperiment("simple", js3, 80)

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
	n.AddExperiment("random ops", js2, 10)
	y := n.availableSegments

	if reflect.DeepEqual(x, y) == false {
		t.Errorf("Removing and re-adding experiment to a namespace resulted in mismatched allocations. X: %v, Y: %v\n", x, y)
	}
}
