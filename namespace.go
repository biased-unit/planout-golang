package planout

import (
	"fmt"
	"sort"
)

type Namespace interface {
	AddExperiment(name string, code map[string]interface{}, segments int) error
	RemoveExperiment(name string)
}

type SimpleNamespace struct {
	Name               string
	PrimaryUnit        string
	NumSegments        int
	Inputs             map[string]interface{}
	segmentAllocations map[uint64]string
	availableSegments  []int
	currentExperiments map[string]interface{}
	defaultExperiment  map[string]interface{}
}

func NewSimpleNamespace(name string, numSegments int, primaryUnit string, inputs map[string]interface{}) SimpleNamespace {
	avail := make([]int, 0, numSegments)
	for i := 0; i < numSegments; i++ {
		avail = append(avail, i)
	}

	return SimpleNamespace{
		Name:               name,
		PrimaryUnit:        primaryUnit,
		NumSegments:        numSegments,
		Inputs:             inputs,
		segmentAllocations: make(map[uint64]string),
		availableSegments:  avail,
		currentExperiments: make(map[string]interface{}),
	}
}

func (n *SimpleNamespace) Run() *Interpreter {
	// Is the unit allocated to an experiment ?
	interpreter := &Interpreter{
		Name:      n.Name,
		Salt:      n.Name,
		Code:      n.defaultExperiment,
		Evaluated: false,
		Inputs:    n.Inputs,
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
	}

	if name, ok := n.segmentAllocations[n.getSegment()]; ok {
		interpreter.Name = n.Name + "-" + name
		interpreter.Salt = n.Name + "." + name
		interpreter.Code = n.currentExperiments[name]
	}

	interpreter.Run()
	return interpreter
}

func (n *SimpleNamespace) AddDefaultExperiment(code map[string]interface{}) {
	n.defaultExperiment = code
}

func (n *SimpleNamespace) AddExperiment(name string, code map[string]interface{}, segments int) error {
	avail := len(n.availableSegments)
	if avail < segments {
		return fmt.Errorf("Not enough segments available %v to add the new experiment %v\n", avail, name)
	}

	if _, ok := n.currentExperiments[name]; ok {
		return fmt.Errorf("There is already and experiment called %s\n", name)
	}

	n.allocateExperiment(name, segments)

	n.currentExperiments[name] = code
	return nil
}

func (n *SimpleNamespace) RemoveExperiment(name string) error {
	_, exists := n.currentExperiments[name]
	if !exists {
		return fmt.Errorf("Experiment %v does not exists in the namespace\n", name)
	}

	segmentsToFree := make([]int, 0, n.NumSegments)
	for i := range n.segmentAllocations {
		if n.segmentAllocations[i] == name {
			segmentsToFree = append(segmentsToFree, int(i))
		}
	}

	for i := range segmentsToFree {
		n.availableSegments = append(n.availableSegments, segmentsToFree[i])
	}

	sort.Ints(n.availableSegments)

	delete(n.currentExperiments, name)
	return nil
}

func (n *SimpleNamespace) allocateExperiment(name string, segments int) {
	// Sample(choices=available_segments, draws=segments, unit=name)
	expt := &Interpreter{
		Salt:      n.Name,
		Evaluated: false,
		Inputs:    n.Inputs,
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
	}

	// Compile Sample operator
	var availableSegmentsAsInterface []interface{} = make([]interface{}, len(n.availableSegments))
	for i, d := range n.availableSegments {
		availableSegmentsAsInterface[i] = d
	}

	args := make(map[string]interface{})
	args["choices"] = availableSegmentsAsInterface
	args["unit"] = name
	args["salt"] = n.Name
	args["draws"] = segments
	s := &sample{}
	shuffle := s.execute(args, expt).([]interface{})

	// Allocate sampled_segments to experiment
	// Remove segment from available_segments
	for i := range shuffle {
		j := shuffle[i].(int)
		n.segmentAllocations[uint64(j)] = name
		n.availableSegments = deallocateSegments(n.availableSegments, j)
	}
}

func (n *SimpleNamespace) getSegment() uint64 {
	// generate random integer min=0, max=num_segments, unit=primary_unit
	// RandomInteger(min=0, max=self.num_segments, unit=itemgetter(*self.primary_unit)(self.inputs))
	expt := &Interpreter{
		Salt:      n.Name,
		Evaluated: false,
		Inputs:    n.Inputs,
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
	}

	// Compile RandomInteger operator
	args := make(map[string]interface{})
	args["salt"] = n.Name
	args["min"] = 0
	args["max"] = n.NumSegments - 1
	args["unit"] = n.Inputs[n.PrimaryUnit]
	s := &randomInteger{}
	return s.execute(args, expt).(uint64)
}

func deallocateSegments(allocated []int, segmentToRemove int) []int {
	for i := range allocated {
		if allocated[i] == segmentToRemove {
			outputs := make([]int, 0, len(allocated)-1)
			outputs = append(outputs, allocated[:i]...)
			outputs = append(outputs, allocated[i+1:]...)
			return outputs
		}
	}
	return allocated
}
