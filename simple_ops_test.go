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
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func readTest(f string) map[string]interface{} {
	data, _ := ioutil.ReadFile(f)
	var js map[string]interface{}
	json.Unmarshal(data, &js)
	return js
}

func TestSimpleOps(t *testing.T) {
	js := readTest("test/simple_ops.json")
	params := make(map[string]interface{})

	ok := Experiment(js, params)
	if !ok {
		t.Errorf("Error running experiment 'test/simple_ops.json'\n")
		return
	}

	a := params["a"]
	if compare(a, 10.0) != 0 {
		t.Errorf("Variable 'a'. Expected 10. Actual %v\n", a)
	}

	b := params["b"]
	if compare(b, 3.14) != 0 {
		t.Errorf("Variable 'b'. Expected 3.14. Actual %v\n", b)
	}

	c := params["c"]
	if compare(c, "hello") != 0 {
		t.Errorf("Variable 'c'. Expected 'hello'. Actual %v\n", c)
	}

	d := params["d"].([]interface{})
	if len(d) != 3 {
		t.Errorf("Expected length of variable 'd' = 3. Actual len %v\n", len(d))
	}

	if reflect.DeepEqual(d[0].(float64), 1.0) == false ||
		reflect.DeepEqual(d[1].(float64), 2.0) == false ||
		reflect.DeepEqual(d[2].(float64), 3.0) == false {
		t.Errorf("Variable 'd'. Expected [1 2 3]. Actual %v\n", d)
	}

	e := params["e"].([]interface{})
	if len(e) != 3 {
		t.Errorf("Expected length of variable 'e' = 3. Actual len %v\n", len(e))
	}

	if reflect.DeepEqual(e[0].(float64), 10.0) == false ||
		reflect.DeepEqual(e[1].(float64), 3.14) == false {
		t.Errorf("Variable 'e'. Expected [10 3.14 [1 2 3]] Actual %v\n", e)
	}

	if reflect.DeepEqual(e[2], d) == false {
		t.Errorf("Variable 'e'. Expected last element to be [1 2 3]] Actual %v\n", e[2])
	}

	f := params["f"]
	if compare(f, 3.0) != 0 {
		t.Errorf("Variable 'f'. Expected 3. Actual %v\n", f)
	}

	g := params["g"]
	if compare(g, 1.0) != 0 {
		t.Errorf("Variable 'g'. Expected 1.0. Actual %v\n", g)
	}

	h := params["h"]
	if compare(h, 1.0) != 0 {
		t.Errorf("Variable 'h'. Expected 1.0. Actual %v\n", h)
	}

	i := params["i"]
	if compare(i, 1.0) != 0 {
		t.Errorf("Variable 'i'. Expected 1.0. Actual %v\n", i)
	}

	j := params["j"]
	if compare(j, 13.14) != 0 {
		t.Errorf("Variable 'j'. Expected 1.0. Actual %v\n", j)
	}

	k := params["k"]
	if compare(k, 31.4) != 0 {
		t.Errorf("Variable 'k'. Expected 31.4. Actual %v\n", k)
	}

	l := params["l"]
	if compare(l, 3.1847) != 0 {
		t.Errorf("Variable 'k'. Expected 3.1847. Actual %v\n", l)
	}

	m := params["m"]
	if compare(m, 1.0) != 0 {
		t.Errorf("Variable 'm'. Expected 1. Actual %v\n", m)
	}

	n := params["n"]
	if compare(n, 10.0) != 0 {
		t.Errorf("Variable 'n'. Expected 10.0. Actual %v\n", n)
	}

	o := params["o"]
	if compare(o, 3.14) != 0 {
		t.Errorf("Variable 'o'. Expected 3.14. Actual %v\n", o)
	}

	s := params["s"]
	if s.(bool) == true {
		t.Errorf("Variable 's'. Expected false.\n")
	}

	tval := params["t"]
	if tval.(bool) == false {
		t.Errorf("Variable 't'. Expected true.\n")
	}

	u := params["u"]
	if u.(bool) == true {
		t.Errorf("Variable 'u'. Expected false (NOT p).\n")
	}

	w := params["w"]
	if reflect.DeepEqual(w, d) == false {
		t.Errorf("Variable 'w' %v. Expected [1, 2, 3] (same as variable 'd')\n", w)
	}

	x := params["x"]
	if compare(x, 2) != 0 {
		t.Errorf("Variable 'x' %v. Expected 2\n", x)
	}
}
