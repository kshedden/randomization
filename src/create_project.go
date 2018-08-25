package randomization

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// createProjectStep1 gets the project name from the user.
func createProjectStep1(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	tvals := struct {
		User     string
		LoggedIn bool
	}{
		User:     user.String(),
		LoggedIn: user != nil,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step1.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep1 failed to execute template: %v", err)
	}
}

// createProjectStep2 asks if the subject-level data are to be logged.
func createProjectStep2(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	user := user.Current(ctx)
	projectName := r.FormValue("project_name")

	// Check if the project name has already been used.
	pkey := user.String() + "::" + projectName
	key := datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	var pr EncodedProject
	err := datastore.Get(ctx, key, &pr)
	if err == nil {
		msg := fmt.Sprintf("A project named \"%s\" belonging to user %s already exists.", projectName, user.String())
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	tvals := struct {
		User     string
		LoggedIn bool
		Name     string
		Pkey     string
	}{
		User:     user.String(),
		LoggedIn: user != nil,
		Name:     r.FormValue("project_name"),
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step2.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep2 failed to execute template: %v", err)
	}
}

// createProjectStep3 gets the number of treatment groups.
func createProjectStep3(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	user := user.Current(ctx)

	tvals := struct {
		User         string
		LoggedIn     bool
		Name         string
		Pkey         string
		StoreRawData bool
	}{
		User:         user.String(),
		LoggedIn:     user != nil,
		Name:         r.FormValue("project_name"),
		StoreRawData: r.FormValue("store_rawdata") == "yes",
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step3.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep3 failed to execute template: %v", err)
	}
}

// createProjectStep4
func createProjectStep4(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	user := user.Current(ctx)

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))

	// Group numbers (they don't have names yet)
	IX := make([]int, numgroups, numgroups)
	for i := 0; i < numgroups; i++ {
		IX[i] = i + 1
	}

	tvals := struct {
		User         string
		LoggedIn     bool
		Name         string
		Pkey         string
		StoreRawData bool
		NumGroups    int
		IX           []int
	}{
		User:         user.String(),
		LoggedIn:     user != nil,
		Name:         r.FormValue("project_name"),
		StoreRawData: r.FormValue("store_rawdata") == "true",
		IX:           IX,
		NumGroups:    numgroups,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step4.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep4 failed to execute template: %v", err)
	}
}

// createProjectStep5
func createProjectStep5(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	user := user.Current(ctx)

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))

	// Indices for the groups
	IX := make([]int, numgroups, numgroups)
	for i := 0; i < numgroups; i++ {
		IX[i] = i
	}

	// Get the group names from the previous page
	GroupNames := make([]string, numgroups, numgroups)
	for i := 0; i < numgroups; i++ {
		GroupNames[i] = r.FormValue(fmt.Sprintf("name%d", i+1))
	}

	tvals := struct {
		User           string
		LoggedIn       bool
		Name           string
		Pkey           string
		GroupNames     string
		GroupNames_arr []string
		StoreRawData   bool
		NumGroups      int
		IX             []int
	}{
		User:           user.String(),
		LoggedIn:       user != nil,
		Name:           r.FormValue("project_name"),
		GroupNames:     strings.Join(GroupNames, ","),
		GroupNames_arr: GroupNames,
		NumGroups:      len(GroupNames),
		StoreRawData:   r.FormValue("store_rawdata") == "true",
		IX:             IX,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step5.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep5 failed to execute template: %v", err)
	}
}

// createProjectStep6
func createProjectStep6(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	user := user.Current(ctx)

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))

	// Get the sampling rates from the previous page
	groupNamesArr := cleanSplit(r.FormValue("group_names"), ",")
	samplingRates := make([]string, numgroups, numgroups)
	for i := 0; i < numgroups; i++ {
		samplingRates[i] = r.FormValue(fmt.Sprintf("rate%s", groupNamesArr[i]))

		x, err := strconv.ParseFloat(samplingRates[i], 64)
		if (err != nil) || (x <= 0) {
			msg := "The sampling rates must be positive numbers."
			rmsg := "Return to dashboard"
			messagePage(w, r, user, msg, rmsg, "/dashboard")
			return
		}
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		Pkey          string
		GroupNames    string
		StoreRawData  bool
		SamplingRates string
		NumGroups     int
	}{
		User:          user.String(),
		LoggedIn:      user != nil,
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: strings.Join(samplingRates, ","),
		NumGroups:     numgroups,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step6.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep6 failed to execute template: %v", err)
	}
}

// createProjectStep7
func createProjectStep7(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	user := user.Current(ctx)

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))
	numvar, _ := strconv.Atoi(r.FormValue("numvar"))

	IX := make([]int, numvar, numvar)
	for i := 0; i < numvar; i++ {
		IX[i] = i + 1
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		Pkey          string
		IX            []int
		GroupNames    string
		StoreRawData  bool
		NumGroups     int
		NumVar        int
		Any_vars      bool
		SamplingRates string
	}{
		User:          user.String(),
		LoggedIn:      user != nil,
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		IX:            IX,
		NumGroups:     numgroups,
		NumVar:        numvar,
		Any_vars:      (numvar > 0),
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: r.FormValue("rates"),
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step7.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep7 failed to execute template: %v", err)
	}
}

