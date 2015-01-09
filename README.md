[![Build Status](https://travis-ci.org/URXtech/planout-golang.svg?branch=master)](https://travis-ci.org/URXtech/planout-golang)

(Multi Variate Testing) interpreter for [PlanOut](http://github.com/facebook/planout) code written in Golang

# What what ?
[PlanOut](http://github.com/facebook/planout) is a framework for providing randomised parameter assignment for controlling parameters and defaults used in code. It exists as a combination of both a generalised methodology and as a DSL for constructing online field experiments.

An excellent introduction can be found both in the original research, [Designing and Deploying Online Field Experiments (Bakshy, Eckles and Bernstein)](http://arxiv.org/pdf/1409.3174v1.pdf), and in the following lecture https://www.youtube.com/watch?v=Ayd4sqPH2DE.

# So what is this ?
This is an interpreter that provides the basic functionality for running [PlanOut](http://github.com/facebook/planout) interpreter code, allowing for integrating experiments into GoLang applications.

This is not a full implementation of a complete [PlanOut](http://github.com/facebook/planout) stack, as such it lacks the compiler needed to turn [PlanOut](http://github.com/facebook/planout) DSL into the interpreter code, as well as the general test management tooling needed.

Much of the additional tooling for [PlanOut](http://github.com/facebook/planout) can be found the in original project.

This code will however run [PlanOut](http://github.com/facebook/planout) programs in an idiomatic Golang fashion.

# How to run a basic experiment ?
Here's an example program that consumes compiled [PlanOut](http://github.com/facebook/planout) code and executes the associated experiment using the Golang interpreter.

```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"github.com/URXtech/planout-golang"
)

// Helper function to generate random string.
func generateString() string {
	s := make([]byte, 10)
	for j := 0; j < 10; j++ {
		s[j] = 'a' + byte(rand.Int()%26)
	}
	return string(s)
}

func main() {
	// Read PlanOut code from file on disk.
	data, _ := ioutil.ReadFile("test/simple_ops.json")

	// The PlanOut code is expected to use json.
	// This format is the same as the output of
	// the PlanOut compiler webapp
	// http://facebook.github.io/planout/demo/planout-compiler.html
	var js map[string]interface{}
	json.Unmarshal(data, &js)

	// Set the necessary input parameters required to run
	// the experiments. For instance, simple_ops.json expects
	// the value for 'userid' to be set.
	params := make(map[string]interface{})
	params["experiment_salt"] = "expt"
	params["userid"] = generateString()

	// Construct an instance of the Interpreter object.
	// Initialize Salt and set Inputs to params.
	expt := &goplanout.Interpreter{
		Salt: "global_salt",
		Evaluated:      false,
		Inputs:         params,
		Outputs:        map[string]interface{}{},
		Overrides:      map[string]interface{}{},
        Code: js,
	}
	
	// Call the Run() method on the Interpreter instance.
	// The output of the run will contain the dictionary 
	// of variables and associated values that were evaluated
	// as part of the experiment.
	output, ok := expt.Run()
	if !ok {
		fmt.Println("Failed to run the experiment")
	} else {
		fmt.Printf("Params: %v\n", params)
	}
	
	fmt.Println(output)
}
```

Suppose we want to run the following experiment:
```go
id = uniformChoice(choices=[1, 2, 3, 4], unit=userid);
```

The [PlanOut](http://github.com/facebook/planout) code generated by the compiler looks like:

```json
{
  "op": "seq",
  "seq": [
    {
      "op": "set",
      "var": "id",
      "value": {
        "choices": {
          "op": "array",
          "values": [
            1,
            2,
            3,
            4
          ]
        },
        "unit": {
          "op": "get",
          "var": "userid"
        },
        "op": "uniformChoice"
      }
    }
  ]
}
```

Each execution of the above experiment will result in setting the variable 'id'. The output to stdout will look like:

```go
Params: map[experiment_salt:expt userid:noocavzddw salt:id id:2]
Params: map[experiment_salt:expt userid:cuncjyqmmz salt:id id:1]
```

# How to run a experiments in an allocated namespace ?
This example consumes multiple compiled [PlanOut](http://github.com/facebook/planout) experiments and executes within a namespace.

```go
package main

func main() {
	js1 := readTest("test/simple_ops.json")
	js2 := readTest("test/random_ops.json")
	js3 := readTest("test/simple.json")

	inputs := make(map[string]interface{})
	inputs["userid"] = "test-id"

	n := planout.NewSimpleNamespace("simple_namespace", 100, "userid", inputs)
	n.AddExperiment("simple ops", js1, 10)
	n.AddExperiment("random ops", js2, 10)
	n.AddExperiment("simple", js3, 80)

    out, ok = := n.Run()
}

```
