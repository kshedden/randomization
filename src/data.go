package randomization

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
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

// cleanSplit splits a string into tokens delimited by a given
// separator.  If S equals the empty string, this function returns an
// empty list, rather than a list containing an empty string as its
// sole element.  Leading and trailing whitespace is removed from each
// element of the returned list.
func cleanSplit(s string, sep string) []string {

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
func copyEncodedProject(proj *EncodedProject) *EncodedProject {

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

// getProjectfromKey
func getProjectFromKey(ctx context.Context, pkey string) (*Project, error) {

	Key := datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	var eproj EncodedProject
	err := datastore.Get(ctx, Key, &eproj)
	if err != nil {
		log.Errorf(ctx, "Project_dashboard: %v", err)
		return nil, err
	}

	project := decodeProject(&eproj)

	// This field was added later, so may be missing on some
	// projects.  Provide a default here.
	if project.SamplingRates == nil {
		arr := make([]float64, len(project.GroupNames))
		for i := range arr {
			arr[i] = 1.0
		}
		project.SamplingRates = arr
	}

	return project, nil
}

// encodeProject takes a Project struct and converts it into a form
// that can be stored in the datastore.
func encodeProject(proj *Project) (*EncodedProject, error) {

	ep := new(EncodedProject)

	ep.Owner = proj.Owner
	ep.Created = proj.Created
	ep.Name = proj.Name

	dc, err := json.Marshal(proj.Data)
	if err != nil {
		return nil, err
	}
	ep.Data = dc

	ep.Assignments = proj.Assignments
	ep.Bias = proj.Bias
	ep.Modified = proj.Modified
	ep.StoreRawData = proj.StoreRawData
	ep.NumAssignments = proj.NumAssignments
	ep.RemovedSubjects = proj.RemovedSubjects
	ep.Open = proj.Open
	ep.SamplingRates = proj.SamplingRates

	// Group names
	x1, err := json.Marshal(proj.GroupNames)
	if err != nil {
		return nil, err
	}
	ep.GroupNames = x1

	// Variables
	x2, err := json.Marshal(proj.Variables)
	if err != nil {
		return nil, err
	}
	ep.Variables = x2

	// Raw data
	if proj.StoreRawData {
		x3, err := json.Marshal(proj.RawData)
		if err != nil {
			return nil, err
		}
		ep.RawData = x3
	}

	// Comments
	x4, err := json.Marshal(proj.Comments)
	if err != nil {
		return nil, err
	}
	ep.Comments = x4

	return ep, nil
}

// storeProject
func storeProject(ctx context.Context, proj *Project, projectKey string) error {

	ep, err := encodeProject(proj)
	if err != nil {
		return err
	}

	pkey := datastore.NewKey(ctx, "EncodedProject", projectKey, 0, nil)
	_, err = datastore.Put(ctx, pkey, ep)

	return err
}

// decodeProject takes a project in its encoded form (storable in the
// datastore) and converts it to a Project struct.
func decodeProject(eproj *EncodedProject) *Project {

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

	var groupNames []string
	json.Unmarshal(eproj.GroupNames, &groupNames)
	proj.GroupNames = groupNames

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

// formatProject returns a ProjectView object corresponding to the
// given Project and Key object.
func formatProject(project *Project) *ProjectView {

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

	rateStr := make([]string, len(project.SamplingRates))
	for i, x := range project.SamplingRates {
		rateStr[i] = fmt.Sprintf("%.0f", x)
	}
	B.SamplingRates = strings.Join(rateStr, ",")

	for i, pv := range project.Variables {
		B.Variables[i] = formatVariable(pv)
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

// formatEncodedProject returns an EncodedProjectView object
// corresponding to the given Encoded_project object.
func formatEncodedProject(encProject *EncodedProject) *EncodedProjectView {

	view := new(EncodedProjectView)
	view.Owner = encProject.Owner
	view.Name = encProject.Name
	view.Data = encProject.Data
	view.Assignments = encProject.Assignments
	view.Bias = encProject.Bias
	view.Key = encProject.Owner + "::" + encProject.Name
	view.NumAssignments = encProject.NumAssignments

	var s []string
	json.Unmarshal(encProject.GroupNames, &s)
	view.GroupNames = s

	view.RemovedSubjects = encProject.RemovedSubjects
	//view.Comments = enc_project.Comments
	view.Open = encProject.Open

	rateStr := make([]string, len(encProject.SamplingRates))
	for i, x := range encProject.SamplingRates {
		rateStr[i] = fmt.Sprintf("%.0f", x)
	}
	view.SamplingRates = strings.Join(rateStr, ",")

	// Created date
	t := encProject.Created
	loc, _ := time.LoadLocation("America/New_York")
	t = t.In(loc)
	view.CreatedDate = t.Format("2006-1-2")
	view.CreatedTime = t.Format("3:04pm")

	// Modified date
	t = encProject.Modified

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

// formatEncodedProject returns an array of EncodedProjectView
// objects corresponding to the given array of EncodedProject
// objects.
func formatEncodedProjects(proj []*EncodedProject) []*EncodedProjectView {

	n := len(proj)
	B := make([]*EncodedProjectView, n, n)
	for i := 0; i < n; i++ {
		B[i] = formatEncodedProject(proj[i])
	}

	return B
}

// formatProjects
func formatProjects(projects []*Project) []*ProjectView {

	n := len(projects)
	fmtProjects := make([]*ProjectView, n, n)
	for i := 0; i < n; i++ {
		fmtProjects[i] = formatProject(projects[i])
	}

	return fmtProjects
}

// formatVariables returns an array of VariableView objects
// corresponding to the given array of Variable objects.
func formatVariables(val []Variable) []VariableView {

	valf := make([]VariableView, len(val))

	for i, va := range val {
		valf[i] = formatVariable(va)
		valf[i].Index = i
	}

	return valf
}

// formatVariable returns a VariableView object corresponding to the
// given Variable object.
func formatVariable(va Variable) VariableView {

	var vv VariableView
	vv.Name = va.Name
	vv.Levels = strings.Join(va.Levels, ",")
	vv.Weight = fmt.Sprintf("%.0f", va.Weight)
	vv.Func = va.Func

	return vv
}

// getSharedUsers returns a list of user id's for for users who are
// shared for the given project.
func getSharedUsers(ctx context.Context, projectName string) ([]string, error) {

	key := datastore.NewKey(ctx, "SharingByProject", projectName, 0, nil)
	var sproj SharingByProject
	err := datastore.Get(ctx, key, &sproj)
	if err == datastore.ErrNoSuchEntity {
		return []string{}, nil
	} else if err != nil {
		return []string{}, err
	}
	if len(sproj.Users) == 0 {
		return []string{}, nil
	}
	users := cleanSplit(sproj.Users, ",")
	return users, nil
}

// addSharing adds all the given users to be shared for the given
// project.
func addSharing(ctx context.Context, projectName string, userNames []string) error {

	if len(userNames) == 0 {
		return nil
	}

	// Update SharingByProject.
	key := datastore.NewKey(ctx, "SharingByProject", projectName, 0, nil)
	sbproj := new(SharingByProject)
	err := datastore.Get(ctx, key, sbproj)
	if err == datastore.ErrNoSuchEntity {
		log.Errorf(ctx, "Add_sharing [1]: %v", err)
		// Create a new SharingByProject and carry on
		sbproj.ProjectName = projectName
		sbproj.Users = strings.Join(userNames, ",")
	} else if err != nil {
		return err
	} else {
		U := cleanSplit(sbproj.Users, ",")
		m := make(map[string]bool)
		for _, u := range U {
			m[u] = true
		}
		for _, u := range userNames {
			m[u] = true
		}
		A := make([]string, len(m))
		i := 0
		for k := range m {
			A[i] = k
			i++
		}
		sbproj.Users = strings.Join(A, ",")
	}

	_, err = datastore.Put(ctx, key, sbproj)
	if err != nil {
		return err
	}

	// Update SharingByUser.
	for _, uname := range userNames {
		key = datastore.NewKey(ctx, "SharingByUser", strings.ToLower(uname), 0, nil)
		sbuser := new(SharingByUser)
		err := datastore.Get(ctx, key, sbuser)
		if err == datastore.ErrNoSuchEntity {
			sbuser = new(SharingByUser)
			sbuser.User = uname
			sbuser.Projects = projectName
		} else if err != nil {
			return err
		} else {
			U := cleanSplit(sbuser.Projects, ",")
			m := make(map[string]bool)
			for _, u := range U {
				m[u] = true
			}
			m[projectName] = true
			A := make([]string, len(m))
			i := 0
			for k := range m {
				A[i] = k
				i++
			}
			sbuser.Projects = strings.Join(A, ",")
		}
		_, err = datastore.Put(ctx, key, sbuser)
		if err != nil {
			return err
		}
	}

	return nil
}

// uniqueSvec returns an array containing the unique elements of the
// given array.
func uniqueSvec(vec []string) []string {

	mp := make(map[string]bool)
	for _, x := range vec {
		mp[x] = true
	}

	uvec := make([]string, len(mp))
	i := 0
	for k := range mp {
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

// removeSharing removes the given users from the access list for the given project.
func removeSharing(ctx context.Context, projectName string, userNames []string) error {

	// Map whose keys are the users to remove.
	rmu := make(map[string]bool)
	for i := 0; i < len(userNames); i++ {
		rmu[userNames[i]] = true
	}

	// Update SharingByProject.
	key := datastore.NewKey(ctx, "SharingByProject", projectName, 0, nil)
	sproj := new(SharingByProject)
	err := datastore.Get(ctx, key, sproj)
	if err == datastore.ErrNoSuchEntity {
		// OK
	} else if err != nil {
		return err
	} else {
		users := cleanSplit(sproj.Users, ",")
		users = uniqueSvec(users)
		users = sdiff(users, rmu)
		sproj.Users = strings.Join(users, ",")
		_, err = datastore.Put(ctx, key, sproj)
		if err != nil {
			return err
		}
	}

	// Update SharingByUser.
	for _, name := range userNames {
		pkey := datastore.NewKey(ctx, "SharingByUser", strings.ToLower(name), 0, nil)
		suser := new(SharingByUser)
		err := datastore.Get(ctx, pkey, suser)
		if err == datastore.ErrNoSuchEntity {
			// should not reach here
		} else if err != nil {
			return err
		} else {
			projlist := cleanSplit(suser.Projects, ",")
			projlist = uniqueSvec(projlist)
			projlist = sdiff(projlist, map[string]bool{projectName: true})
			suser.Projects = strings.Join(projlist, ",")

			_, err = datastore.Put(ctx, pkey, suser)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getProjects returns all projects owned by the given user.
// Optionally also include projects that are shared with the user.
func getProjects(ctx context.Context, user string, includeShared bool) ([]*datastore.Key, []*EncodedProject, error) {

	qr := datastore.NewQuery("EncodedProject").
		Filter("Owner = ", user).
		Order("-Created").Limit(100)

	keyset := make(map[string]bool)

	var projlist []*EncodedProject
	keylist, err := qr.GetAll(ctx, &projlist)
	if err != nil {
		log.Errorf(ctx, "GetProjects[1]: %v", err)
		return nil, nil, err
	}

	if !includeShared {
		return keylist, projlist, err
	}

	for _, k := range keylist {
		keyset[k.String()] = true
	}

	// Get project ids that are shared with this user
	ky2 := datastore.NewKey(ctx, "SharingByUser", strings.ToLower(user), 0, nil)
	var spu SharingByUser
	err = datastore.Get(ctx, ky2, &spu)
	if err == datastore.ErrNoSuchEntity {
		// No projects shared with this user
		return keylist, projlist, nil
	}
	if err != nil {
		log.Errorf(ctx, "GetProjects[2]: %v", err)
		return nil, nil, err
	}

	// Get the shared projects
	spvl := cleanSplit(spu.Projects, ",")
	for _, spv := range spvl {
		ky := datastore.NewKey(ctx, "EncodedProject", spv, 0, nil)
		_, ok := keyset[ky.String()]
		if ok {
			continue
		}
		keyset[ky.String()] = true
		pr := new(EncodedProject)
		err = datastore.Get(ctx, ky, pr)
		if err != nil {
			log.Infof(ctx, "GetProjects [3]: %v\n%v", spv, err)
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

func ServeError(ctx context.Context, w http.ResponseWriter, err error) {

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "Internal Server Error")
	log.Errorf(ctx, "ServeError [1]: %v", err)
	fmt.Printf("\n%v\n", err)
}

// messagePage presents a simple message page and presents the user
// with a link that leads to a followup page.
func messagePage(w http.ResponseWriter, r *http.Request, loginUser *user.User,
	msg string, rmsg string, returnURL string) {

	ctx := appengine.NewContext(r)

	type TV struct {
		User      string
		LoggedIn  bool
		Msg       string
		ReturnUrl string
		ReturnMsg string
		LogoutUrl string
	}

	templateValues := new(TV)
	if loginUser != nil {
		templateValues.User = loginUser.String()
	} else {
		templateValues.User = ""
	}
	templateValues.LoggedIn = loginUser != nil
	templateValues.Msg = msg
	templateValues.ReturnUrl = returnURL
	templateValues.ReturnMsg = rmsg

	if err := tmpl.ExecuteTemplate(w, "message.html", templateValues); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// getIndex returns the position of `val` within `vec`.
func getIndex(vec []string, val string) int {

	for i, x := range vec {
		if val == x {
			return i
		}
	}
	return -1
}

// removeFromAggregate updates the aggregate statistics (count per
// treatment arm for each level of each variable) for the given data
// record.
func removeFromAggregate(rec *DataRecord, proj *Project) {

	grpIx := getIndex(proj.GroupNames, rec.CurrentGroup)

	// Update the overall assignment totals
	proj.Assignments[grpIx]--

	// Update the within-variable assignment totals
	data := proj.Data
	for j, va := range proj.Variables {
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				data[j][k][grpIx]--
			}
		}
	}
}

// addToAggregate updates the aggregate statistics (count per
// treatment arm for each level of each variable) for the given data
// record.
func addToAggregate(rec *DataRecord,
	proj *Project) {

	grpIx := getIndex(proj.GroupNames, rec.CurrentGroup)

	// Update the overall assignment totals
	proj.Assignments[grpIx]++

	// Update the within-variable assignment totals
	data := proj.Data
	for j, va := range proj.Variables {
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				data[j][k][grpIx]++
			}
		}
	}
}
