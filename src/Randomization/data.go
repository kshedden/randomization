package randomization

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"
)

// DataRecord stores one record of raw data.
type DataRecord struct {
	SubjectId     string
	AssignedTime  time.Time
	AssignedGroup string
	CurrentGroup  string
	Included      bool
	Data          []string
	Assigner      string
}

// Project stores all information about one project.
type Project struct {
	// The Google id of the project owner.
	Owner string

	// The date and time that the project was created.
	Created time.Time

	// The name of the project.
	Name string

	// The names of the groups
	GroupNames  []string
	Variables   []Variable
	Assignments []int
	Data        [][][]float64

	// Controls the level of determinism in the group assignments
	Bias     int
	Comments []*Comment

	// The date and time of the last assignment
	Modified time.Time

	// If true, store the individual-level data, otherwise only store aggregates
	StoreRawData bool

	// The individual-level data, if stored
	RawData []*DataRecord

	// The number of subjects who have had assignments made
	NumAssignments int

	RemovedSubjects []string

	// If true, the project is currently open for enrollment
	Open bool

	// The sampling rates for each treatment group.
	SamplingRates []float64
}

// EncodedProject is a version of Project that can be stored in the
// datastore.  Appengine datastore doesn't handle structs containing
// slices of other structs.
type EncodedProject struct {
	Owner           string
	Created         time.Time
	Name            string
	GroupNames      []byte
	Variables       []byte
	Assignments     []int
	Data            []byte
	Bias            int
	Comments        []byte
	Modified        time.Time
	StoreRawData    bool
	RawData         []byte
	NumAssignments  int
	RemovedSubjects []string
	Open            bool
	SamplingRates   []float64
}

type EncodedProjectView struct {
	Owner           string
	CreatedDate     string
	CreatedTime     string
	Name            string
	Key             string
	GroupNames      []string
	Index           int
	Assignments     []int
	Data            []byte
	Bias            int
	Comments        []*Comment
	ModifiedDate    string
	ModifiedTime    string
	ModifiedEver    bool
	StoreRawData    bool
	RawData         []byte
	NumAssignments  int
	RemovedSubjects []string
	Open            bool
	SamplingRates   string
}

// ProjectView is a printable version of Project.
type ProjectView struct {
	Owner           string
	CreatedDate     string
	CreatedTime     string
	GroupNames      string
	Name            string
	Variables       []VariableView
	Key             string
	Assignments     []int
	Data            [][][]float64
	Bias            string
	Comments        []*Comment
	ModifiedDate    string
	ModifiedTime    string
	StoreRawData    bool
	RawData         []byte
	NumAssignments  int
	RemovedSubjects []string
	Open            bool
	SamplingRates   string
}

// Variable contains information about one variable that will be used
// as part of the treatment assignment.
type Variable struct {
	Name   string
	Levels []string
	Weight float64
	Func   string
}

// VariableView is a printable version of a variable.
type VariableView struct {
	Name   string
	Levels string
	Index  int
	Weight string
	Func   string
}

// SharingByUser is a record of all projects to which the given user
// has access.
type SharingByUser struct {
	User     string
	Projects string // Comma separated list of projects
}

// SharingByProject is a record of all users who have access to a
// given project.
type SharingByProject struct {
	ProjectName string
	Users       string
}

// Comment stores a single comment.
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
func Clean_split(s string, sep string) []string {

	if len(s) == 0 {
		return []string{}
	}

	parts := strings.Split(s, sep)

	for i, v := range parts {
		parts[i] = strings.TrimSpace(v)
	}

	return parts
}

// Copy_encoded_project creates a copy of a given project and returns
// it.  This is not necessarily a deep copy.
func Copy_encoded_project(proj *EncodedProject) *EncodedProject {

	newproj := new(EncodedProject)

	newproj.Owner = proj.Owner
	newproj.Name = proj.Name
	newproj.Created = time.Now()

	newproj.GroupNames = make([]byte, len(proj.GroupNames))
	copy(newproj.GroupNames, proj.GroupNames)

	newproj.Variables = make([]byte, len(proj.Variables))
	copy(newproj.Variables, proj.Variables)

	newproj.Assignments = make([]int, len(proj.Assignments))
	copy(newproj.Assignments, proj.Assignments)

	newproj.Data = make([]byte, len(proj.Data))
	copy(newproj.Data, proj.Data)

	newproj.Bias = proj.Bias

	newproj.Comments = make([]byte, len(proj.Comments))
	copy(newproj.Comments, proj.Comments)

	newproj.Modified = proj.Modified
	newproj.StoreRawData = proj.StoreRawData

	newproj.RawData = make([]byte, len(proj.RawData))
	copy(newproj.RawData, proj.RawData)

	newproj.NumAssignments = proj.NumAssignments

	newproj.RemovedSubjects = make([]string, len(proj.RemovedSubjects))
	copy(newproj.RemovedSubjects, proj.RemovedSubjects)

	newproj.Open = proj.Open

	newproj.SamplingRates = make([]float64, len(proj.SamplingRates))
	copy(newproj.SamplingRates, proj.SamplingRates)

	return (newproj)
}

