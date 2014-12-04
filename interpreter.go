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

type PlanoutContext struct {
	env             map[string]interface{}
	override        map[string]interface{}
	experiment_salt string
	in_experiment   bool
}

func (ctx *PlanoutContext) getEnv(s string) interface{} {
	return ctx.env[s]
}

func (ctx *PlanoutContext) setEnv(k string, v interface{}) {
	ctx.env[k] = v
}

func (ctx *PlanoutContext) getOverride(s string) interface{} {
	return ctx.override[s]
}

func (ctx *PlanoutContext) getExperimentSalt() string {
	return ctx.experiment_salt
}

func (ctx *PlanoutContext) isInExperiment() bool {
	return ctx.in_experiment
}

type Experiment struct {
	ctx *PlanoutContext
}

func evaluate(code interface{}, params map[string]interface{}) interface{} {

	js, ok := code.(map[string]interface{})
	if ok {
		opconstruct, exists := isOperator(js)
		if exists {
			e := opconstruct(params)
			return e.execute(js)
		}
	}

	arr, ok := code.([]interface{})
	if ok {
		v := make([]interface{}, len(arr))
		for i := range arr {
			v[i] = evaluate(arr[i], params)
		}
		return v
	}

	return code
}

func (e *Experiment) Run(code interface{}, params map[string]interface{}) bool {

	defer func() bool {
		if r := recover(); r != nil {
			fmt.Println("Recovered ", r)
			return false
		}
		return true
	}()

	evaluate(code, params)

	return true
}
