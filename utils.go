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

func assertKind(lhs, rhs interface{}, opstr string) {
	lhstype, rhstype := reflect.ValueOf(lhs), reflect.ValueOf(rhs)
	if lhstype.Kind() != rhstype.Kind() {
		panic(fmt.Sprintf("%v: Type mismatch between LHS %v (%v) and RHS %v (%v)\n", opstr, lhs, lhstype.Kind(), rhs, rhstype.Kind()))
	}
}

func compare(lhs, rhs interface{}) int {
	switch lhs.(type) {
	case string:
		return cmpString(lhs.(string), rhs.(string))
	default:
		return cmpFloat(toNumber(lhs), toNumber(rhs))
	}
	panic(fmt.Sprintf("Compare: Unsupported type\n"))
}

func isTrue(value interface{}) bool {
	switch value.(type) {
	case string:
		return len(value.(string)) > 0
	case bool:
		return value.(bool)
	default:
		n := toNumber(value)
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

func cmpInt(lhs, rhs int64) int {
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

	assertKind(x, y, "Addition")

	a, b := reflect.ValueOf(x), reflect.ValueOf(y)

	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() + b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return a.Uint() + b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() + b.Float()
	case reflect.String:
		return a.String() + b.String()
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

	assertKind(x, y, "Multiplication")

	a, b := reflect.ValueOf(x), reflect.ValueOf(y)

	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() * b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return a.Uint() * b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() * b.Float()
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
	unitval := reflect.ValueOf(units)
	switch unitval.Kind() {
	case reflect.Array, reflect.Slice:
		v := units.([]interface{})
		n := len(v)
		var buffer bytes.Buffer
		buffer.WriteString(v[0].(string))
		for i := 0; i < n; i++ {
			if i != 0 {
				buffer.WriteString(".")
				buffer.WriteString(v[i].(string))
			}
		}
		return buffer.String()
	case reflect.String:
		return unitval.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(unitval.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(unitval.Float(), 'f', -1, 64)
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

func toString(unit interface{}) string {
	unitval := reflect.ValueOf(unit)
	switch unitval.Kind() {
	case reflect.String:
		return unitval.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(unitval.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(unitval.Float(), 'f', -1, 64)
	}
	return ""
}

func toNumber(value interface{}) float64 {
	switch value := value.(type) {
	case float32, float64:
		x, ok := value.(float64)
		if !ok {
			x = float64(value.(float32))
		}
		return x
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(value)
		return float64(i.Int())
	default:
		panic(fmt.Sprintf("Rounding operation: Unsupported type %v\n", value))
	}
	return 0.0
}

func roundNumber(value interface{}) interface{} {
	switch value := value.(type) {
	case float32, float64:
		x, ok := value.(float64)
		if !ok {
			x = float64(value.(float32))
		}
		x_floor := math.Floor(x)
		if math.Abs(x-x_floor) < 0.5 {
			return x_floor
		}
		return math.Ceil(x)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return value
	default:
		panic(fmt.Sprintf("Rounding operation: Unsupported type %v\n", value))
	}
	return value
}
