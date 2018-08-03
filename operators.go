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
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

var ops map[string]operator

func init() {
	ops = map[string]operator{
		"seq":             &seq{},
		"set":             &set{},
		"get":             &get{},
		"array":           &array{},
		"map":             &dict{},
		"index":           &index{},
		"length":          &length{},
		"coalesce":        &coalesce{},
		"cond":            &cond{},
		">":               &gt{},
		">=":              &gte{},
		"<":               &lt{},
		"<=":              &lte{},
		"equals":          &eq{},
		"and":             &and{},
		"or":              &or{},
		"not":             &not{},
		"min":             &min{},
		"max":             &max{},
		"sum":             &sum{},
		"product":         &mul{},
		"negative":        &neg{},
		"round":           &round{},
		"%":               &mod{},
		"/":               &div{},
		"literal":         &literal{},
		"uniformChoice":   &uniformChoice{},
		"bernoulliTrial":  &bernoulliTrial{},
		"bernoulliFilter": &bernoulliFilter{},
		"weightedChoice":  &weightedChoice{},
		"randomInteger":   &randomInteger{},
		"randomFloat":     &randomFloat{},
		"sample":          &sample{},
		"return":          &stopPlanout{},
	}

	rand.Seed(time.Now().UTC().UnixNano())
}

type operator interface {
	execute(map[string]interface{}, *Interpreter) interface{}
}

func isOperator(expr interface{}) (operator, bool) {
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

type seq struct{}

func (s *seq) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"seq"}, "Seq")
	return interpreter.evaluate(m["seq"])
}

type set struct{}

func (s *set) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"var", "value"}, "Set")
	lhs := m["var"].(string)
	interpreter.ParameterSalt = lhs
	value := interpreter.evaluate(m["value"])
	interpreter.Outputs[lhs] = value
	return true
}

type get struct{}

func (s *get) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"var"}, "Get")
	value, exists := interpreter.Get(m["var"].(string))
	if !exists {
		panic(fmt.Sprintf("No input for key %v\n", m["var"]))
	}

	return value
}

type array struct{}

func (s *array) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Array")
	ret := interpreter.evaluate(m["values"])
	return ret
}

type dict struct{}

func (s *dict) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	dictionary := make(map[string]interface{})
	for k, v := range m {
		if k != "op" {
			dictionary[k] = interpreter.evaluate(v)
		}
	}
	return dictionary
}

type index struct{}

func (s *index) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"base", "index"}, "Index")
	base := interpreter.evaluate(m["base"])
	index := interpreter.evaluate(m["index"])

	base_type := reflect.ValueOf(base)
	for {
		if base_type.Kind() != reflect.Ptr {
			break
		}
		base_type = base_type.Elem()
	}

	if base_type.Kind() == reflect.Array || base_type.Kind() == reflect.Slice {
		index_num, ok := toNumber(index)
		if !ok {
			panic("Indexing an array with a non-number")
		}
		return unwrapValue(base_type.Index(int(index_num)))
	}

	if base_type.Kind() == reflect.Map {
		return unwrapValue(base_type.MapIndex(reflect.ValueOf(index)))
	}

	if index_str, isStr := toString(index); isStr {
		if base_type.Kind() != reflect.Invalid && base_type.Kind() == reflect.Struct {
			if field, hasField := base_type.Type().FieldByName(strings.Title(index_str)); hasField {
				// Only use exported fields
				if field.PkgPath == "" {
					value := base_type.FieldByIndex(field.Index)
					return unwrapValue(value)
				}
			}
		}
	}

	return nil
}

func unwrapValue(value reflect.Value) interface{} {
	switch value.Kind() {
	case reflect.Int:
		return int(value.Int())
	case reflect.Int8:
		return int8(value.Int())
	case reflect.Int16:
		return int16(value.Int())
	case reflect.Int32:
		return int32(value.Int())
	case reflect.Int64:
		return value.Int()
	case reflect.Uint:
		return uint(value.Uint())
	case reflect.Uint8:
		return uint8(value.Uint())
	case reflect.Uint16:
		return uint16(value.Uint())
	case reflect.Uint32:
		return uint32(value.Uint())
	case reflect.Uint64:
		return uint64(value.Uint())
	case reflect.Float32:
		return float32(value.Float())
	case reflect.Float64:
		return value.Float()
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return value.Bool()
	default:
		if value.IsValid() {
			return value.Interface()
		} else {
			return nil
		}
	}
}

type length struct{}

func (s *length) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Length")
	values := interpreter.evaluate(m["values"])
	return len(values.([]interface{}))
}

