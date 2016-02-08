package randomization

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
	//	"strconv"
	//	"strings"
)

// Do_assignment
func Do_assignment(M *map[string]string, project *Project, subject_id string,
	user_id string) (string, string) {

	// Set the seed to a random time.  Not sure if this is needed,
	// but since each assignment runs as a new instance we might
	// be getting the same "random numbers" every time if we don't
	// do this.
	source := rand.NewSource(time.Now().UnixNano())
	rgen := rand.New(source)

	numgroups := len(project.Group_names)
	numvar := len(project.Variables)
	rates := project.Sampling_rates

	data := Decode_data(project.Data)

	msg := ""

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
	N := len(project.Group_names)
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
	ur = ur * ur //!!!! Temporary, to improve balance
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

	msg += fmt.Sprintf("potential_scores=%v\n", potential_scores)
	msg += fmt.Sprintf("sorted_scores=%v\n", sorted_scores)
	msg += fmt.Sprintf("prob=%v\n", prob)
	msg += fmt.Sprintf("cumprob=%v\n", cumprob)
	msg += fmt.Sprintf("ur=%f\n", ur)
	msg += fmt.Sprintf("jr=%d\n", jr)
	msg += fmt.Sprintf("ties=%v\n", ties)
	msg += fmt.Sprintf("ii=%v\n", ii)

	// Update the project.
	project.Assignments[ii] += 1
	for j := 0; j < numvar; j++ {

		VA := project.Variables[j]
		x := (*M)[VA.Name]
		numlevels := len(VA.Levels)

		kk := -1
		for k, v := range VA.Levels {
			if x == v {
				kk = k
				break
			}
		}
		if kk == -1 {
			fmt.Printf("\n\n???????????\n\n")
		}
		data[j][ii*numlevels+kk] += 1
	}

	// Update the stored data
	if project.Store_RawData {
		rec := new(DataRecord)
		rec.Subject_id = subject_id
		rec.Assigned_time = time.Now()
		rec.Assigned_group = project.Group_names[ii]
		rec.Current_group = project.Group_names[ii]
		rec.Included = true
		data := make([]string, len(project.Variables))
		for j, v := range project.Variables {
			data[j] = (*M)[v.Name]
		}
		rec.Data = data
		rec.Assigner = user_id
		project.RawData = append(project.RawData, rec)
	}

	project.Num_assignments += 1

	project.Data = Encode_data(data)

	return project.Group_names[ii], msg
}

// Range returns the numerical range of the values in M.
func Range(M []float64) float64 {

	mn := M[0]
	mx := M[0]
	for _, x := range M {
		if x < mn {
			mn = x
		}
		if x > mx {
			mx = x
		}
	}
	return mx - mn
}

// StDev returns the standard deviation of the values in M.
func StDev(M []float64) float64 {

	m := 0.0
	for _, x := range M {
		m += x
	}
	m /= float64(len(M))

	v := 0.0
	for _, x := range M {
		d := x - m
		v += d * d
	}
	v /= float64(len(M))

	return math.Sqrt(v)
}

// Score calculates the contribution to the overall score if we put a
// subject with data value `x` into group `grp` for a given variable
// `va`.  `counts` contains the current cell counts for each level x group
// combination for this variable.
func Score(x string,
	grp int,
	counts []float64,
	rates []float64,
	va *Variable) float64 {

	nlevel := len(va.Levels)
	num_groups := len(counts) / nlevel

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
				new_counts[i] = counts[i*nlevel+j]
			} else {
				new_counts[i] = counts[i*nlevel+j] + 1
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
			fmt.Printf("Error: Unknown scoring function\n")
		}
	}

	return score_change
}
