package randomization

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// Do_assignment
func Do_assignment(M *map[string]string, project *Project, subject_id string,
	user_id string) (string, error) {

	// Set the seed to a random time.  Not sure if this is needed,
	// but since each assignment runs as a new instance we might
	// be getting the same "random numbers" every time if we don't
	// do this.
	source := rand.NewSource(time.Now().UnixNano())
	rgen := rand.New(source)

	numgroups := len(project.GroupNames)
	numvar := len(project.Variables)
	rates := project.SamplingRates

	data := project.Data

	// Calculate the scores if assigning the new subject
	// to each possible group.
	potential_scores := make([]float64, numgroups)
	for i := 0; i < numgroups; i++ {

		// The score is a weighted linear combination over the
		// variables.
		potential_scores[i] = 0
		for j, va := range project.Variables {
			x := (*M)[va.Name]
			score := Score(x, i, data[j], rates, &va)
			potential_scores[i] += va.Weight * score
		}
	}

	// Get a sorted copy of the scores.
	sorted_scores := make([]float64, len(potential_scores))
	copy(sorted_scores, potential_scores)
	sort.Float64s(sorted_scores)

	// Construct the Pocock/Simon probabilities.
	N := len(project.GroupNames)
	qmin := 1 / float64(N)
	qmax := 2 / float64(N-1)
	qq := qmin + float64(project.Bias-1)*(qmax-qmin)/9.0
	prob := make([]float64, N)
	for j, _ := range prob {
		prob[j] = qq - 2*(float64(N)*qq-1)*float64(j+1)/float64(N*(N+1))
	}

	// The cumulative Pocock Simon probabilities.
	cumprob := make([]float64, N)
	copy(cumprob, prob)
	for j := 1; j < len(cumprob); j++ {
		cumprob[j] += cumprob[j-1]
	}

	// A random value distributed according to the Pocock Simon
	// probabilities.
	ur := rgen.Float64()
	jr := 0
	for ii, x := range cumprob {
		if x > ur {
			jr = ii
			break
		}
	}

	// Get all values tied with the selected value.
	ties := make([]int, 0, len(potential_scores))
	for ii, x := range potential_scores {
		if x == sorted_scores[jr] {
			ties = append(ties, ii)
		}
	}

	// Assign to this group.
	ii := ties[rgen.Intn(len(ties))]

	// Update the project.
	project.Assignments[ii] += 1
	for j := 0; j < numvar; j++ {

		VA := project.Variables[j]
		x := (*M)[VA.Name]

		kk := -1
		for k, v := range VA.Levels {
			if x == v {
				kk = k
				break
			}
		}
		if kk == -1 {
			return "", fmt.Errorf("Invalid state in Do_assignment")
		}
		data[j][kk][ii] += 1
	}

	// Update the stored data
	if project.StoreRawData {
		rec := new(DataRecord)
		rec.SubjectId = subject_id
		rec.AssignedTime = time.Now()
		rec.AssignedGroup = project.GroupNames[ii]
		rec.CurrentGroup = project.GroupNames[ii]
		rec.Included = true
		data := make([]string, len(project.Variables))
		for j, v := range project.Variables {
			data[j] = (*M)[v.Name]
		}
		rec.Data = data
		rec.Assigner = user_id
		project.RawData = append(project.RawData, rec)
	}

	project.NumAssignments += 1

	return project.GroupNames[ii], nil
}

// Range returns the numerical range of the values in vec.
func Range(vec []float64) float64 {

	mn := vec[0]
	mx := vec[0]
	for _, x := range vec {
		if x < mn {
			mn = x
		}
		if x > mx {
			mx = x
		}
	}
	return mx - mn
}

// StDev returns the standard deviation of the values in vec.
func StDev(vec []float64) float64 {

	m := 0.0
	for _, x := range vec {
		m += x
	}
	m /= float64(len(vec))

	v := 0.0
	for _, x := range vec {
		d := x - m
		v += d * d
	}
	v /= float64(len(vec))

	return math.Sqrt(v)
}

// Score calculates the contribution to the overall score if we put a
// subject with data value `x` into group `grp` for a given variable
// `va`.  `counts` contains the current cell counts for each level x group
// combination for this variable, `va` contains variable information.
func Score(x string,
	grp int,
	counts [][]float64,
	rates []float64,
	va *Variable) float64 {

	nlevel := len(va.Levels)
	num_groups := len(counts[0])

	new_counts := make([]float64, num_groups)
	score_change := 0.0
	for j := 0; j < nlevel; j++ {

		if x != va.Levels[j] {
			continue
		}

		// Get the count for each group if we were to assign
		// this unit to group `grp`.
		for i := 0; i < num_groups; i++ {
			if i != grp {
				new_counts[i] = counts[j][i]
			} else {
				new_counts[i] = counts[j][i] + 1
			}
		}

		// Adjust the counts to account for the intended
		// marginal frequencies.
		for i := 0; i < num_groups; i++ {
			new_counts[i] /= rates[i]
		}

		if va.Func == "Range" {
			score_change += Range(new_counts)
		} else if va.Func == "StDev" {
			score_change += StDev(new_counts)
		} else {
			panic("Error: Unknown scoring function\n")
		}
	}

	return score_change
}
