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
)

type PlanOutCode interface {
	Run() (map[string]interface{}, bool)
}

type Interpreter struct {
	Name                       string
	Salt                       string
	Inputs, Outputs, Overrides map[string]interface{}
	Code                       interface{}
	Evaluated, InExperiment    bool
	ParameterSalt              string
}

func (interpreter *Interpreter) Run(force ...bool) (map[string]interface{}, bool) {

	if len(force) > 0 && force[0] == false {
		if interpreter.Evaluated {
			return interpreter.Outputs, true
		}
	}

	defer func() (map[string]interface{}, bool) {
		if r := recover(); r != nil {
			fmt.Println("Recovered ", r)
			return nil, false
		}
		interpreter.Evaluated = true
		return interpreter.Outputs, true
	}()

	interpreter.evaluate(interpreter.Code)
	return interpreter.Outputs, true
}

// ReSet方法重置experiment的三个参数Inputs、Overrides、Outputs，保留namespace的分区信息，就可以复用该namespace，
// 而不是每次都重新生成一个namespace，耗费性能
func (interpreter *Interpreter) ReSet() {
	// 置空参数
	interpreter.Inputs, interpreter.Overrides, interpreter.Outputs = make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{})
	interpreter.Evaluated, interpreter.InExperiment = false, false
}

func (interpreter *Interpreter) Get(name string) (interface{}, bool) {
	value, ok := interpreter.Overrides[name]
	if ok {
		return value, true
	}

	value, ok = interpreter.Inputs[name]
	if ok {
		return value, true
	}

	value, ok = interpreter.Outputs[name]
	if ok {
		return value, true
	}
	return nil, false
}

func (interpreter *Interpreter) set(name string, value interface{}) {
	interpreter.Outputs[name] = value
}

func (interpreter *Interpreter) getOverrides() map[string]interface{} {
	return interpreter.Overrides
}

func (interpreter *Interpreter) hasOverrides(name string) bool {
	_, exists := interpreter.Overrides[name]
	return exists
}

func (interpreter *Interpreter) evaluate(code interface{}) interface{} {

	js, ok := code.(map[string]interface{})
	if ok {
		opptr, exists := isOperator(js)
		if exists {
			return opptr.execute(js, interpreter)
		}
	}

	arr, ok := code.([]interface{})
	if ok {
		if len(arr) == 1 {
			_, ok := arr[0].(map[string]interface{})
			if ok {
				_, ok := isOperator(arr[0])
				if ok {
					return interpreter.evaluate(arr[0])
				}
			}
		}
		v := make([]interface{}, len(arr))
		for i := range arr {
			v[i] = interpreter.evaluate(arr[i])
		}
		return v
	}

	return code
}
