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

package goplanout

import (
	"reflect"
	"testing"
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
