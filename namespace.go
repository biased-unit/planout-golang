package goplanout

import (
	"fmt"
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
	availableSegments  []interface{}
	currentExperiments map[string]interface{}
}

func NewSimpleNamespace(name string, numSegments int, primaryUnit string, inputs map[string]interface{}) SimpleNamespace {
	avail := make([]interface{}, 0, numSegments)
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

func (n *SimpleNamespace) Run() (map[string]interface{}, bool) {
	out := make(map[string]interface{})
	// Is the unit allocated to an experiment ?
	if name, ok := n.segmentAllocations[n.getSegment()]; ok {
		interpreter := &Interpreter{
			Name:      n.Name + "-" + name,
			Salt:      n.Name + "." + name,
			Code:      n.currentExperiments[name],
			Evaluated: false,
			Inputs:    n.Inputs,
			Outputs:   out,
			Overrides: map[string]interface{}{},
		}
		return interpreter.Run()
	}

	return out, true
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

	segmentsToFree := make([]uint64, 0, n.NumSegments)
	for i := range n.segmentAllocations {
		if n.segmentAllocations[i] == name {
			segmentsToFree = append(segmentsToFree, uint64(i))
		}
	}

	for i := range segmentsToFree {
		n.availableSegments = append(n.availableSegments, segmentsToFree[i])
	}

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
	args := make(map[string]interface{})
	args["choices"] = n.availableSegments
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
		n.availableSegments = removeByValue(n.availableSegments, j).([]interface{})
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
	args["max"] = n.NumSegments
	args["unit"] = n.Inputs[n.PrimaryUnit]
	s := &randomInteger{}
	return s.execute(args, expt).(uint64)
}
