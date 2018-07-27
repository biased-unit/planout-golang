/*
 * Copyright 2015 Biasedunit
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
	"testing"
)

func TestFixture01(t *testing.T) {
    testFixture("test/fixtures/1.json", t);
}

func testFixture(fixture string, t *testing.T) {
	js := readTest(fixture)

	var runs int = 100
	var userid int = 123454

	for i := 0; i < runs; i++ {
		params := make(map[string]interface{})
		params["userid"] = userid

		expt := &Interpreter{
			Salt:      "foo",
			Evaluated: false,
			Inputs:    params,
			Outputs:   map[string]interface{}{},
			Overrides: map[string]interface{}{},
			Code:      js,
		}

		_, ok := expt.Run()
		if !ok {
			t.Errorf("Error running experiment %s\n", fixture);
			return
		}
	}
}
