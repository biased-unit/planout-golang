package goplanout

import ()

type SimpleNamespace struct {
	name                string
	primary_unit        string
	num_segments        int
	segment_allocations []string
	available_segments  []interface{}
	current_experiments map[string]bool
}

func (n *SimpleNamespace) addExperiment(name string, code map[string]interface{}, segments int) {
	avail := len(n.available_segments)
	if avail < segments {
		// error
	}

	// sample available_segments based on primary_unit
	params := make(map[string]interface{})
	params["unit"] = name

	expt := &Interpreter{
		Experiment_salt: n.name,
		Evaluated:       false,
		Inputs:          params,
		Outputs:         map[string]interface{}{},
		Overrides:       map[string]interface{}{},
	}

	m := make(map[string]interface{})
	m["choices"] = n.available_segments
	m["unit"] = name
	m["salt"] = n.name
	m["draws"] = segments
	s := &sample{}
	shuffle := s.execute(m, expt).([]interface{})

	// allocate sampled_segments to experiment
	// remove segment from available_segments
	for i := range shuffle {
		j := shuffle[i].(int)
		n.segment_allocations[j] = name
		n.available_segments = removeByValue(n.available_segments, j).([]interface{})
	}

	// update current_experiments
	n.current_experiments[name] = true
}

func (n *SimpleNamespace) removeExperiment(name string) {
	segments_to_free := make([]int, 0, n.num_segments)
	for i := range n.segment_allocations {
		if n.segment_allocations[i] == name {
			segments_to_free = append(segments_to_free, i)
		}
	}

	for i := range segments_to_free {
		n.available_segments = append(n.available_segments, segments_to_free[i])
	}

	delete(n.current_experiments, name)
}

func (n *SimpleNamespace) getSegment() int {
	// generate random integer min=0, max=num_segments, unit=primary_unit
	return 0
}
