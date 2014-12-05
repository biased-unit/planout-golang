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
)

type Interpreter struct {
	experiment_salt            string
	inputs, outputs, overrides map[string]interface{}
	evaluated                  bool
}

func (interpreter *Interpreter) get(name string) interface{} {
	value, ok := interpreter.overrides[name]
	if ok {
		return value
	}

	value, ok = interpreter.inputs[name]
	if ok {
		return value
	}

	value, ok = interpreter.outputs[name]
	if ok {
		return value
	}
	return nil
}

func (interpreter *Interpreter) set(name string, value interface{}) {
	interpreter.outputs[name] = value
}

func (interpreter *Interpreter) getOverrides() map[string]interface{} {
	return interpreter.overrides
}

func (interpreter *Interpreter) hasOverrides(name string) bool {
	_, exists := interpreter.overrides[name]
	return exists
}

func (interpreter *Interpreter) evaluate(code interface{}, params map[string]interface{}) interface{} {

	js, ok := code.(map[string]interface{})
	if ok {
		opconstruct, exists := isOperator(js)
		if exists {
			e := opconstruct(params)
			return e.execute(js, interpreter)
		}
	}

	arr, ok := code.([]interface{})
	if ok {
		v := make([]interface{}, len(arr))
		for i := range arr {
			v[i] = interpreter.evaluate(arr[i], params)
		}
		return v
	}

	return code
}

func (interpreter *Interpreter) Run(code interface{}, params map[string]interface{}) (map[string]interface{}, bool) {

	if interpreter.evaluated {
		return interpreter.outputs, true
	}

	defer func() (map[string]interface{}, bool) {
		if r := recover(); r != nil {
			fmt.Println("Recovered ", r)
			return nil, false
		}
		interpreter.evaluated = true
		return interpreter.outputs, true
	}()

	interpreter.evaluate(code, params)

	return interpreter.outputs, true
}
