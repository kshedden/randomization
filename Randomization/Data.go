package randomization

import (
	"fmt"
	//	"strconv"
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DataRecord struct {
	Subject_id     string
	Assigned_time  time.Time
	Assigned_group string
	Current_group  string
	Included       bool
	Data           []string
	Assigner       string
}

type Project struct {
	// The Google id of the project owner.
	Owner string

	// The date and time that the project was created.
	Created time.Time

	// The name of the project.
	Name string

	// The names of the groups
	Group_names []string
	Variables   []Variable
	Assignments []int
	Data        string

	// Controls the level of determinism in the group assignments
	Bias     int
	Comments []*Comment

	// The date and time of the last assignment
	Modified time.Time

	// If true, store the individual-level data, otherwise only store aggregates
	Store_RawData bool

	// The individual-level data, if stored
	RawData []*DataRecord

	// The number of subjects who have had assignments made
	Num_assignments int

	Removed_subjects []string

	// If true, the project is currently open for enrollment
	Open bool

	// The sampling rates for each treatment group.
	Sampling_rates []float64
}

// Encoded_Project is a version of Project that can be stored in the
// datastore.
type Encoded_Project struct {
	Owner            string
	Created          time.Time
	Name             string
	Group_names      []byte
	Variables        []byte
	Assignments      []int
	Data             string
	Bias             int
	Comments         []byte
	Modified         time.Time
	Store_RawData    bool
	RawData          []byte
	Num_assignments  int
	Removed_subjects []string
	Open             bool
	Sampling_rates   []float64
}

type Encoded_Project_view struct {
	Owner            string
	Created_date     string
	Created_time     string
	Name             string
	Key              string
	Group_names      []string
	Index            int
	Assignments      []int
	Data             string
	Bias             int
	Comments         []*Comment
	Modified_date    string
	Modified_time    string
	Modified_ever    bool
	Store_RawData    bool
	RawData          []byte
	Num_assignments  int
	Removed_subjects []string
	Open             bool
	Sampling_rates   string
}

// Project_view is a printable version of Project.
type Project_view struct {
	Owner            string
	Created_date     string
	Created_time     string
	Group_names      string
	Name             string
	Variables        []Variable_view
	Key              string
	Assignments      []int
	Data             string
	Bias             string
	Comments         []*Comment
	Modified_date    string
	Modified_time    string
	Store_RawData    bool
	RawData          []byte
	Num_assignments  int
	Removed_subjects []string
	Open             bool
	Sampling_rates   string
}

// Variable contains information about one variable that will be used
// as part of the treatment assignment.
type Variable struct {
	Name   string
	Levels []string
	Weight float64
	Func   string
}

// Variable_view is a printable version of Variable.
type Variable_view struct {
	Name   string
	Levels string
	Index  int
	Weight string
	Func   string
}

// Sharing_by_user is a record of all projects to which the given user
// has access.
type Sharing_by_user struct {
	User     string
	Projects string // Comma separated list of projects
}

// Sharing_by_project is a record of all users who have access to the
// given project.
type Sharing_by_project struct {
	Project_name string
	Users        string
}

// Comment
type Comment struct {
	Person   string
	DateTime time.Time
	Date     string
	Time     string
	Comment  []string
}

// Clean_split splits a string into tokens delimited by a given
// separator.  If S equals the empty string, this function returns an
// empty list, rather than a list containing an empty string as its
// sole element.  Leading and trailing whitespace is removed from each
// element of the returned list.
func Clean_split(S string, sep string) []string {

	if len(S) == 0 {
		return []string{}
	}

	A := strings.Split(S, sep)

	for i, v := range A {
		A[i] = strings.TrimSpace(v)
	}

	return A
}

// Copy_encoded_project creates a copy of a given project and returns
// it.
func Copy_encoded_project(proj *Encoded_Project) *Encoded_Project {

	newproj := new(Encoded_Project)

	newproj.Owner = proj.Owner
	newproj.Name = proj.Name
	newproj.Created = time.Now()

	newproj.Group_names = make([]byte, len(proj.Group_names))
	copy(newproj.Group_names, proj.Group_names)

	newproj.Variables = make([]byte, len(proj.Variables))
	copy(newproj.Variables, proj.Variables)

	newproj.Assignments = make([]int, len(proj.Assignments))
	copy(newproj.Assignments, proj.Assignments)

	newproj.Data = proj.Data
	newproj.Bias = proj.Bias

	newproj.Comments = make([]byte, len(proj.Comments))
	copy(newproj.Comments, proj.Comments)

	newproj.Modified = proj.Modified
	newproj.Store_RawData = proj.Store_RawData

	newproj.RawData = make([]byte, len(proj.RawData))
	copy(newproj.RawData, proj.RawData)

	newproj.Num_assignments = proj.Num_assignments

	newproj.Removed_subjects = make([]string, len(proj.Removed_subjects))
	copy(newproj.Removed_subjects, proj.Removed_subjects)

	newproj.Open = proj.Open

	newproj.Sampling_rates = make([]float64, len(proj.Sampling_rates))
	copy(newproj.Sampling_rates, proj.Sampling_rates)

	return (newproj)
}

// Decode_data
func Decode_data(D string) [][]float64 {

	DL := strings.Split(D, ";")
	Z := make([][]float64, len(DL))

	for i, V := range DL {
		U := strings.Split(V, ",")
		W := make([]float64, len(U))
		for j, x := range U {
			W[j], _ = strconv.ParseFloat(x, 64)
		}
		Z[i] = W
	}

	return Z
}

// Encode_data represents the treatment assignment frequencies as a
// string.  The variables are separated by ";", and the treatmtent
// group frequencies within a level are separated by ",".
func Encode_data(Z [][]float64) string {

	D := make([]string, len(Z))

	for i, V := range Z {
		U := make([]string, len(V))
		for j, x := range V {
			U[j] = fmt.Sprintf("%.4f", x)
		}
		D[i] = strings.Join(U, ",")
	}

	return strings.Join(D, ";")
}

// Get_project_from_key
func Get_project_from_key(pkey string,
	c *appengine.Context) (*Project, error) {

	Key := datastore.NewKey(*c, "Encoded_Project", pkey, 0, nil)
	var EP Encoded_Project
	err := datastore.Get(*c, Key, &EP)
	if err != nil {
		(*c).Errorf("Project_dashboard: %v", err)
		return nil, err
	}

	project := Decode_Project(&EP)

	// This field was added later, so may be missing on some
	// projects.  Provide a default here.
	if project.Sampling_rates == nil {
		arr := make([]float64, len(project.Group_names))
		for i, _ := range arr {
			arr[i] = 1.0
		}
		project.Sampling_rates = arr
	}

	return project, nil
}

// Encode_Project takes a Project struct and converts it into a form
// that can be stored in the datastore.
func Encode_Project(P *Project) (*Encoded_Project, error) {

	EP := new(Encoded_Project)

	EP.Owner = P.Owner
	EP.Created = P.Created
	EP.Name = P.Name
	EP.Data = P.Data
	EP.Assignments = P.Assignments
	EP.Bias = P.Bias
	EP.Modified = P.Modified
	EP.Store_RawData = P.Store_RawData
	EP.Num_assignments = P.Num_assignments
	EP.Removed_subjects = P.Removed_subjects
	EP.Open = P.Open
	EP.Sampling_rates = P.Sampling_rates

	// Group names
	x1, err := json.Marshal(P.Group_names)
	if err != nil {
		return nil, err
	}
	EP.Group_names = x1

	// Variables
	x2, err := json.Marshal(P.Variables)
	if err != nil {
		return nil, err
	}
	EP.Variables = x2

	// Raw data
	if P.Store_RawData {
		x3, err := json.Marshal(P.RawData)
		if err != nil {
			return nil, err
		}
		EP.RawData = x3
	}

	// Comments
	x4, err := json.Marshal(P.Comments)
	if err != nil {
		return nil, err
	}
	EP.Comments = x4

	return EP, nil
}

// Store_project
func Store_project(Project *Project,
	project_key string,
	c *appengine.Context) error {

	EP, err := Encode_Project(Project)
	if err != nil {
		return err
	}

	Key := datastore.NewKey(*c, "Encoded_Project", project_key, 0, nil)
	_, err = datastore.Put(*c, Key, EP)
	return err
}

// Decode_Project takes a project in its encoded form (storable in the
// datastore) and converts it to a Project struct.
func Decode_Project(EP *Encoded_Project) *Project {

	P := new(Project)

	P.Owner = EP.Owner
	P.Created = EP.Created
	P.Name = EP.Name
	P.Data = EP.Data
	P.Assignments = EP.Assignments
	P.Bias = EP.Bias
	P.Store_RawData = EP.Store_RawData
	P.Num_assignments = EP.Num_assignments
	P.Removed_subjects = EP.Removed_subjects
	P.Modified = EP.Modified
	P.Open = EP.Open
	P.Sampling_rates = EP.Sampling_rates

	var S []string
	json.Unmarshal(EP.Group_names, &S)
	P.Group_names = S

	var V []Variable
	json.Unmarshal(EP.Variables, &V)
	P.Variables = V

	var C []*Comment
	json.Unmarshal(EP.Comments, &C)
	P.Comments = C

	if EP.Store_RawData {
		var rawdata []*DataRecord
		json.Unmarshal(EP.RawData, &rawdata)
		P.RawData = rawdata
	}

	return P
}

// Format_project returns a Project_view object corresponding to the
// given Project and Key object.
func Format_project(project *Project) *Project_view {

	B := new(Project_view)
	B.Owner = project.Owner
	B.Data = project.Data
	B.Name = project.Name
	B.Comments = project.Comments
	B.Assignments = project.Assignments
	B.Bias = fmt.Sprintf("%d", project.Bias)
	t := project.Created
	loc, _ := time.LoadLocation("America/New_York")
	t = t.In(loc)
	B.Created_date = t.Format("2006-1-2")
	B.Created_time = t.Format("3:04pm")
	B.Group_names = strings.Join(project.Group_names, ",")
	B.Variables = make([]Variable_view, len(project.Variables))

	rate_str := make([]string, len(project.Sampling_rates))
	for i, x := range project.Sampling_rates {
		rate_str[i] = fmt.Sprintf("%.0f", x)
	}
	B.Sampling_rates = strings.Join(rate_str, ",")

	for i, pv := range project.Variables {
		B.Variables[i] = Format_Variable(pv)
	}
	B.Removed_subjects = project.Removed_subjects
	B.Open = project.Open

	t = project.Modified
	loc, _ = time.LoadLocation("America/New_York")
	t = t.In(loc)
	B.Modified_date = t.Format("2006-1-2")
	B.Modified_time = t.Format("3:04pm")

	return B
}

// Format_Encoded_Project returns an Encoded_Project_view object
// corresponding to the given Encoded_project object.
func Format_encoded_project(enc_project *Encoded_Project) *Encoded_Project_view {

	view := new(Encoded_Project_view)
	view.Owner = enc_project.Owner
	view.Name = enc_project.Name
	view.Data = enc_project.Data
	view.Assignments = enc_project.Assignments
	view.Bias = enc_project.Bias
	view.Key = enc_project.Owner + "::" + enc_project.Name
	view.Num_assignments = enc_project.Num_assignments

	var s []string
	json.Unmarshal(enc_project.Group_names, &s)
	view.Group_names = s

	view.Removed_subjects = enc_project.Removed_subjects
	//view.Comments = enc_project.Comments
	view.Open = enc_project.Open

	rate_str := make([]string, len(enc_project.Sampling_rates))
	for i, x := range enc_project.Sampling_rates {
		rate_str[i] = fmt.Sprintf("%.0f", x)
	}
	view.Sampling_rates = strings.Join(rate_str, ",")

	// Created date
	t := enc_project.Created
	loc, _ := time.LoadLocation("America/New_York")
	t = t.In(loc)
	view.Created_date = t.Format("2006-1-2")
	view.Created_time = t.Format("3:04pm")

	// Modified date
	t = enc_project.Modified

	if t.IsZero() {
		view.Modified_ever = false
	} else {
		view.Modified_ever = true
		t = t.In(loc)
		view.Modified_date = t.Format("2006-1-2")
		view.Modified_time = t.Format("3:04pm")
	}

	return view
}

// Format_Encoded_Project returns an array of Encoded_Project_view
// objects corresponding to the given array of Encoded_Project
// objects.
func Format_encoded_projects(P []*Encoded_Project) []*Encoded_Project_view {

	n := len(P)
	B := make([]*Encoded_Project_view, n, n)
	for i := 0; i < n; i++ {
		B[i] = Format_encoded_project(P[i])
	}

	return B
}

// Format_projects
func Format_projects(projects []*Project) []*Project_view {

	n := len(projects)
	fmt_projects := make([]*Project_view, n, n)
	for i := 0; i < n; i++ {
		fmt_projects[i] = Format_project(projects[i])
	}

	return fmt_projects
}

// Format_variables returns an array of Variable_view objects
// corresponding to the given array of Variable objects.
func Format_variables(VA []Variable) []Variable_view {

	VAF := make([]Variable_view, len(VA))

	for i, va := range VA {
		VAF[i] = Format_Variable(va)
		VAF[i].Index = i
	}

	return VAF
}

// Format_variable returns a Variable_view object corresponding to the
// given Variable object.
func Format_Variable(VA Variable) Variable_view {

	var VV Variable_view
	VV.Name = VA.Name
	VV.Levels = strings.Join(VA.Levels, ",")
	VV.Weight = fmt.Sprintf("%.0f", VA.Weight)
	VV.Func = VA.Func

	return VV
}

// Get_shared_users returns a list of user id's for for users who are
// shared for the given project.
func Get_shared_users(project_name string,
	c *appengine.Context) ([]string, error) {

	Key := datastore.NewKey(*c, "Sharing_by_project",
		project_name, 0, nil)
	var SP Sharing_by_project
	err := datastore.Get(*c, Key, &SP)
	if err == datastore.ErrNoSuchEntity {
		return []string{}, nil
	} else if err != nil {
		return []string{}, err
	}
	if len(SP.Users) == 0 {
		return []string{}, nil
	}
	A := Clean_split(SP.Users, ",")
	return A, nil
}

// Add_sharing adds all the given users to be shared for the given
// project.
func Add_sharing(project_name string,
	user_names []string,
	c *appengine.Context) error {

	if len(user_names) == 0 {
		return nil
	}

	// Update Sharing_by_project.
	Key := datastore.NewKey(*c, "Sharing_by_project", project_name, 0, nil)
	SP := new(Sharing_by_project)
	err := datastore.Get(*c, Key, SP)
	if err == datastore.ErrNoSuchEntity {
		SP = new(Sharing_by_project)
		SP.Project_name = project_name
		SP.Users = strings.Join(user_names, ",")
	} else if err != nil {
		return err
	} else {
		U := Clean_split(SP.Users, ",")
		m := make(map[string]bool)
		for _, u := range U {
			m[u] = true
		}
		for _, u := range user_names {
			m[u] = true
		}
		A := make([]string, len(m))
		i := 0
		for k, _ := range m {
			A[i] = k
			i++
		}
		SP.Users = strings.Join(A, ",")
	}

	_, err = datastore.Put(*c, Key, SP)
	if err != nil {
		return err
	}

	// Update Sharing_by_user.
	for _, user_name := range user_names {
		Key = datastore.NewKey(*c, "Sharing_by_user",
			strings.ToLower(user_name), 0, nil)
		SU := new(Sharing_by_user)
		err := datastore.Get(*c, Key, SU)
		if err == datastore.ErrNoSuchEntity {
			SU = new(Sharing_by_user)
			SU.User = user_name
			SU.Projects = project_name
		} else if err != nil {
			return err
		} else {
			U := Clean_split(SU.Projects, ",")
			m := make(map[string]bool)
			for _, u := range U {
				m[u] = true
			}
			m[project_name] = true
			A := make([]string, len(m))
			i := 0
			for k, _ := range m {
				A[i] = k
				i++
			}
			SU.Projects = strings.Join(A, ",")
		}
		_, err = datastore.Put(*c, Key, SU)
		if err != nil {
			return err
		}
	}

	return nil
}

// Remove_sharing
func Remove_sharing(project_name string,
	user_names []string,
	c *appengine.Context) error {

	// Map whose keys are the users to remove.
	h := make(map[string]bool)
	for i := 0; i < len(user_names); i++ {
		h[user_names[i]] = true
	}

	// Update Sharing_by_project.
	Key := datastore.NewKey(*c, "Sharing_by_project", project_name, 0, nil)
	SP := new(Sharing_by_project)
	err := datastore.Get(*c, Key, SP)
	if err == datastore.ErrNoSuchEntity {
		// OK
	} else if err != nil {
		return err
	} else {
		U := Clean_split(SP.Users, ",")
		m := make(map[string]bool)
		for _, u := range U {
			if _, ok := h[u]; !ok {
				m[u] = true
			}
		}
		A := make([]string, len(m))
		i := 0
		for k, _ := range m {
			A[i] = k
			i++
		}
		SP.Users = strings.Join(A, ",")

		_, err = datastore.Put(*c, Key, SP)
		if err != nil {
			return err
		}
	}

	// Update Sharing_by_user.
	for i := 0; i < len(user_names); i++ {
		Key := datastore.NewKey(*c, "Sharing_by_user",
			strings.ToLower(user_names[i]), 0, nil)
		SU := new(Sharing_by_user)
		err := datastore.Get(*c, Key, SU)
		if err == datastore.ErrNoSuchEntity {
			// OK
		} else if err != nil {
			return err
		} else {
			U := Clean_split(SU.Projects, ",")
			m := make(map[string]bool)
			for _, u := range U {
				if u != project_name {
					m[u] = true
				}
			}
			A := make([]string, len(m))
			i := 0
			for k, _ := range m {
				A[i] = k
				i++
			}
			SU.Projects = strings.Join(A, ",")

			_, err = datastore.Put(*c, Key, SU)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Get_projects returns all projects owned by the given user.
func Get_projects(user string,
	include_shared bool,
	c *appengine.Context) ([]*datastore.Key, []*Encoded_Project, error) {

	q := datastore.NewQuery("Encoded_Project").
		Filter("Owner = ", user).
		Order("-Created").Limit(100)

	var PR []*Encoded_Project
	KY, err := q.GetAll(*c, &PR)
	if err != nil {
		return nil, nil, err
	}

	if !include_shared {
		return KY, PR, err
	}

	Key := datastore.NewKey(*c, "Sharing_by_user", strings.ToLower(user),
		0, nil)
	(*c).Infof("User=%v", user)
	var SP Sharing_by_user
	err = datastore.Get(*c, Key, &SP)
	if err == datastore.ErrNoSuchEntity {
		return KY, PR, nil
	}
	if err != nil {
		return nil, nil, err
	}
	SPV := Clean_split(SP.Projects, ",")

	SEP := make([]*Encoded_Project, len(SPV))
	SKY := make([]*datastore.Key, len(SPV))
	for i, spv := range SPV {
		Key = datastore.NewKey(*c, "Encoded_Project", spv, 0, nil)
		P := new(Encoded_Project)
		err = datastore.Get(*c, Key, P)
		SEP[i] = P
		SKY[i] = Key
	}

	KY1 := make([]*datastore.Key, len(KY)+len(SKY))
	copy(KY1, KY)
	copy(KY1[len(KY):], SKY)
	PR1 := make([]*Encoded_Project, len(PR)+len(SEP))
	copy(PR1, PR)
	copy(PR1[len(PR):], SEP)
	return KY1, PR1, nil
}

// Used when the GET/POST method is mismatched to the handler.
func Serve404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "Not Found")
}

func ServeError(c *appengine.Context,
	w http.ResponseWriter,
	err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "Internal Server Error")
	(*c).Errorf("%v", err)
	fmt.Printf("\n%v\n", err)
}

// Message_page presents a simple message page and presents the user
// with a link that leads to a followup page.
func Message_page(w http.ResponseWriter,
	r *http.Request,
	login_user *user.User,
	Msg string,
	Return_msg string,
	Return_url string) {

	c := appengine.NewContext(r)

	tmpl, err := template.ParseFiles("header.html",
		"message.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	type TV struct {
		User       string
		Logged_in  bool
		Msg        string
		Return_url string
		Return_msg string
		Logout_url string
	}

	template_values := new(TV)
	if login_user != nil {
		template_values.User = login_user.String()
	} else {
		template_values.User = ""
	}
	template_values.Logged_in = login_user != nil
	template_values.Msg = Msg
	template_values.Return_url = Return_url
	template_values.Return_msg = Return_msg

	if err := tmpl.ExecuteTemplate(w, "message.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

func get_index(vec []string, val string) int {

	for i, x := range vec {
		if val == x {
			return i
		}
	}
	return -1
}

// Remove_from_aggregate
func Remove_from_aggregate(rec *DataRecord,
	proj *Project) {

	grp_ix := get_index(proj.Group_names, rec.Current_group)

	// Update the overall assignment totals
	proj.Assignments[grp_ix] -= 1

	// Update the within-variable assignment totals
	D := Decode_data(proj.Data)
	for j, va := range proj.Variables {
		numlevels := len(va.Levels)
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				D[j][grp_ix*numlevels+k] -= 1
			}
		}
	}
	proj.Data = Encode_data(D)
}

// Add_to_aggregate
func Add_to_aggregate(rec *DataRecord,
	proj *Project) {

	grp_ix := get_index(proj.Group_names, rec.Current_group)

	// Update the overall assignment totals
	proj.Assignments[grp_ix] += 1

	// Update the within-variable assignment totals
	D := Decode_data(proj.Data)
	for j, va := range proj.Variables {
		numlevels := len(va.Levels)
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				D[j][grp_ix*numlevels+k] += 1
			}
		}
	}
	proj.Data = Encode_data(D)
}
