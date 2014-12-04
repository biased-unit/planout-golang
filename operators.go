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
	"fmt"
	"math/rand"
	"time"
)

type OpFunc func(map[string]interface{}) operator

var ops map[string]OpFunc

func init() {
	ops = map[string]OpFunc{
		"seq":             func(p map[string]interface{}) operator { return &seq{p} },
		"set":             func(p map[string]interface{}) operator { return &set{p} },
		"get":             func(p map[string]interface{}) operator { return &get{p} },
		"array":           func(p map[string]interface{}) operator { return &array{p} },
		"index":           func(p map[string]interface{}) operator { return &index{p} },
		"length":          func(p map[string]interface{}) operator { return &length{p} },
		"coalesce":        func(p map[string]interface{}) operator { return &coalesce{p} },
		"cond":            func(p map[string]interface{}) operator { return &cond{p} },
		">":               func(p map[string]interface{}) operator { return &gt{p} },
		">=":              func(p map[string]interface{}) operator { return &gte{p} },
		"<":               func(p map[string]interface{}) operator { return &lt{p} },
		"<=":              func(p map[string]interface{}) operator { return &lte{p} },
		"equals":          func(p map[string]interface{}) operator { return &eq{p} },
		"and":             func(p map[string]interface{}) operator { return &and{p} },
		"or":              func(p map[string]interface{}) operator { return &or{p} },
		"not":             func(p map[string]interface{}) operator { return &not{p} },
		"min":             func(p map[string]interface{}) operator { return &min{p} },
		"max":             func(p map[string]interface{}) operator { return &max{p} },
		"sum":             func(p map[string]interface{}) operator { return &sum{p} },
		"product":         func(p map[string]interface{}) operator { return &mul{p} },
		"negative":        func(p map[string]interface{}) operator { return &neg{p} },
		"round":           func(p map[string]interface{}) operator { return &round{p} },
		"%":               func(p map[string]interface{}) operator { return &mod{p} },
		"/":               func(p map[string]interface{}) operator { return &div{p} },
		"literal":         func(p map[string]interface{}) operator { return &literal{p} },
		"uniformChoice":   func(p map[string]interface{}) operator { return &uniformChoice{p} },
		"bernoulliTrial":  func(p map[string]interface{}) operator { return &bernoulliTrial{p} },
		"bernoulliFilter": func(p map[string]interface{}) operator { return &bernoulliFilter{p} },
		"weightedChoice":  func(p map[string]interface{}) operator { return &weightedChoice{p} },
		"randomInteger":   func(p map[string]interface{}) operator { return &randomInteger{p} },
		"randomFloat":     func(p map[string]interface{}) operator { return &randomFloat{p} },
		"sample":          func(p map[string]interface{}) operator { return &sample{p} },
		"return":          func(p map[string]interface{}) operator { return &stopPlanout{p} },
	}

	rand.Seed(time.Now().UTC().UnixNano())
}

type operator interface {
	execute(map[string]interface{}, *Interpreter) interface{}
}

func isOperator(expr interface{}) (OpFunc, bool) {
	js, ok := expr.(map[string]interface{})
	if !ok {
		return nil, false
	}

	opstr, exists := js["op"]
	if !exists {
		return nil, false
	}

	opfunc, exists := ops[opstr.(string)]
	if !exists {
		return nil, false
	}

	return opfunc, true
}

type seq struct{ params map[string]interface{} }

func (s *seq) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"seq"}, "Seq")
	return interpreter.evaluate(m["seq"], s.params)
}

type set struct{ params map[string]interface{} }

func (s *set) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"var", "value"}, "Set")
	lhs := m["var"].(string)
	s.params["salt"] = lhs
	value := interpreter.evaluate(m["value"], s.params)
	s.params[lhs] = value
	return true
}

type get struct{ params map[string]interface{} }

func (s *get) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"var"}, "Get")
	value, exists := s.params[m["var"].(string)]
	if !exists {
		panic(fmt.Sprintf("No input for key %v\n", m["var"]))
	}
	return value
}

type array struct{ params map[string]interface{} }

func (s *array) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Array")
	return interpreter.evaluate(m["values"], s.params)
}

type index struct{ params map[string]interface{} }

func (s *index) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"base", "index"}, "Index")
	base := interpreter.evaluate(m["base"], s.params)
	index := interpreter.evaluate(m["index"], s.params)

	base_arr, ok := base.([]interface{})
	if ok {
		index_num, ok := toNumber(index)
		if ok {
			if compare(len(base_arr), index_num) > 0 {
				return base_arr[int(index_num)]
			}
		}
	}

	base_map, ok := base.(map[string]interface{})
	if ok {
		index_str, ok := toString(index)
		if ok {
			return base_map[index_str]
		}
	}

	return nil
}

type length struct{ params map[string]interface{} }

func (s *length) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Length")
	values := interpreter.evaluate(m["values"], s.params).([]interface{})
	return len(values[0].([]interface{}))
}

type coalesce struct{ params map[string]interface{} }

func (s *coalesce) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Array")
	values := m["values"].([]interface{})
	nvalues := len(values)
	ret := make([]interface{}, 0, len(values))
	if nvalues != 1 {
		return ret
	}

	value := interpreter.evaluate(values[0], s.params).([]interface{})
	for i := range value {
		if value[i] != nil {
			ret = append(ret, value[i])
		}
	}
	return ret
}