// Get_project_from_key
func Get_project_from_key(pkey string,
	c *appengine.Context) (*Project, error) {

	Key := datastore.NewKey(*c, "EncodedProject", pkey, 0, nil)
	var eproj EncodedProject
	err := datastore.Get(*c, Key, &eproj)
	if err != nil {
		(*c).Errorf("Project_dashboard: %v", err)
		return nil, err
	}

	project := Decode_Project(&eproj)

	// This field was added later, so may be missing on some
	// projects.  Provide a default here.
	if project.SamplingRates == nil {
		arr := make([]float64, len(project.GroupNames))
		for i, _ := range arr {
			arr[i] = 1.0
		}
		project.SamplingRates = arr
	}

	return project, nil
}

// Encode_Project takes a Project struct and converts it into a form
// that can be stored in the datastore.
func Encode_Project(proj *Project) (*EncodedProject, error) {

	eproj := new(EncodedProject)

	eproj.Owner = proj.Owner
	eproj.Created = proj.Created
	eproj.Name = proj.Name

	dc, err := json.Marshal(proj.Data)
	if err != nil {
		return nil, err
	}
	eproj.Data = dc

	eproj.Assignments = proj.Assignments
	eproj.Bias = proj.Bias
	eproj.Modified = proj.Modified
	eproj.StoreRawData = proj.StoreRawData
	eproj.NumAssignments = proj.NumAssignments
	eproj.RemovedSubjects = proj.RemovedSubjects
	eproj.Open = proj.Open
	eproj.SamplingRates = proj.SamplingRates

	// Group names
	x1, err := json.Marshal(proj.GroupNames)
	if err != nil {
		return nil, err
	}
	eproj.GroupNames = x1

	// Variables
	x2, err := json.Marshal(proj.Variables)
	if err != nil {
		return nil, err
	}
	eproj.Variables = x2

	// Raw data
	if proj.StoreRawData {
		x3, err := json.Marshal(proj.RawData)
		if err != nil {
			return nil, err
		}
		eproj.RawData = x3
	}

	// Comments
	x4, err := json.Marshal(proj.Comments)
	if err != nil {
		return nil, err
	}
	eproj.Comments = x4

	return eproj, nil
}

// Store_project
func Store_project(proj *Project,
	project_key string,
	c *appengine.Context) error {

	eproj, err := Encode_Project(proj)
	if err != nil {
		return err
	}

	pkey := datastore.NewKey(*c, "EncodedProject", project_key, 0, nil)
	_, err = datastore.Put(*c, pkey, eproj)
	return err
}

// Decode_Project takes a project in its encoded form (storable in the
// datastore) and converts it to a Project struct.
func Decode_Project(eproj *EncodedProject) *Project {

	proj := new(Project)

	proj.Owner = eproj.Owner
	proj.Created = eproj.Created
	proj.Name = eproj.Name

	var data [][][]float64
	json.Unmarshal(eproj.Data, &data)
	proj.Data = data

	proj.Assignments = eproj.Assignments
	proj.Bias = eproj.Bias
	proj.StoreRawData = eproj.StoreRawData
	proj.NumAssignments = eproj.NumAssignments
	proj.RemovedSubjects = eproj.RemovedSubjects
	proj.Modified = eproj.Modified
	proj.Open = eproj.Open
	proj.SamplingRates = eproj.SamplingRates

	var group_names []string
	json.Unmarshal(eproj.GroupNames, &group_names)
	proj.GroupNames = group_names

	var vbls []Variable
	json.Unmarshal(eproj.Variables, &vbls)
	proj.Variables = vbls

	var comments []*Comment
	json.Unmarshal(eproj.Comments, &comments)
	proj.Comments = comments

	if eproj.StoreRawData {
		var rawdata []*DataRecord
		json.Unmarshal(eproj.RawData, &rawdata)
		proj.RawData = rawdata
	}

	return proj
}

