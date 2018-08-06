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
	SegmentAllocations map[uint64]string
	AvailableSegments  []int
	CurrentExperiments map[string]*Interpreter
	DefaultExperiment  *Interpreter
	SelectedExperiment uint64
}

func NewSimpleNamespace(name string, numSegments int, primaryUnit string, inputs map[string]interface{}) SimpleNamespace {
	avail := make([]int, 0, numSegments)
	for i := 0; i < numSegments; i++ {
		avail = append(avail, i)
	}

	noop := &Interpreter{
		Name:   name,
		Salt:   name,
		Inputs: inputs,
		Code:   make(map[string]interface{}),
	}

	return SimpleNamespace{
		Name:               name,
		PrimaryUnit:        primaryUnit,
		NumSegments:        numSegments,
		Inputs:             inputs,
		SegmentAllocations: make(map[uint64]string),
		AvailableSegments:  avail,
		CurrentExperiments: make(map[string]*Interpreter),
		SelectedExperiment: uint64(numSegments + 1),
		DefaultExperiment:  noop,
	}
}

func (n *SimpleNamespace) SetInputs(inputs map[string]interface{}) {
	for _, exp := range n.CurrentExperiments {
		exp.Inputs = inputs
	}
	n.DefaultExperiment.Inputs = inputs
	n.Inputs = inputs
}

func (n *SimpleNamespace) ReSet() {
	for _, exp := range n.CurrentExperiments {
		exp.ReSet()
	}
	n.DefaultExperiment.ReSet()
	n.Inputs = make(map[string]interface{})
	n.SelectedExperiment = uint64(n.NumSegments + 1)
}

func (n *SimpleNamespace) Run() *Interpreter {
	interpreter := n.DefaultExperiment

	if name, ok := n.SegmentAllocations[n.getSegment()]; ok {
		interpreter = n.CurrentExperiments[name]
		//interpreter.Name = n.Name + "-" + interpreter.Name
		//interpreter.Salt = n.Name + "." + interpreter.Name
	}

	interpreter.Run()
	return interpreter
}

func (n *SimpleNamespace) AddDefaultExperiment(defaultExperiment *Interpreter) {
	n.DefaultExperiment = defaultExperiment
}

func (n *SimpleNamespace) AddExperiment(name string, interpreter *Interpreter, segments int) error {
	avail := len(n.AvailableSegments)
	if avail < segments {
		return fmt.Errorf("Not enough segments available %v to add the new experiment %v\n", avail, name)
	}

	if _, ok := n.CurrentExperiments[name]; ok {
		return fmt.Errorf("There is already and experiment called %s\n", name)
	}

	n.allocateExperiment(name, segments)

	n.CurrentExperiments[name] = interpreter
	return nil
}

func (n *SimpleNamespace) RemoveExperiment(name string) error {
	_, exists := n.CurrentExperiments[name]
	if !exists {
		return fmt.Errorf("Experiment %v does not exists in the namespace\n", name)
	}

	segmentsToFree := make([]int, 0, n.NumSegments)
	for i := range n.SegmentAllocations {
		if n.SegmentAllocations[i] == name {
			segmentsToFree = append(segmentsToFree, int(i))
			delete(n.SegmentAllocations, i)
		}
	}

	for i := range segmentsToFree {
		n.AvailableSegments = append(n.AvailableSegments, segmentsToFree[i])
	}

	sort.Ints(n.AvailableSegments)

	delete(n.CurrentExperiments, name)
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
	var availableSegmentsAsInterface = make([]interface{}, len(n.AvailableSegments))
	for i, d := range n.AvailableSegments {
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
		n.SegmentAllocations[uint64(j)] = name
		n.AvailableSegments = deallocateSegments(n.AvailableSegments, j)
	}
}

func (n *SimpleNamespace) getSegment() uint64 {

	if n.SelectedExperiment != uint64(n.NumSegments+1) {
		return n.SelectedExperiment
	}

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
	n.SelectedExperiment = s.execute(args, expt).(uint64)
	return n.SelectedExperiment
}

func deallocateSegments(allocated []int, segmentToRemove int) []int {
	i := 0
	n := len(allocated)
	for i < n && allocated[i] != segmentToRemove {
		i = i + 1
	}
	if i < n {
		allocated = append(allocated[:i], allocated[i+1:]...)
	}
	return allocated
}
