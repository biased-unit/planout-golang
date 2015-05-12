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
	"testing"
)

func TestSimpleAssignment(t *testing.T) {
	js := readTest("test/assignment_test.json")

	var runs int = 100

	var tester_unit_value int = 4

	for i := 0; i < runs; i++ {
		params := make(map[string]interface{})
		params["tester_unit"] = tester_unit_value

		expt := &Interpreter{
			Salt:      "test_salt",
			Evaluated: false,
			Inputs:    params,
			Outputs:   map[string]interface{}{},
			Overrides: map[string]interface{}{},
			Code:      js,
		}

		output, ok := expt.Run()
		if !ok {
			t.Errorf("Error running experiment 'test/assignment_test.json'\n")
			return
		}

		foo := output["foo"]
		if compare(foo, "b") != 0 {
			t.Errorf("Variable 'foo'. Expected 'b'. Actual %v\n", foo)
		}

		bar := output["bar"]
		if compare(bar, "a") != 0 {
			t.Errorf("Variable 'bar'. Expected 'a'. Actual %v\n", bar)
		}

		baz := output["baz"]
		if compare(baz, "a") != 0 {
			t.Errorf("Variable 'baz'. Expected 'a'. Actual %v\n", baz)
		}
	}
}

func TestSimpleOverride(t *testing.T) {
	js := readTest("test/override_test.json")

	var runs int = 100

	var tester_unit_value int = 4

	for i := 0; i < runs; i++ {
		params := make(map[string]interface{})
		params["tester_unit"] = tester_unit_value

		overrides := make(map[string]interface{})
		overrides["x"] = 42
		overrides["y"] = 43

		expt := &Interpreter{
			Salt:      "test_salt",
			Evaluated: false,
			Inputs:    params,
			Outputs:   map[string]interface{}{},
			Overrides: overrides,
			Code:      js,
		}

		_, ok := expt.Run()
		if !ok {
			t.Errorf("Error running experiment 'test/assignment_test.json'\n")
			return
		}

		x, _ := expt.Get("x")
		if compare(x, 42) != 0 {
			t.Errorf("Variable 'x'. Expected override 42. Actual %v\n", x)
		}

		y, _ := expt.Get("y")
		if compare(y, 43) != 0 {
			t.Errorf("Variable 'y'. Expected override 43. Actual %v\n", y)
		}
	}
}