// Format_project returns a ProjectView object corresponding to the
// given Project and Key object.
func Format_project(project *Project) *ProjectView {

	B := new(ProjectView)
	B.Owner = project.Owner
	B.Data = project.Data
	B.Name = project.Name
	B.Comments = project.Comments
	B.Assignments = project.Assignments
	B.Bias = fmt.Sprintf("%d", project.Bias)
	t := project.Created
	loc, _ := time.LoadLocation("America/New_York")
	t = t.In(loc)
	B.CreatedDate = t.Format("2006-1-2")
	B.CreatedTime = t.Format("3:04pm")
	B.GroupNames = strings.Join(project.GroupNames, ",")
	B.Variables = make([]VariableView, len(project.Variables))

	rate_str := make([]string, len(project.SamplingRates))
	for i, x := range project.SamplingRates {
		rate_str[i] = fmt.Sprintf("%.0f", x)
	}
	B.SamplingRates = strings.Join(rate_str, ",")

	for i, pv := range project.Variables {
		B.Variables[i] = Format_Variable(pv)
	}
	B.RemovedSubjects = project.RemovedSubjects
	B.Open = project.Open

	t = project.Modified
	loc, _ = time.LoadLocation("America/New_York")
	t = t.In(loc)
	B.ModifiedDate = t.Format("2006-1-2")
	B.ModifiedTime = t.Format("3:04pm")

	return B
}

// Format_EncodedProject returns an EncodedProjectView object
// corresponding to the given Encoded_project object.
func Format_encoded_project(enc_project *EncodedProject) *EncodedProjectView {

	view := new(EncodedProjectView)
	view.Owner = enc_project.Owner
	view.Name = enc_project.Name
	view.Data = enc_project.Data
	view.Assignments = enc_project.Assignments
	view.Bias = enc_project.Bias
	view.Key = enc_project.Owner + "::" + enc_project.Name
	view.NumAssignments = enc_project.NumAssignments

	var s []string
	json.Unmarshal(enc_project.GroupNames, &s)
	view.GroupNames = s

	view.RemovedSubjects = enc_project.RemovedSubjects
	//view.Comments = enc_project.Comments
	view.Open = enc_project.Open

	rate_str := make([]string, len(enc_project.SamplingRates))
	for i, x := range enc_project.SamplingRates {
		rate_str[i] = fmt.Sprintf("%.0f", x)
	}
	view.SamplingRates = strings.Join(rate_str, ",")

	// Created date
	t := enc_project.Created
	loc, _ := time.LoadLocation("America/New_York")
	t = t.In(loc)
	view.CreatedDate = t.Format("2006-1-2")
	view.CreatedTime = t.Format("3:04pm")

	// Modified date
	t = enc_project.Modified

	if t.IsZero() {
		view.ModifiedEver = false
	} else {
		view.ModifiedEver = true
		t = t.In(loc)
		view.ModifiedDate = t.Format("2006-1-2")
		view.ModifiedTime = t.Format("3:04pm")
	}

	return view
}

// Format_EncodedProject returns an array of EncodedProjectView
// objects corresponding to the given array of EncodedProject
// objects.
func Format_encoded_projects(proj []*EncodedProject) []*EncodedProjectView {

	n := len(proj)
	B := make([]*EncodedProjectView, n, n)
	for i := 0; i < n; i++ {
		B[i] = Format_encoded_project(proj[i])
	}

	return B
}

// Format_projects
func Format_projects(projects []*Project) []*ProjectView {

	n := len(projects)
	fmt_projects := make([]*ProjectView, n, n)
	for i := 0; i < n; i++ {
		fmt_projects[i] = Format_project(projects[i])
	}

	return fmt_projects
}

// Format_variables returns an array of VariableView objects
// corresponding to the given array of Variable objects.
func Format_variables(val []Variable) []VariableView {

	valf := make([]VariableView, len(val))

	for i, va := range val {
		valf[i] = Format_Variable(va)
		valf[i].Index = i
	}

	return valf
}

