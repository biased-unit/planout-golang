package goplanout

import (
	"fmt"
)

type SimpleNamespace struct {
	Name               string
	PrimaryUnit        string
	NumSegments        int
	SegmentAllocations []string
	AvailableSegments  []interface{}
	CurrentExperiments map[string]PlanOutCode
	Inputs             map[string]interface{}
}

func (n *SimpleNamespace) addExperiment(name string, code map[string]interface{}, segments int) error {
	avail := len(n.AvailableSegments)
	if avail < segments {
		return fmt.Errorf("Not enough segments available %v to add the new experiment %v\n", avail, name)
	}

	// Sample(choices=available_segments, draws=segments, unit=name)
	expt := &Interpreter{
		ExperimentSalt: n.Name,
		Evaluated:      false,
		Inputs:         n.Inputs,
		Outputs:        map[string]interface{}{},
		Overrides:      map[string]interface{}{},
	}

	// Compile Sample operator
	m := make(map[string]interface{})
	m["choices"] = n.AvailableSegments
	m["unit"] = name
	m["salt"] = n.Name
	m["draws"] = segments
	s := &sample{}
	shuffle := s.execute(m, expt).([]interface{})

	// Allocate sampled_segments to experiment
	// Remove segment from available_segments
	for i := range shuffle {
		j := shuffle[i].(int)
		n.SegmentAllocations[j] = name
		n.AvailableSegments = removeByValue(n.AvailableSegments, j).([]interface{})
	}

	// Update current_experiments
	n.CurrentExperiments[name] = code

	return nil
}

func (n *SimpleNamespace) removeExperiment(name string) error {
	_, exists := n.CurrentExperiments[name]
	if !exists {
		return fmt.Errorf("Experiment %v does not exists in the namespace\n", name)
	}

	segmentsToFree := make([]int, 0, n.NumSegments)
	for i := range n.SegmentAllocations {
		if n.SegmentAllocations[i] == name {
			segmentsToFree = append(segmentsToFree, i)
		}
	}

	for i := range segmentsToFree {
		n.AvailableSegments = append(n.AvailableSegments, segmentsToFree[i])
	}

	delete(n.CurrentExperiments, name)

	return nil
}

func (n *SimpleNamespace) getSegment() uint64 {
	// generate random integer min=0, max=num_segments, unit=primary_unit
	// RandomInteger(min=0, max=self.num_segments, unit=itemgetter(*self.primary_unit)(self.inputs))
	expt := &Interpreter{
		ExperimentSalt: n.Name,
		Evaluated:      false,
		Inputs:         n.Inputs,
		Outputs:        map[string]interface{}{},
		Overrides:      map[string]interface{}{},
	}

	// Compile RandomInteger operator
	m := make(map[string]interface{})
	m["salt"] = n.Name
	m["min"] = 0
	m["max"] = n.NumSegments
	m["unit"] = n.Inputs[n.PrimaryUnit]
	s := &randomInteger{}
	randInt := s.execute(m, expt).(uint64)
	return randInt
}