type coalesce struct{}

func (s *coalesce) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Coalesce")

	raw_input_values := interpreter.evaluate(m["values"]).([]interface{})
	nvalues := len(raw_input_values)
	ret := make([]interface{}, 0, nvalues)

	for i := range raw_input_values {
		if raw_input_values[i] != nil {
			ret = append(ret, raw_input_values[i])
		}
	}

	return ret
}

type and struct{}

func (s *and) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "And")

	values := m["values"].([]interface{})
	if len(values) == 0 {
		return false
	}

	for i := range values {
		value := interpreter.evaluate(values[i])
		if isTrue(value) == false {
			return false
		}
	}
	return true
}

type or struct{}

func (s *or) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Or")

	values := m["values"].([]interface{})
	if len(values) == 0 {
		return false
	}

	for i := range values {
		value := interpreter.evaluate(values[i])
		if isTrue(value) {
			return true
		}
	}

	return false
}

type not struct{}

func (s *not) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Not")
	value := interpreter.evaluate(m["value"])
	return !isTrue(value)
}

type cond struct{}

func (s *cond) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"cond"}, "Condition")
	conditions := m["cond"].([]interface{})
	for i := range conditions {
		c := conditions[i].(map[string]interface{})
		existOrPanic(c, []string{"if", "then"}, "Condition")
		if_value := interpreter.evaluate(c["if"])
		if isTrue(if_value) {
			return interpreter.evaluate(c["then"])
		}
	}
	return true
}

type lt struct{}

func (s *lt) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "LessThan")
	lhs := interpreter.evaluate(m["left"])
	rhs := interpreter.evaluate(m["right"])
	return compare(lhs, rhs) < 0
}

type lte struct{}

func (s *lte) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "LessThanEqual")
	lhs := interpreter.evaluate(m["left"])
	rhs := interpreter.evaluate(m["right"])
	return compare(lhs, rhs) <= 0
}

type gt struct{}

func (s *gt) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "GreaterThan")
	lhs := interpreter.evaluate(m["left"])
	rhs := interpreter.evaluate(m["right"])
	return compare(lhs, rhs) > 0
}

type gte struct{}

func (s *gte) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "GreaterThanEqual")
	lhs := interpreter.evaluate(m["left"])
	rhs := interpreter.evaluate(m["right"])
	return compare(lhs, rhs) >= 0
}

type eq struct{}

func (s *eq) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "Equality")
	lhs := interpreter.evaluate(m["left"])
	rhs := interpreter.evaluate(m["right"])
	return compare(lhs, rhs) == 0
}

type min struct{}

func (s *min) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Minimum")
	values := interpreter.evaluate(m["values"]).([]interface{})
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

type max struct{}

func (s *max) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Maximum")
	values := interpreter.evaluate(m["values"]).([]interface{})
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

type sum struct{}

func (s *sum) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Addition")
	values := interpreter.evaluate(m["values"]).([]interface{})
	return addSlice(values)
}

type mul struct{}

func (s *mul) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Multiplication")
	values := interpreter.evaluate(m["values"]).([]interface{})
	return multiplySlice(values)
}

type neg struct{}

func (s *neg) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Negative")
	value := interpreter.evaluate(m["value"])
	values := []interface{}{-1.0, value}
	return multiplySlice(values)
}

type round struct{}

func (s *round) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"values"}, "Rounding")
	values := interpreter.evaluate(m["values"]).([]interface{})
	ret := make([]interface{}, len(values))
	for i := range values {
		ret[i] = roundNumber(values[i])
	}
	return ret
}

type mod struct{}

func (s *mod) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "Modulo")
	var ret int64 = 0
	lhs := interpreter.evaluate(m["left"]).(float64)
	rhs := interpreter.evaluate(m["right"]).(float64)
	ret = int64(lhs) % int64(rhs)
	return float64(ret)
}

type div struct{}

func (s *div) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"left", "right"}, "Division")
	var ret float64 = 0
	lhs := interpreter.evaluate(m["left"]).(float64)
	rhs := interpreter.evaluate(m["right"]).(float64)
	ret = lhs / rhs
	return ret
}

type literal struct{}

func (s *literal) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Literal")
	return m["value"]
}

type stopPlanout struct{}

func (s *stopPlanout) execute(m map[string]interface{}, interpreter *Interpreter) interface{} {
	existOrPanic(m, []string{"value"}, "Literal")
	value := interpreter.evaluate(m["value"])
	interpreter.InExperiment = isTrue(value)
	panic(nil)
}
