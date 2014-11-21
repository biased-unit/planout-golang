planout-go-interpreter
======================

(Multi Variate Testing) Interpreter for Planout code written in Golang

```
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/URXtech/planout-go-interpreter"
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
	// Read planout code from file on disk.
	data, _ := ioutil.ReadFile("test/simple_ops.json")

	// The planout code is expected to use json.
	// This format is the same as the output of
	// the planout compiler webapp
	// http://facebook.github.io/planout/demo/planout-compiler.html
	var js map[string]interface{}
	json.Unmarshal(data, &js)

	// Set the necessary input parameters required to run
	// the experiments. For instance, simple_ops.json expects
	// the value for 'userid' to be set.
	params := make(map[string]interface{})
	params["experiment_salt"] = "expt"
	params["userid"] = generateString()

  // Calling goplanout.Experiment runs the planout code
  // given the input params. It returns true if no errors
  // were encountered during the run. False, otherwise.
  // During the run, any variables that are evaluated
  // are added to the dictionary along with its value.
	ok := goplanout.Experiment(js, params)
	if !ok {
		fmt.Println("Failed to run the experiment")
	} else {
		fmt.Printf("Params: %v\n", params)
	}
}
```
