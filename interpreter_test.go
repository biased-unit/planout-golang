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

func TestInterpreter(t *testing.T) {
	js := readTest("test/interpreter_test.json")

	var runs int = 100

	var userid int = 123454

	for i := 0; i < runs; i++ {
		params := make(map[string]interface{})
		params["userid"] = userid

		expt := &Interpreter{
			Salt:      "foo",
			Evaluated: false,
			Inputs:    params,
			Outputs:   map[string]interface{}{},
			Overrides: map[string]interface{}{},
			Code:      js,
		}

		_, ok := expt.Run()
		if !ok {
			t.Errorf("Error running experiment 'test/interpreter_test.json'\n")
			return
		}

		x, _ := expt.get("specific_goal")
		if compare(x, 1) != 0 {
			t.Errorf("Variable 'x'. Expected override 1. Actual %v\n", x)
		}

		y, _ := expt.get("ratings_goal")
		if compare(y, 320) != 0 {
			t.Errorf("Variable 'y'. Expected override 320. Actual %v\n", y)
		}
	}

}

func TestInterpreterWithOverride(t *testing.T) {
	js := readTest("test/interpreter_test.json")

	var runs int = 100

	var userid int = 123454

	for i := 0; i < runs; i++ {

		params := make(map[string]interface{})
		params["userid"] = userid

		overrides := make(map[string]interface{})
		overrides["specific_goal"] = 0

		expt := &Interpreter{
			Salt:      "foo",
			Evaluated: false,
			Inputs:    params,
			Outputs:   map[string]interface{}{},
			Overrides: overrides,
			Code:      js,
		}

		_, ok := expt.Run()
		if !ok {
			t.Errorf("Error running experiment 'test/interpreter_test.json'\n")
			return
		}

		x, _ := expt.get("specific_goal")
		if compare(x, 0) != 0 {
			t.Errorf("Variable 'x'. Expected override 1. Actual %v\n", x)
		}

		y, _ := expt.get("ratings_goal")
		if y != nil {
			t.Errorf("Variable 'y'. Expected nil. Actual %v\n", y)
		}

	}

	// For this next test, reset the userid unit set in the experiment
	// to 123453 but include an override for userid to 123454 (same
	// value as previous experiment). Ensure that the override kicks
	// in and the outcome is the same as before.

	userid = 123453

	for i := 0; i < runs; i++ {

		params := make(map[string]interface{})
		params["userid"] = userid

		overrides := make(map[string]interface{})
		overrides["userid"] = 123454

		expt := &Interpreter{
			Salt:      "foo",
			Evaluated: false,
			Inputs:    params,
			Outputs:   map[string]interface{}{},
			Overrides: overrides,
			Code:      js,
		}

		_, ok := expt.Run()
		if !ok {
			t.Errorf("Error running experiment 'test/interpreter_test.json'\n")
			return
		}

		x, _ := expt.get("specific_goal")
		if compare(x, 1) != 0 {
			t.Errorf("Variable 'x'. Expected override 1. Actual %v\n", x)
		}
	}
}