// Format_variable returns a VariableView object corresponding to the
// given Variable object.
func Format_Variable(va Variable) VariableView {

	var vv VariableView
	vv.Name = va.Name
	vv.Levels = strings.Join(va.Levels, ",")
	vv.Weight = fmt.Sprintf("%.0f", va.Weight)
	vv.Func = va.Func

	return vv
}

// Get_shared_users returns a list of user id's for for users who are
// shared for the given project.
func Get_shared_users(project_name string,
	c *appengine.Context) ([]string, error) {

	key := datastore.NewKey(*c, "SharingByProject", project_name, 0, nil)
	var sproj SharingByProject
	err := datastore.Get(*c, key, &sproj)
	if err == datastore.ErrNoSuchEntity {
		return []string{}, nil
	} else if err != nil {
		return []string{}, err
	}
	if len(sproj.Users) == 0 {
		return []string{}, nil
	}
	users := Clean_split(sproj.Users, ",")
	return users, nil
}

// Add_sharing adds all the given users to be shared for the given
// project.
func Add_sharing(project_name string,
	user_names []string,
	c *appengine.Context) error {

	if len(user_names) == 0 {
		return nil
	}

	// Update SharingByProject.
	key := datastore.NewKey(*c, "SharingByProject", project_name, 0, nil)
	sbproj := new(SharingByProject)
	err := datastore.Get(*c, key, sbproj)
	if err == datastore.ErrNoSuchEntity {
		(*c).Errorf("Add_sharing [1]: %v", err)
		// Create a new SharingByProject and carry on
		sbproj.ProjectName = project_name
		sbproj.Users = strings.Join(user_names, ",")
	} else if err != nil {
		return err
	} else {
		U := Clean_split(sbproj.Users, ",")
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
		sbproj.Users = strings.Join(A, ",")
	}

	_, err = datastore.Put(*c, key, sbproj)
	if err != nil {
		return err
	}

	// Update SharingByUser.
	for _, user_name := range user_names {
		key = datastore.NewKey(*c, "SharingByUser",
			strings.ToLower(user_name), 0, nil)
		sbuser := new(SharingByUser)
		err := datastore.Get(*c, key, sbuser)
		if err == datastore.ErrNoSuchEntity {
			sbuser = new(SharingByUser)
			sbuser.User = user_name
			sbuser.Projects = project_name
		} else if err != nil {
			return err
		} else {
			U := Clean_split(sbuser.Projects, ",")
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
			sbuser.Projects = strings.Join(A, ",")
		}
		_, err = datastore.Put(*c, key, sbuser)
		if err != nil {
			return err
		}
	}

	return nil
}

// unique_svec returns an array containing the unique elements of the
// given array.
func unique_svec(vec []string) []string {
	mp := make(map[string]bool)
	for _, x := range vec {
		mp[x] = true
	}
	uvec := make([]string, len(mp))
	i := 0
	for k, _ := range mp {
		uvec[i] = k
		i++
	}
	return uvec
}

// sdiff returns the given vector with the elements in rm removed.
func sdiff(vec []string, rm map[string]bool) []string {

	dvec := make([]string, 0, len(vec))
	for _, v := range vec {
		_, ok := rm[v]
		if !ok {
			dvec = append(dvec, v)
		}
	}
	return dvec
}

