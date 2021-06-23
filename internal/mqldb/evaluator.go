package mqldb

import (
	"fmt"

	"github.com/materials-commons/gomcdb/mcmodel"
)

func EvalStatement(db *DB, selection Selection, statement Statement) ([]mcmodel.Activity, []mcmodel.Entity) {
	var (
		matchingProcesses []mcmodel.Activity
		matchingSamples   []mcmodel.Entity
	)
	switch {
	case selection.ProcessSelection.All && selection.SampleSelection.All:
		matchingProcesses, matchingSamples = evalSelectProcessesAndSamples(db, statement)
	case selection.SampleSelection.All:
		matchingSamples = evalSelectSamples(db, statement)
	case selection.ProcessSelection.All:
		matchingProcesses = evalSelectProcesses(db, statement)
	}

	return matchingProcesses, matchingSamples
}

func evalSelectProcessesAndSamples(db *DB, statement Statement) ([]mcmodel.Activity, []mcmodel.Entity) {
	processes := evalSelectProcesses(db, statement)
	samples := evalSelectSamples(db, statement)
	return processes, samples
}

func evalSelectSamples(db *DB, statement Statement) []mcmodel.Entity {
	var matchingSamples []mcmodel.Entity
	var matchingProcesses []mcmodel.Activity

	if hasSampleMatchStatement(statement) {
		matchingSamples = evalMatchingSamples(db, statement)
	}

	if hasProcessMatchStatement(statement) {
		matchingProcesses = evalMatchingProcesses(db, statement)
	}

	processSamples := uniqueSamplesForProcesses(db, matchingProcesses)

	// We now have two lists of samples - samples that matched from evalMatchingSamples, and samples that
	// were matched by finding the processes and then the samples for those processes. These two lists of
	// samples may share the same sample. Since we want to return just the unique list of matches we go
	// through each of these lists to identify the samples that are unique across the lists. We do this
	// by creating a hash table of sample ids. Then taking those unique samples and creating a list of
	// the unique samples to return.
	var samplesToReturn []mcmodel.Entity
	uniqueSamples := make(map[int]mcmodel.Entity)
	for _, sample := range matchingSamples {
		uniqueSamples[sample.ID] = sample
	}

	for _, sample := range processSamples {
		uniqueSamples[sample.ID] = sample
	}

	for _, sample := range uniqueSamples {
		samplesToReturn = append(samplesToReturn, sample)
	}

	return samplesToReturn
}

func uniqueSamplesForProcesses(db *DB, processes []mcmodel.Activity) []mcmodel.Entity {
	var samples []mcmodel.Entity
	uniqueSamples := make(map[int]*mcmodel.Entity)
	for _, process := range processes {
		for _, sample := range db.ProcessSamples[process.ID] {
			uniqueSamples[sample.ID] = sample
		}
	}

	for _, sample := range uniqueSamples {
		samples = append(samples, *sample)
	}

	return samples
}

func evalSelectProcesses(db *DB, statement Statement) []mcmodel.Activity {
	var matchingProcesses []mcmodel.Activity
	var matchingSamples []mcmodel.Entity

	if hasProcessMatchStatement(statement) {
		matchingProcesses = evalMatchingProcesses(db, statement)
	}

	if hasSampleMatchStatement(statement) {
		matchingSamples = evalMatchingSamples(db, statement)
	}

	sampleProcesses := uniqueProcessesForSamples(db, matchingSamples)

	// We now have two lists of processes - processes that matched from evalMatchingProcesses, and processes that
	// were matched by finding the samples and then the processes for those samples. These two lists of
	// processes may share the same process. Since we want to return just the unique list of matches we go
	// through each of these lists to identify the processes that are unique across the lists. We do this
	// by creating a hash table of process ids. Then taking those unique processes and creating a list of
	// the unique processes to return.
	var processesToReturn []mcmodel.Activity
	uniqueProcesses := make(map[int]mcmodel.Activity)
	for _, process := range matchingProcesses {
		uniqueProcesses[process.ID] = process
	}

	for _, process := range sampleProcesses {
		uniqueProcesses[process.ID] = process
	}

	for _, sample := range uniqueProcesses {
		processesToReturn = append(processesToReturn, sample)
	}

	return processesToReturn
}

func uniqueProcessesForSamples(db *DB, samples []mcmodel.Entity) []mcmodel.Activity {
	var processes []mcmodel.Activity
	uniqueProcesses := make(map[int]*mcmodel.Activity)
	for _, sample := range samples {
		for _, process := range db.SampleProcesses[sample.ID] {
			uniqueProcesses[process.ID] = process
		}
	}

	for _, process := range uniqueProcesses {
		processes = append(processes, *process)
	}

	return processes
}

func evalMatchingProcesses(db *DB, statement Statement) []mcmodel.Activity {
	var matchingProcesses []mcmodel.Activity
	uniqueProcessMatches := make(map[int]mcmodel.Activity)
	for _, process := range db.Processes {
		if eval(db, &process, nil, statement) {
			uniqueProcessMatches[process.ID] = process
		}
	}

	for _, process := range uniqueProcessMatches {
		matchingProcesses = append(matchingProcesses, process)
	}

	return matchingProcesses
}

