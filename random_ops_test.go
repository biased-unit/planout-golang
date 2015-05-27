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
	"math"
	"math/rand"
	"reflect"
	"testing"
	"text/template"
	"time"
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
	x, _ := expt.Get("x")

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
		"seq":[{"op":"set","var":"y","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	y, _ := expt.Get("y")

	if reflect.DeepEqual(x, y) {
		t.Errorf("Variable 'x' and 'y'. Expected inequality. Actual x=%v, y=%v\n", x, y)
	}

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"z","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","salt":"x"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	z, _ := expt.Get("z")

	if reflect.DeepEqual(x, z) == false {
		t.Errorf("Variable z used 'x' as parameter salt. Expected equality. Actual x=%v, z=%v\n", x, z)
	}

	salt = "assign_salt_b"

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"x","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	x, _ = expt.Get("x")

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"y","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	y, _ = expt.Get("y")

	if reflect.DeepEqual(x, y) == false {
		t.Errorf("Variable 'x' and 'y'. Expected equality. Actual x=%v, y=%v\n", x, y)
	}

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"x","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs2"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	x, _ = expt.Get("x")

	expt, _ = runExperimentWithSalt([]byte(`{"op":"seq",
	"seq":[{"op":"set","var":"y","value":{"min":0,"max":100000,"unit":{"op":"get","var":"userid"},"op":"randomInteger","full_salt":"fs2"}}]}`),
		salt, map[string]interface{}{"userid": unit})
	y, _ = expt.Get("y")

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

func (h Histogram) density() map[string]float64 {
	sum := 0
	for _, v := range h.hist {
		sum = sum + v
	}

	dense := make(map[string]float64)
	for k, v := range h.hist {
		dense[k] = float64(v) / float64(sum)
	}
	return dense
}

func (h Histogram) passed(expected map[string]float64) bool {
	dense := h.density()
	for value, density := range dense {
		expected_density, exists := expected[value]
		if !exists {
			return false
		} else {
			if math.Abs(density-expected_density) > 0.05 {
				return false
			}
		}
	}
	return true
}

func randomExperiment(t *testing.T, textTemplate string, data interface{}, runs int) Histogram {

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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < runs; i++ {
		expt, _ := runExperimentWithInputs(code.Bytes(), map[string]interface{}{"i": r.Intn(math.MaxUint32)})
		x[i], _ = expt.Get("x")
		h.add(x[i])
	}
	code.Reset()
	return h
}

func TestBernoulliTrial(t *testing.T) {
	var textTemplate string = `{"op":"seq",
	"seq":[{"op":"set","var":"x","value":{"p":{{.}},"unit":{"op":"get","var":"i"},"op":"bernoulliTrial"}}]}`

	probabilities := []float64{0.5, 0.1, 0.9, 0.75}
	for i := range probabilities {
		h := randomExperiment(t, textTemplate, probabilities[i], 1000)
		expected_density := map[string]float64{"1": probabilities[i], "0": 1.0 - probabilities[i]}
		if h.passed(expected_density) == false {
			t.Errorf("Bernoulli Trial. Expected %v. Actual %v\n", expected_density, h)
		}
	}
}

func TestUniformChoice(t *testing.T) {
	var textTemplate string = `{"op":"seq",
	"seq":[{"op":"set","var":"x",
	"value":{"choices":{"op":"array","values":{{.}}},
	"unit":{"op":"get","var":"i"},"op":"uniformChoice"}}]}`

	inputs := []struct {
		Choices  string
		Expected map[string]float64
	}{{`["a"]`, map[string]float64{"a": 1.0}},
		{`["a", "b"]`, map[string]float64{"a": 0.5, "b": 0.5}},
		{`[1, 2, 3, 4]`, map[string]float64{"1": 0.25, "2": 0.25, "3": 0.25, "4": 0.25}},
	}
	for i := range inputs {
		h := randomExperiment(t, textTemplate, inputs[i].Choices, 1000)
		if h.passed(inputs[i].Expected) == false {
			t.Errorf("Uniform choice. Expected %v. Actual %v\n", inputs[i].Expected, h)
		}
	}
}

func TestWeightedChoice(t *testing.T) {
	var textTemplate string = `{"op":"seq",
	"seq":[{"op":"set","var":"x",
	"value":{"choices":{"op":"array","values":{{.Choices}}},
	"weights":{"op":"array","values":{{.Weights}}},
	"unit":{"op":"get","var":"i"},"op":"weightedChoice"}}]}`

	inputs := []struct {
		Choices, Weights string
		Expected         map[string]float64
	}{{`["a", "b", "c"]`, `[0.8, 0.1, 0.1]`, map[string]float64{"a": 0.8, "b": 0.1, "c": 0.1}},
		{`["a", "b"]`, `[0.3333, 0.6667]`, map[string]float64{"a": 0.3333, "b": 0.6667}},
		{`["a", "b", "c"]`, `[0, 1, 0]`, map[string]float64{"a": 0.0, "b": 1.0, "c": 0.0}},
		{`["a", "b", "c", "a"]`, `[0.2, 0.4, 0, 0.4]`, map[string]float64{"a": 0.6, "b": 0.4}},
		{`["a", "b", "c"]`, `[0.6, 0.4, 0]`, map[string]float64{"a": 0.6, "b": 0.4}},
	}
	for i := range inputs {
		h := randomExperiment(t, textTemplate, inputs[i], 1000)
		if h.passed(inputs[i].Expected) == false {
			t.Errorf("Weighted choice. Expected %v. Actual %v\n", inputs[i].Expected, h)
		}
	}
}

func TestSampling(t *testing.T) {
	var textTemplate string = `{"op":"seq",
	"seq":[{"op":"set","var":"x",
	"value":{"choices":{"op":"array","values":{{.Choices}}},"draws":{{.Draws}}, {{.Unit}} "op":"sample"}}]}`

	unit := `"unit":{"op":"get","var":"i"},`

	inputs := []struct {
		Choices, Draws, Unit string
		Expected             map[string]float64
	}{{`[1, 2, 3]`, `3`, unit, map[string]float64{`[3 1 2]`: 0.167, `[2 1 3]`: 0.167, `[1 3 2]`: 0.167, `[3 2 1]`: 0.167, `[1 2 3]`: 0.167, `[2 3 1]`: 0.167}},
		{`[1, 2, 3]`, `2`, unit, map[string]float64{`[3 1]`: 0.167, `[2 1]`: 0.167, `[1 3]`: 0.167, `[3 2]`: 0.167, `[1 2]`: 0.167, `[2 3]`: 0.167}},
		{`["a", "a", "b"]`, `3`, unit, map[string]float64{`[b a a]`: 0.334, `[a a b]`: 0.334, `[a b a]`: 0.334}},
	}
	for i := range inputs {
		h := randomExperiment(t, textTemplate, inputs[i], 1000)
		if h.passed(inputs[i].Expected) == false {
			t.Errorf("Weighted choice. Expected %v. Actual %v\n", inputs[i].Expected, h)
		}
	}
}