// Remove_sharing removes the given users from the access list for the given project.
func Remove_sharing(project_name string,
	user_names []string,
	c *appengine.Context) error {

	// Map whose keys are the users to remove.
	rmu := make(map[string]bool)
	for i := 0; i < len(user_names); i++ {
		rmu[user_names[i]] = true
	}

	// Update SharingByProject.
	key := datastore.NewKey(*c, "SharingByProject", project_name, 0, nil)
	sproj := new(SharingByProject)
	err := datastore.Get(*c, key, sproj)
	if err == datastore.ErrNoSuchEntity {
		// OK
	} else if err != nil {
		return err
	} else {
		users := Clean_split(sproj.Users, ",")
		users = unique_svec(users)
		users = sdiff(users, rmu)
		sproj.Users = strings.Join(users, ",")
		_, err = datastore.Put(*c, key, sproj)
		if err != nil {
			return err
		}
	}

	// Update SharingByUser.
	for _, name := range user_names {
		pkey := datastore.NewKey(*c, "SharingByUser", strings.ToLower(name), 0, nil)
		suser := new(SharingByUser)
		err := datastore.Get(*c, pkey, suser)
		if err == datastore.ErrNoSuchEntity {
			// should not reach here
		} else if err != nil {
			return err
		} else {
			projlist := Clean_split(suser.Projects, ",")
			projlist = unique_svec(projlist)
			projlist = sdiff(projlist, map[string]bool{project_name: true})
			suser.Projects = strings.Join(projlist, ",")

			_, err = datastore.Put(*c, pkey, suser)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetProjects returns all projects owned by the given user.
// Optionally also include projects that are shared with the user.
func GetProjects(user string,
	include_shared bool,
	c *appengine.Context) ([]*datastore.Key, []*EncodedProject, error) {

	qr := datastore.NewQuery("EncodedProject").
		Filter("Owner = ", user).
		Order("-Created").Limit(100)

	keyset := make(map[string]bool)

	var projlist []*EncodedProject
	keylist, err := qr.GetAll(*c, &projlist)
	if err != nil {
		(*c).Errorf("GetProjects[1]: %v", err)
		return nil, nil, err
	}

	if !include_shared {
		return keylist, projlist, err
	}

	for _, k := range keylist {
		keyset[k.String()] = true
	}

	// Get project ids that are shared with this user
	ky2 := datastore.NewKey(*c, "SharingByUser", strings.ToLower(user),
		0, nil)
	var spu SharingByUser
	err = datastore.Get(*c, ky2, &spu)
	if err == datastore.ErrNoSuchEntity {
		// No projects shared with this user
		return keylist, projlist, nil
	}
	if err != nil {
		(*c).Errorf("GetProjects[2]: %v", err)
		return nil, nil, err
	}

	// Get the shared projects
	spvl := Clean_split(spu.Projects, ",")
	for _, spv := range spvl {
		ky := datastore.NewKey(*c, "EncodedProject", spv, 0, nil)
		_, ok := keyset[ky.String()]
		if ok {
			continue
		}
		keyset[ky.String()] = true
		pr := new(EncodedProject)
		err = datastore.Get(*c, ky, pr)
		if err != nil {
			(*c).Infof("GetProjects [3]: %v\n%v", spv, err)
			continue
		}
		keylist = append(keylist, ky)
		projlist = append(projlist, pr)
	}
	return keylist, projlist, nil
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
	(*c).Errorf("ServeError [1]: %v", err)
	fmt.Printf("\n%v\n", err)
}

// Message_page presents a simple message page and presents the user
// with a link that leads to a followup page.
func Message_page(w http.ResponseWriter,
	r *http.Request,
	login_user *user.User,
	msg string,
	return_msg string,
	return_url string) {

	c := appengine.NewContext(r)
	tmpl, err := template.ParseFiles("header.html",
		"message.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	type TV struct {
		User      string
		LoggedIn  bool
		Msg       string
		ReturnUrl string
		ReturnMsg string
		LogoutUrl string
	}

	template_values := new(TV)
	if login_user != nil {
		template_values.User = login_user.String()
	} else {
		template_values.User = ""
	}
	template_values.LoggedIn = login_user != nil
	template_values.Msg = msg
	template_values.ReturnUrl = return_url
	template_values.ReturnMsg = return_msg

	if err := tmpl.ExecuteTemplate(w, "message.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// get_index returns the position of `val` within `vec`.
func get_index(vec []string, val string) int {

	for i, x := range vec {
		if val == x {
			return i
		}
	}
	return -1
}

// Remove_from_aggregate updates the aggregate statistics (count per
// treatment arm for each level of each variable) for the given data
// record.
func Remove_from_aggregate(rec *DataRecord,
	proj *Project) {

	grp_ix := get_index(proj.GroupNames, rec.CurrentGroup)

	// Update the overall assignment totals
	proj.Assignments[grp_ix] -= 1

	// Update the within-variable assignment totals
	data := proj.Data
	for j, va := range proj.Variables {
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				data[j][k][grp_ix] -= 1
			}
		}
	}
}

// Add_to_aggregate updates the aggregate statistics (count per
// treatment arm for each level of each variable) for the given data
// record.
func Add_to_aggregate(rec *DataRecord,
	proj *Project) {

	grp_ix := get_index(proj.GroupNames, rec.CurrentGroup)

	// Update the overall assignment totals
	proj.Assignments[grp_ix] += 1

	// Update the within-variable assignment totals
	data := proj.Data
	for j, va := range proj.Variables {
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				data[j][k][grp_ix] += 1
			}
		}
	}
}