type SampleState struct {
	sample        *mcmodel.Entity
	EntityStateID int
}

func evalMatchingSamples(db *DB, statement Statement) []mcmodel.Entity {
	var matchingSamples []mcmodel.Entity
	uniqueSampleMatches := make(map[int]mcmodel.Entity)
	for _, sample := range db.Samples {
		for _, entityState := range sample.EntityStates {
			sampleState := SampleState{&sample, entityState.ID}
			if eval(db, nil, &sampleState, statement) {
				// Found a match on the sample, no need to check other sample states so break out of the state loop
				uniqueSampleMatches[sample.ID] = sample
				break
			}
		}
	}

	for _, sample := range uniqueSampleMatches {
		matchingSamples = append(matchingSamples, sample)
	}

	return matchingSamples
}

func eval(db *DB, process *mcmodel.Activity, sampleState *SampleState, statement Statement) bool {
	switch s := statement.(type) {
	case MatchStatement:
		return evalMatchStatement(db, process, sampleState, s)
	case AndStatement:
		return evalAndStatement(db, process, sampleState, s)
	case OrStatement:
		return evalOrStatement(db, process, sampleState, s)
	default:
		return false
	}
}

func evalAndStatement(db *DB, process *mcmodel.Activity, sampleState *SampleState, statement AndStatement) bool {
	if !eval(db, process, sampleState, statement.Left) {
		return false
	}

	return eval(db, process, sampleState, statement.Right)
}

func evalOrStatement(db *DB, process *mcmodel.Activity, sampleState *SampleState, statement OrStatement) bool {
	if eval(db, process, sampleState, statement.Left) {
		return true
	}

	return eval(db, process, sampleState, statement.Right)
}

func evalMatchStatement(db *DB, process *mcmodel.Activity, sampleState *SampleState, match MatchStatement) bool {
	switch match.FieldType {
	case ProcessFieldType:
		return evalProcessFieldMatch(process, match)
	case ProcessAttributeFieldType:
		return evalProcessAttributeFieldMatch(process, db, match)
	case SampleFieldType:
		return evalSampleFieldMatch(sampleState, match)
	case SampleAttributeFieldType:
		return evalSampleAttributeFieldMatch(sampleState, db, match)
	case ProcessFuncType:
		return evalProcessFuncMatch(process, db, match)
	case SampleFuncType:
		return evalSampleFuncMatch(sampleState, db, match)
	}

	return false
}

func evalSampleFuncMatch(state *SampleState, db *DB, match MatchStatement) bool {
	switch {
	case match.Operation == "has-process":
		return evalSampleFuncMatchHasProcess(state, db, match.Value.(string))
	}
	return false
}

func evalSampleFuncMatchHasProcess(state *SampleState, db *DB, processType string) bool {
	processes, ok := db.SampleProcesses[state.sample.ID]
	if !ok {
		// weird... sample doesn't have any processes... (shouldn't happen)
		return false
	}

	for _, process := range processes {
		if process.Name == processType {
			return true
		}
	}

	return false
}

func evalProcessFuncMatch(process *mcmodel.Activity, db *DB, match MatchStatement) bool {
	return false
}

func evalProcessAttributeFieldMatch(process *mcmodel.Activity, db *DB, match MatchStatement) bool {
	if process == nil {
		return false
	}

	attributes, ok := db.ProcessAttributesByProcessID[process.ID]
	if !ok {
		fmt.Printf("    Process %d/%s has no attributes\n", process.ID, process.Name)
		return false
	}

	attribute, ok := attributes[match.FieldName]
	if !ok {
		return false
	}

	for _, value := range attribute.AttributeValues {
		switch value.ValueType {
		case mcmodel.ValueTypeInt:
			return tryEvalAttributeIntMatch(value.ValueInt, match)
		case mcmodel.ValueTypeFloat:
			return tryEvalAttributeFloatMatch(value.ValueFloat, match)
		case mcmodel.ValueTypeString:
			return tryEvalAttributeStringMatch(value.ValueString, match)
		}
	}

	return false
}

func evalSampleAttributeFieldMatch(sampleState *SampleState, db *DB, match MatchStatement) bool {
	if sampleState == nil {
		return false
	}
	states, ok := db.SampleAttributesBySampleIDAndStates[sampleState.sample.ID]
	if !ok {
		fmt.Printf("    Sample %d/%s has no states\n", sampleState.sample.ID, sampleState.sample.Name)
		return false
	}

	attributes, ok := states[sampleState.EntityStateID]
	if !ok {
		fmt.Printf("     Sample %d/%s with state %d has no attributes\n", sampleState.sample.ID, sampleState.sample.Name, sampleState.EntityStateID)
	}

	attribute, ok := attributes[match.FieldName]
	if !ok {
		return false
	}

	for _, value := range attribute.AttributeValues {
		switch value.ValueType {
		case mcmodel.ValueTypeInt:
			return tryEvalAttributeIntMatch(value.ValueInt, match)
		case mcmodel.ValueTypeFloat:
			return tryEvalAttributeFloatMatch(value.ValueFloat, match)
		case mcmodel.ValueTypeString:
			return tryEvalAttributeStringMatch(value.ValueString, match)
		}
	}

	return false
}
