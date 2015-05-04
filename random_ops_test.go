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
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"text/template"
)

func TestRandomOps(t *testing.T) {
	js := readTest("test/random_ops.json")

	var runs int = 5

	a := make([]float64, 0, runs)
	b := make([]float64, 0, runs)
	c := make([]float64, 0, runs)
	userids := make([]string, 0, runs)

	for i := 0; i < runs; i++ {
		params := make(map[string]interface{})
		userid := generateString()
		params["userid"] = userid
		userids = append(userids, userid)

		expt := &Interpreter{
			Salt:      "global_salt",
			Evaluated: false,
			Inputs:    params,
			Outputs:   map[string]interface{}{},
			Overrides: map[string]interface{}{},
			Code:      js,
		}

		output, ok := expt.Run()
		if !ok {
			t.Errorf("Error running experiment 'test/random_ops.json'\n")
			return
		}

		numbers := output["numbers"].([]interface{})

		a = append(a, output["a"].(float64))
		b = append(b, output["b"].(float64))
		c = append(c, output["c"].(float64))

		e := output["e"].(float64)
		f := output["f"].(float64)
		if e == 3 || e == 4 {
			t.Errorf("WeightedChoice(%v) Variable 'e' must only be 1 or 2\n", userid)
		}
		if f == 1 || f == 2 {
			t.Errorf("WeightedChoice(%v) Variable 'f' must only be 3 or 4\n", userid)
		}

		g := output["g"]
		h := output["h"]
		if compare(g, 0) != 0 {
			t.Errorf("BernoulliTrial(%v) Variable 'g' (%v) must be 0\n", userid, g)
		}
		if compare(h, 1) != 0 {
			t.Errorf("BernoulliTrial(%v) Variable 'h' (%v) must be 1\n", userid, h)
		}

		i := output["i"].([]interface{})
		j := output["j"].([]interface{})
		if len(i) != len(numbers) {
			t.Errorf("Sample(%v) Expected length of 'i' (%v) to be the same as the input\n", userid, i)
		}
		if len(j) != 3 {
			t.Errorf("Sample(%v) Expected length of 'j' (%v) to be the same as the draw count (3)\n", userid, j)
		}
	}

	if reflect.DeepEqual(a, b) != false {
		t.Errorf("UniformChoice(%v): Expected 'a' (%v) and 'b' (%v) to be different", userids, a, b)
	}

	if reflect.DeepEqual(a, c) != true {
		t.Errorf("UniformChoice(%v): Expected 'a' (%v) and 'c' (%v) to be equal (same parameter salt)", userids, a, c)
	}
}

func runExperimentWithSalt(rawCode []byte, salt string, inputs map[string]interface{}) (*Interpreter, bool) {

	code := make(map[string]interface{})
	json.Unmarshal(rawCode, &code)

	expt := &Interpreter{
		Salt:      salt,
		Evaluated: false,
		Inputs:    inputs,
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
		Code:      code,
	}

	_, ok := expt.Run()

	return expt, ok
}

