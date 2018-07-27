/*
 * Copyright 2017 biased-unit
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
	"io/ioutil"
	"testing"
)

// Example input structure
type ExampleStruct struct {
	Member int
	String string
}

func TestExample(t *testing.T) {
	// Read PlanOut code from file on disk.
	data, _ := ioutil.ReadFile("test/simple_ops.json")

	// The PlanOut code is expected to use json.
	// This format is the same as the output of
	// the PlanOut compiler webapp
	// http://facebook.github.io/planout/demo/planout-compiler.html
	var js map[string]interface{}
	json.Unmarshal(data, &js)

	// Set the necessary input parameters required to run
	// the experiments. For instance, simple_ops.json expects
	// the value for 'userid' to be set.
	example := ExampleStruct{Member: 101, String: "test-string"}
	params := make(map[string]interface{})
	params["experiment_salt"] = "expt"
	params["userid"] = generateString()
	params["struct"] = example

	// Construct an instance of the Interpreter object.
	// Initialize Salt and set Inputs to params.
	expt := &Interpreter{
		Salt:      "global_salt",
		Evaluated: false,
		Inputs:    params,
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
		Code:      js,
	}

	// Call the Run() method on the Interpreter instance.
	// The output of the run will contain the dictionary
	// of variables and associated values that were evaluated
	outputs, ok := expt.Run()
	if !ok {
		t.Errorf("Error running the interpreter for 'test/simple_ops.json'\n")
	}

	if outputs["z2"] != "test-string" {
		t.Errorf("Outputs wrong")
	}

}
