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
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
)

func existOrPanic(m map[string]interface{}, keys []string, opstr string) bool {
	for i := range keys {
		_, exist := m[keys[i]]
		if !exist {
			panic(fmt.Sprintf("Operator %s: Missing operand %s\n", opstr, keys[i]))
		}
	}
	return true
}

func getOrElse(m map[string]interface{}, key string, def interface{}) interface{} {
	v, exists := m[key]
	if !exists {
		return def
	}
	return v
}

func compare(lhs, rhs interface{}) int {
	l_str, l_ok := lhs.(string)
	r_str, r_ok := rhs.(string)
	if l_ok && r_ok {
		return cmpString(l_str, r_str)
	}

	l_num, l_ok := toNumber(lhs)
	r_num, r_ok := toNumber(rhs)
	if l_ok && r_ok {
		return cmpFloat(l_num, r_num)
	}

	panic(fmt.Sprintf("Compare: Unsupported type\n"))
}

func isTrue(value interface{}) bool {
	switch value.(type) {
	case bool:
		return value.(bool)
	case string:
		return len(value.(string)) > 0
	}

	n, ok := toNumber(value)
	if ok {
		return cmpFloat(n, 1.0) == 0
	}

	panic(fmt.Sprintf("IsTrue: Unsupported type\n"))
}

func cmpFloat(lhs, rhs float64) int {
	ret := 0
	if math.Abs(lhs-rhs) < 0.0001 {
		ret = 0
	} else if lhs < rhs {
		ret = -1
	} else {
		ret = 1
	}
	return ret
}

func cmpString(lhs, rhs string) int {
	ret := 0
	if lhs == rhs {
		ret = 0
	} else if lhs < rhs {
		ret = -1
	} else {
		ret = 1
	}
	return ret
}

func add(x, y interface{}) interface{} {
	x_num, x_ok := toNumber(x)
	y_num, y_ok := toNumber(y)
	if x_ok && y_ok {
		return x_num + y_num
	}

	x_str, x_ok := toString(x)
	y_str, y_ok := toString(y)
	if x_ok && y_ok {
		return x_str + y_str
	}

	panic("Addition: Unsupported type")
}

func addSlice(x []interface{}) interface{} {
	ret := x[0]
	for i := range x {
		if i != 0 {
			ret = add(ret, x[i])
		}
	}
	return ret
}

func multiply(x, y interface{}) interface{} {

	x_num, x_ok := toNumber(x)
	y_num, y_ok := toNumber(y)
	if x_ok && y_ok {
		return x_num * y_num
	}

	panic("Multiplication: Unsupported type")
}

func multiplySlice(x []interface{}) interface{} {
	ret := x[0]
	for i := range x {
		if i != 0 {
			ret = multiply(ret, x[i])
		}
	}
	return ret
}

func generateUnitStr(units interface{}) string {

	unit_arr, ok := units.([]interface{})
	if ok {
		var buffer bytes.Buffer
		n := len(unit_arr)
		s, _ := toString(unit_arr[0])
		buffer.WriteString(s)
		for i := 1; i < n; i++ {
			buffer.WriteString(".")
			s, _ = toString(unit_arr[i])
			buffer.WriteString(s)
		}
		return buffer.String()
	}

	unit_str, ok := toString(units)
	if ok {
		return unit_str
	}

	return ""
}

func getCummulativeWeights(weights []interface{}) (float64, []float64) {
	nweights := len(weights)
	cweights := make([]float64, nweights)
	sum := 0.0
	for i := range weights {
		sum = sum + weights[i].(float64)
		cweights[i] = sum
	}
	return sum, cweights
}

func generateString() string {
	s := make([]byte, 10)
	for j := 0; j < 10; j++ {
		s[j] = 'a' + byte(rand.Int()%26)
	}
	return string(s)
}

func toString(unit interface{}) (string, bool) {

	unit_str, ok := unit.(string)
	if ok {
		return unit_str, true
	}

	unit_num, ok := toNumber(unit)
	if ok {
		ret_str := strconv.FormatFloat(unit_num, 'f', -1, 64)
		return ret_str, true
	}

	return "", false
}

func toNumber(value interface{}) (float64, bool) {
	switch value := value.(type) {
	case float32, float64:
		x, ok := value.(float64)
		if !ok {
			x = float64(value.(float32))
		}
		return x, true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(value)
		return float64(i.Int()), true
	case bool:
		if value {
			return 1, true
		} else {
			return 0, true
		}
	}
	return 0.0, false
}

func roundNumber(value interface{}) interface{} {

	value_num, ok := toNumber(value)
	if ok {
		x := value_num
		x_floor := math.Floor(x)
		if math.Abs(x-x_floor) < 0.5 {
			return x_floor
		}
		return math.Ceil(x)
	}

	panic(fmt.Sprintf("Rounding operation: Unsupported type %v\n", value))
}
