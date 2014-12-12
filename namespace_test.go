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
	"testing"
)

func TestSimpleNamespace(t *testing.T) {
	js1 := readTest("test/simple_ops.json")
	js2 := readTest("test/random_ops.json")

	segments := make([]string, 100)
	avail := make([]interface{}, 0, 100)
	for i := 0; i < 100; i++ {
		avail = append(avail, i)
	}

	inputs := make(map[string]interface{})
	inputs["userid"] = generateString()

	n := &SimpleNamespace{
		Name:               "simple_namespace",
		NumSegments:        100,
		PrimaryUnit:        "userid",
		SegmentAllocations: segments,
		AvailableSegments:  avail,
		CurrentExperiments: map[string]PlanOutCode{},
		Inputs:             inputs,
	}

	n.addExperiment("simple ops", js1, 50)
	n.addExperiment("random ops", js2, 50)
	// n.removeExperiment("simple ops")

	fmt.Printf("Random segment alloc: %v\n", n.getSegment())
}