type and struct{ params map[string]interface{} }

func (s *and) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "And")

	values := m["values"].([]interface{})
	if len(values) == 0 {
		return false
	}

	for i := range values {
		value := interpreter.evaluate(values[i], s.params)
		if isTrue(value) == false {
			return false
		}
	}
	return true
}

type or struct{ params map[string]interface{} }

func (s *or) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Or")

	values := m["values"].([]interface{})
	if len(values) == 0 {
		return false
	}

	for i := range values {
		value := interpreter.evaluate(values[i], s.params)
		if isTrue(value) {
			return true
		}
	}

	return false
}

type not struct{ params map[string]interface{} }

func (s *not) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Not")
	value := interpreter.evaluate(m["value"], s.params)
	return !isTrue(value)
}

type cond struct{ params map[string]interface{} }

func (s *cond) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"cond"}, "Condition")
	conditions := m["cond"].([]interface{})
	for i := range conditions {
		c := conditions[i].(map[string]interface{})
		existOrPanic(c, []string{"if", "then"}, "Condition")
		if_value := interpreter.evaluate(c["if"], s.params)
		if isTrue(if_value) {
			return interpreter.evaluate(c["then"], s.params)
		}
	}
	return true
}

type lt struct{ params map[string]interface{} }

func (s *lt) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "LessThan")
	lhs := interpreter.evaluate(m["left"], s.params)
	rhs := interpreter.evaluate(m["right"], s.params)
	return compare(lhs, rhs) < 0
}

type lte struct{ params map[string]interface{} }

func (s *lte) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "LessThanEqual")
	lhs := interpreter.evaluate(m["left"], s.params)
	rhs := interpreter.evaluate(m["right"], s.params)
	return compare(lhs, rhs) <= 0
}

type gt struct{ params map[string]interface{} }

func (s *gt) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "GreaterThan")
	lhs := interpreter.evaluate(m["left"], s.params)
	rhs := interpreter.evaluate(m["right"], s.params)
	return compare(lhs, rhs) > 0
}

type gte struct{ params map[string]interface{} }

func (s *gte) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "GreaterThanEqual")
	lhs := interpreter.evaluate(m["left"], s.params)
	rhs := interpreter.evaluate(m["right"], s.params)
	return compare(lhs, rhs) >= 0
}

type eq struct{ params map[string]interface{} }

func (s *eq) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "Equality")
	lhs := interpreter.evaluate(m["left"], s.params)
	rhs := interpreter.evaluate(m["right"], s.params)
	return compare(lhs, rhs) == 0
}

type min struct{ params map[string]interface{} }

func (s *min) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Minimum")
	values := interpreter.evaluate(m["values"], s.params).([]interface{})
	if len(values) == 0 {
		panic(fmt.Sprintf("Executing min() with no arguments\n"))
	}
	minval := values[0]
	for i := range values {
		if compare(values[i], minval) < 0 {
			minval = values[i]
		}
	}
	return minval
}

type max struct{ params map[string]interface{} }

func (s *max) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Maximum")
	values := interpreter.evaluate(m["values"], s.params).([]interface{})
	if len(values) == 0 {
		panic(fmt.Sprintf("Executing max() with no arguments\n"))
	}
	maxval := values[0]
	for i := range values {
		if compare(values[i], maxval) > 0 {
			maxval = values[i]
		}
	}
	return maxval
}

type sum struct{ params map[string]interface{} }

func (s *sum) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Addition")
	values := interpreter.evaluate(m["values"], s.params).([]interface{})
	return addSlice(values)
}

type mul struct{ params map[string]interface{} }

func (s *mul) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Multiplication")
	values := interpreter.evaluate(m["values"], s.params).([]interface{})
	return multiplySlice(values)
}

type neg struct{ params map[string]interface{} }

func (s *neg) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Negative")
	value := interpreter.evaluate(m["value"], s.params)
	values := []interface{}{-1.0, value}
	return multiplySlice(values)
}

type round struct{ params map[string]interface{} }

func (s *round) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Rounding")
	values := interpreter.evaluate(m["values"], s.params).([]interface{})
	ret := make([]interface{}, len(values))
	for i := range values {
		ret[i] = roundNumber(values[i])
	}
	return ret
}

type mod struct{ params map[string]interface{} }

func (s *mod) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "Modulo")
	var ret int64 = 0
	lhs := interpreter.evaluate(m["left"], s.params).(float64)
	rhs := interpreter.evaluate(m["right"], s.params).(float64)
	ret = int64(lhs) % int64(rhs)
	return float64(ret)
}

type div struct{ params map[string]interface{} }

func (s *div) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "Division")
	var ret float64 = 0
	lhs := interpreter.evaluate(m["left"], s.params).(float64)
	rhs := interpreter.evaluate(m["right"], s.params).(float64)
	ret = lhs / rhs
	return ret
}

type literal struct{ params map[string]interface{} }

func (s *literal) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Literal")
	return m["value"]
}

type stopPlanout struct{ params map[string]interface{} }

func (s *stopPlanout) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Literal")
	value := interpreter.evaluate(m["value"], s.params)
	m["in_experiment"] = value
	panic(nil)
}