func processVariableInfo(r *http.Request, numvar int) (string, bool) {

	variables := make([]string, numvar, numvar)

	for i := 0; i < numvar; i++ {
		vec := make([]string, 4)

		vname := fmt.Sprintf("name%d", i+1)
		vec[0] = strings.TrimSpace(r.FormValue(vname))
		if len(vec[0]) == 0 {
			return "", false
		}

		vname = fmt.Sprintf("levels%d", i+1)
		vec[1] = r.FormValue(vname)
		levels := cleanSplit(vec[1], ",")
		if len(levels) < 2 {
			return "", false
		}
		for _, x := range levels {
			if len(x) == 0 {
				return "", false
			}
		}

		vec[2] = r.FormValue(fmt.Sprintf("weight%d", i+1))
		vec[3] = r.FormValue(fmt.Sprintf("func%d", i+1))
		variables[i] = strings.Join(vec, ";")
	}

	return strings.Join(variables, ":"), true
}

// createProjectStep8
func createProjectStep8(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))
	numvar, _ := strconv.Atoi(r.FormValue("numvar"))
	variables, ok := processVariableInfo(r, numvar)

	if !ok {
		validationErrorStep8(w, r)
		return
	}

	IX := make([]int, numvar, numvar)
	for i := 0; i < numvar; i++ {
		IX[i] = i + 1
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		Pkey          string
		IX            []int
		GroupNames    string
		StoreRawData  bool
		NumGroups     int
		Numvar        int
		Variables     string
		SamplingRates string
	}{
		User:          user.String(),
		LoggedIn:      user != nil,
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		IX:            IX,
		NumGroups:     numgroups,
		Numvar:        numvar,
		Variables:     variables,
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: r.FormValue("rates"),
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step8.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep8 failed to execute template: %v", err)
	}
}

// createProjectStep9 creates the project using all supplied
// information, and stores the project in the database.
func createProjectStep9(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	user := user.Current(ctx)

	numvar, _ := strconv.Atoi(r.FormValue("numvar"))
	GroupNames := r.FormValue("group_names")
	projectName := r.FormValue("project_name")
	variables := r.FormValue("variables")
	VL := cleanSplit(variables, ":")
	bias, _ := strconv.Atoi(r.FormValue("bias"))
	rates := r.FormValue("rates")

	// Parse and validate the variable information.
	VA := make([]Variable, numvar, numvar)
	for i, vl := range VL {
		vx := cleanSplit(vl, ";")
		var va Variable
		va.Name = vx[0]
		va.Levels = cleanSplit(vx[1], ",")
		va.Weight, _ = strconv.ParseFloat(vx[2], 64)
		va.Func = vx[3]
		VA[i] = va
	}

	var project Project
	project.Owner = user.String()
	project.Created = time.Now()
	project.Name = projectName
	project.Variables = VA
	project.Bias = bias
	project.GroupNames = cleanSplit(GroupNames, ",")
	project.Assignments = make([]int, len(project.GroupNames))
	project.StoreRawData = r.FormValue("store_rawdata") == "true"
	project.Open = true

	// Convert the rates to numbers
	rates = r.FormValue("rates")
	ratesArr := cleanSplit(rates, ",")
	ratesNum := make([]float64, len(ratesArr))
	for i, x := range ratesArr {
		ratesNum[i], _ = strconv.ParseFloat(x, 64)
	}
	project.SamplingRates = ratesNum

	// Set up the data.
	numgroups := len(project.GroupNames)
	data0 := make([][][]float64, len(project.Variables))
	for j, va := range project.Variables {
		data0[j] = make([][]float64, len(va.Levels))
		for k := range va.Levels {
			data0[j][k] = make([]float64, numgroups)
		}
	}
	project.Data = data0

	pkey := user.String() + "::" + projectName
	dkey := datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	eproj, err := encodeProject(&project)
	if err != nil {
		log.Errorf(ctx, "Create_project_step9 [2]: %v", err)
	}
	_, err = datastore.Put(ctx, dkey, eproj)
	if err != nil {
		msg := "A datastore error occured, the project was not created."
		log.Errorf(ctx, "Create_project_step9: %v", err)
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	// Remove any stale SharingByProject entities
	dkey = datastore.NewKey(ctx, "SharingByProject", pkey, 0, nil)
	err = datastore.Delete(ctx, dkey)
	if err != nil {
		log.Errorf(ctx, "Create_project_step9 [3]: %v", err)
	}

	tvals := struct {
		User     string
		LoggedIn bool
	}{
		User:     user.String(),
		LoggedIn: user != nil,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step9.html", tvals); err != nil {
		log.Errorf(ctx, "createProjectStep9 failed to execute template: %v", err)
	}
}

func validationErrorStep8(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	user := user.Current(ctx)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	numgroups, err := strconv.Atoi(r.FormValue("numgroups"))
	if err != nil {
		log.Errorf(ctx, "validationErrorStep8: %v", err)
	}

	numvar, err := strconv.Atoi(r.FormValue("numvar"))
	if err != nil {
		log.Errorf(ctx, "validationErrorStep8: %v", err)
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		NumGroups     int
		Pkey          string
		GroupNames    string
		StoreRawData  bool
		Numvar        int
		SamplingRates string
	}{
		User:          user.String(),
		LoggedIn:      user != nil,
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		NumGroups:     numgroups,
		Numvar:        numvar,
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: r.FormValue("rates"),
	}

	if err := tmpl.ExecuteTemplate(w, "validation_error_step8.html", tvals); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
