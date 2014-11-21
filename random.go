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
	"crypto/sha1"
	"fmt"
	"strconv"
)

func hash(in string) uint64 {

	// Compute 20- byte sha1
	var x [20]byte = sha1.Sum([]byte(in))

	// Get the first 15 characters of the hexdigest.
	var y string = fmt.Sprintf("%x", x[0:8])
	y = y[0 : len(y)-1]

	// Convert hex string into uint64
	var z uint64 = 0
	z, _ = strconv.ParseUint(y, 16, 64)

	return z
}

func generateExperimentId(units interface{}, params map[string]interface{}) string {
	unitstr := generateUnitStr(units)
	var salt string = ""
	full_salt, exists := params["full_salt"]
	if exists {
		salt = full_salt.(string)
	} else {
		expt_salt, exists := params["experiment_salt"]
		if exists {
			salt = expt_salt.(string)
			salt = salt + "." + params["salt"].(string)
		} else {
			salt = params["salt"].(string)
		}
	}
	experimentid := salt
	if unitstr != "" {
		experimentid = experimentid + "." + unitstr
	}
	return experimentid
}

func getHash(m, params map[string]interface{}, appended_units ...string) uint64 {
	units := evaluate(m["unit"], params)

	_, exists := m["salt"]
	if exists {
		parameter_salt := evaluate(m["salt"], params)
		params["salt"] = parameter_salt.(string)
	}

	experimentid := generateExperimentId(units, params)

	if len(appended_units) > 0 {
		for i := range appended_units {
			experimentid = experimentid + "." + appended_units[i]
		}
	}

	return hash(experimentid)
}

func getUniform(m, params map[string]interface{}, min, max float64, appended_units ...string) float64 {
	scale, _ := strconv.ParseUint("FFFFFFFFFFFFFFF", 16, 64)
	append_string := ""
	var h uint64 = 0
	if len(appended_units) == 0 {
		h = getHash(m, params)
	} else {
		append_string = append_string + appended_units[0]
		for i := range appended_units {
			if i > 0 {
				append_string = append_string + "." + appended_units[i]
			}
		}
		h = getHash(m, params, append_string)
	}
	shift := float64(h) / float64(scale)
	return min + shift*(max-min)
}

type uniformChoice struct{ params map[string]interface{} }

func (s *uniformChoice) execute(m map[string]interface{}) interface{} {
	existOrPanic(m, []string{"choices", "unit"}, "UniformChoice")
	choices := evaluate(m["choices"], s.params).([]interface{})
	nchoices := uint64(len(choices))
	idx := getHash(m, s.params) % nchoices
	choice := choices[idx]
	return choice
}

type bernoulliTrial struct{ params map[string]interface{} }

func (s *bernoulliTrial) execute(m map[string]interface{}) interface{} {
	existOrPanic(m, []string{"unit"}, "BernoulliTrial")
	pvalue := evaluate(m["p"], s.params).(float64)
	rand_val := getUniform(m, s.params, 0.0, 1.0)
	if rand_val <= pvalue {
		return 1
	}
	return 0
}

type bernoulliFilter struct{ params map[string]interface{} }

func (s *bernoulliFilter) execute(m map[string]interface{}) interface{} {
	existOrPanic(m, []string{"choices", "unit"}, "BernoulliFilter")
	pvalue := evaluate(m["p"], s.params).(float64)
	choices := evaluate(m["choices"], s.params).([]interface{})
	ret := make([]interface{}, 0, len(choices))
	for i := range choices {
		append_str := toString(choices[i])
		rand_val := getUniform(m, s.params, 0.0, 1.0, append_str)
		if rand_val <= pvalue {
			ret = append(ret, choices[i])
		}
	}
	return ret
}

type weightedChoice struct{ params map[string]interface{} }

func (s *weightedChoice) execute(m map[string]interface{}) interface{} {
	existOrPanic(m, []string{"choices", "unit", "weights"}, "WeightedChoice")
	weights := evaluate(m["weights"], s.params).([]interface{})
	sum, cweights := getCummulativeWeights(weights)
	stop_val := getUniform(m, s.params, 0.0, sum)
	choices := evaluate(m["choices"], s.params).([]interface{})
	for i := range cweights {
		if stop_val <= cweights[i] {
			return choices[i]
		}
	}
	return 0.0
}

type randomFloat struct{ params map[string]interface{} }

func (s *randomFloat) execute(m map[string]interface{}) interface{} {
	existOrPanic(m, []string{"unit"}, "RandomFloat")
	min_val := getOrElse(m, "min", 0.0)
	max_val := getOrElse(m, "max", 1.0)
	return getUniform(m, s.params, min_val.(float64), max_val.(float64))
}

type randomInteger struct{ params map[string]interface{} }

func (s *randomInteger) execute(m map[string]interface{}) interface{} {
	existOrPanic(m, []string{"unit"}, "RandomFloat")
	min_val := uint64(getOrElse(m, "min", 0.0).(float64))
	max_val := uint64(getOrElse(m, "max", 1.0).(float64))
	return min_val + getHash(m, s.params)%(max_val-min_val+1)
}

type sample struct{ params map[string]interface{} }

func (s *sample) execute(m map[string]interface{}) interface{} {
	existOrPanic(m, []string{"unit", "choices"}, "Sample")
	choices := evaluate(m["choices"], s.params).([]interface{})
	nchoices := len(choices)
	for i := nchoices - 1; i >= 0; i-- {
		j := int(getHash(m, s.params) % uint64(i+1))
		choices[i], choices[j] = choices[j], choices[i]
	}
	draws := int(getOrElse(m, "draws", float64(len(choices))).(float64))
	return choices[:draws]
}