func TestSalts(t *testing.T) {
	unit := generateString()
	salt := "assign_salt_a"

	expt, _ := runExperimentWithSalt([]byte(`{"op":"seq",
		"seq":[{"op":"set","var":"x","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	x, _ := expt.get("x")

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
		"seq":[{"op":"set","var":"y","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	y, _ := expt.get("y")

	if reflect.DeepEqual(x, y) {
		t.Errorf("Variable 'x' and 'y'. Expected inequality. Actual x=%v, y=%v\n", x, y)
	}

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"z","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","salt":"x"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	z, _ := expt.get("z")

	if reflect.DeepEqual(x, z) == false {
		t.Errorf("Variable z used 'x' as parameter salt. Expected equality. Actual x=%v, z=%v\n", x, z)
	}

	salt = "assign_salt_b"

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"x","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	x, _ = expt.get("x")

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"y","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	y, _ = expt.get("y")

	if reflect.DeepEqual(x, y) == false {
		t.Errorf("Variable 'x' and 'y'. Expected equality. Actual x=%v, y=%v\n", x, y)
	}

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"x","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs2"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	x, _ = expt.get("x")

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"y","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs2"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	y, _ = expt.get("y")

	if reflect.DeepEqual(x, y) == false {
		t.Errorf("Variable 'x' and 'y'. Expected equality. Actual x=%v, y=%v\n", x, y)
	}
}

func runExperimentWithInputs(rawCode []byte, inputs map[string]interface{}) (*Interpreter, bool) {

	code := make(map[string]interface{})
	json.Unmarshal(rawCode, &code)

	expt := &Interpreter{
		Salt:      "experiment_salt",
		Evaluated: false,
		Inputs:    inputs,
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
		Code:      code,
	}

	_, ok := expt.Run()

	return expt, ok
}

type Histogram struct {
	hist map[string]int
}

func (h Histogram) add(element interface{}) {
	var key string = fmt.Sprintf("%v", element)
	count, exists := h.hist[key]
	if exists {
		delete(h.hist, key)
		h.hist[key] = count + 1
	} else {
		h.hist[key] = 1
	}
}

func randomExperiment(t *testing.T, textTemplate string, data interface{}, runs int) {

	parsedTemplate, err := template.New("test").Parse(textTemplate)
	if err != nil {
		t.Errorf("Error parsing experiment template: %v\n", parsedTemplate)
	}

	var code bytes.Buffer
	err = parsedTemplate.Execute(&code, data)
	if err != nil {
		t.Errorf("Error executing experiment template: %v\n", parsedTemplate)
	}

	x := make([]interface{}, runs)
	h := Histogram{hist: map[string]int{}}
	for i := 0; i < runs; i++ {
		expt, _ := runExperimentWithInputs(code.Bytes(), map[string]interface{}{"i": i})
		x[i], _ = expt.get("x")
		h.add(x[i])
	}
	fmt.Printf("Histogram: %v\n", h)
	code.Reset()
}

func TestBernoulliTrial(t *testing.T) {
	fmt.Println("Testing Bernoulli Trial ...")

	var textTemplate string = `{"op":"seq",
	"seq":[{"op":"set","var":"x","value":{"p":{{.}},"unit":{"op":"get","var":"i"},"op":"bernoulliTrial"}}]}`

	probabilities := []float64{0.5, 0.1, 0.9, 0.75}
	for i := range probabilities {
		randomExperiment(t, textTemplate, probabilities[i], 1000)
	}
}

func TestUniformChoice(t *testing.T) {
	fmt.Println("Testing Uniform Choice ...")

	var textTemplate string = `{"op":"seq",
	"seq":[{"op":"set","var":"x",
	"value":{"choices":{"op":"array","values":{{.}}},
	"unit":{"op":"get","var":"i"},"op":"uniformChoice"}}]}`

	choices := []interface{}{`["a"]`, `["a", "b"]`, `[1, 2, 3, 4]`}
	for i := range choices {
		randomExperiment(t, textTemplate, choices[i], 1000)
	}
}

type WeightedChoice struct {
	Choices, Weights string
}

func TestWeightedChoice(t *testing.T) {
	fmt.Println("Testing Uniform Choice ...")

	var textTemplate string = `{"op":"seq",
	"seq":[{"op":"set","var":"x",
	"value":{"choices":{"op":"array","values":{{.Choices}}},
	"weights":{"op":"array","values":{{.Weights}}},
	"unit":{"op":"get","var":"i"},"op":"weightedChoice"}}]}`

	choices := []interface{}{WeightedChoice{Choices: `["a", "b", "c"]`, Weights: `[0.8, 0.1, 0.1]`}}
	for i := range choices {
		randomExperiment(t, textTemplate, choices[i], 1000)
	}
}
