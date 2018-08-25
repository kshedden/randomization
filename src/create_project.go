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

// CreateProjectStep1 gets the project name from the user.
func CreateProjectStep1(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	type TV struct {
		User     string
		LoggedIn bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil

	if err := tmpl.ExecuteTemplate(w, "create_project_step1.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// CreateProjectStep2 asks if the subject-level data are to be logged.
func CreateProjectStep2(w http.ResponseWriter, r *http.Request) {

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
	project_name := r.FormValue("project_name")

	// Check if the project name has already been used.
	pkey := user.String() + "::" + project_name
	key := datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	var pr EncodedProject
	err := datastore.Get(ctx, key, &pr)
	if err == nil {
		msg := fmt.Sprintf("A project named \"%s\" belonging to user %s already exists.", project_name, user.String())
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	type TV struct {
		User     string
		LoggedIn bool
		Name     string
		Pkey     string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")

	if err := tmpl.ExecuteTemplate(w, "create_project_step2.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// CreateProjectStep3 gets the number of treatment groups.
func CreateProjectStep3(w http.ResponseWriter, r *http.Request) {

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

	type TV struct {
		User         string
		LoggedIn     bool
		Name         string
		Pkey         string
		StoreRawData bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")
	template_values.StoreRawData = r.FormValue("store_rawdata") == "yes"

	if err := tmpl.ExecuteTemplate(w, "create_project_step3.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// CreateProjectStep4
func CreateProjectStep4(w http.ResponseWriter, r *http.Request) {

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

	type TV struct {
		User         string
		LoggedIn     bool
		Name         string
		Pkey         string
		StoreRawData bool
		NumGroups    int
		IX           []int
	}

	// Group numbers (they don't have names yet)
	IX := make([]int, numgroups, numgroups)
	for i := 0; i < numgroups; i++ {
		IX[i] = i + 1
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")
	template_values.StoreRawData = r.FormValue("store_rawdata") == "true"
	template_values.IX = IX
	template_values.NumGroups = numgroups

	if err := tmpl.ExecuteTemplate(w, "create_project_step4.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// CreateProjectStep5
func CreateProjectStep5(w http.ResponseWriter, r *http.Request) {

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

	type TV struct {
		User           string
		LoggedIn       bool
		Name           string
		Pkey           string
		GroupNames     string
		GroupNames_arr []string
		StoreRawData   bool
		NumGroups      int
		IX             []int
	}

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

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")
	template_values.GroupNames = strings.Join(GroupNames, ",")
	template_values.GroupNames_arr = GroupNames
	template_values.NumGroups = len(GroupNames)
	template_values.StoreRawData = r.FormValue("store_rawdata") == "true"
	template_values.IX = IX

	if err := tmpl.ExecuteTemplate(w, "create_project_step5.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// CreateProjectStep6
func CreateProjectStep6(w http.ResponseWriter, r *http.Request) {

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

	type TV struct {
		User          string
		LoggedIn      bool
		Name          string
		Pkey          string
		GroupNames    string
		StoreRawData  bool
		SamplingRates string
		NumGroups     int
	}

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))

	// Get the sampling rates from the previous page
	group_names_arr := cleanSplit(r.FormValue("group_names"), ",")
	sampling_rates := make([]string, numgroups, numgroups)
	for i := 0; i < numgroups; i++ {
		sampling_rates[i] = r.FormValue(fmt.Sprintf("rate%s", group_names_arr[i]))

		x, err := strconv.ParseFloat(sampling_rates[i], 64)
		if (err != nil) || (x <= 0) {
			msg := "The sampling rates must be positive numbers."
			rmsg := "Return to dashboard"
			messagePage(w, r, user, msg, rmsg, "/dashboard")
			return
		}
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")
	template_values.GroupNames = r.FormValue("group_names")
	template_values.StoreRawData = r.FormValue("store_rawdata") == "true"
	template_values.SamplingRates = strings.Join(sampling_rates, ",")
	template_values.NumGroups = numgroups

	if err := tmpl.ExecuteTemplate(w, "create_project_step6.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// CreateProjectStep7
func CreateProjectStep7(w http.ResponseWriter, r *http.Request) {

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

	type TV struct {
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
	}

	IX := make([]int, numvar, numvar)

	for i := 0; i < numvar; i++ {
		IX[i] = i + 1
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")
	template_values.GroupNames = r.FormValue("group_names")
	template_values.IX = IX
	template_values.NumGroups = numgroups
	template_values.NumVar = numvar
	template_values.Any_vars = (numvar > 0)
	template_values.StoreRawData = r.FormValue("store_rawdata") == "true"
	template_values.SamplingRates = r.FormValue("rates")

	if err := tmpl.ExecuteTemplate(w, "create_project_step7.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

func process_variable_info(r *http.Request, numvar int) (string, bool) {
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

// CreateProjectStep8
func CreateProjectStep8(w http.ResponseWriter, r *http.Request) {

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
	variables, ok := process_variable_info(r, numvar)

	if !ok {
		validation_error_step8(w, r)
		return
	}

	type TV struct {
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
	}

	IX := make([]int, numvar, numvar)

	for i := 0; i < numvar; i++ {
		IX[i] = i + 1
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")
	template_values.GroupNames = r.FormValue("group_names")
	template_values.IX = IX
	template_values.NumGroups = numgroups
	template_values.Numvar = numvar
	template_values.Variables = variables
	template_values.StoreRawData = r.FormValue("store_rawdata") == "true"
	template_values.SamplingRates = r.FormValue("rates")

	if err := tmpl.ExecuteTemplate(w, "create_project_step8.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// CreateProjectStep9 creates the project using all supplied
// information, and stores the project in the database.
func CreateProjectStep9(w http.ResponseWriter, r *http.Request) {

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
	project_name := r.FormValue("project_name")
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
	project.Name = project_name
	project.Variables = VA
	project.Bias = bias
	project.GroupNames = cleanSplit(GroupNames, ",")
	project.Assignments = make([]int, len(project.GroupNames))
	project.StoreRawData = r.FormValue("store_rawdata") == "true"
	project.Open = true

	// Convert the rates to numbers
	rates = r.FormValue("rates")
	rates_arr := cleanSplit(rates, ",")
	rates_num := make([]float64, len(rates_arr))
	for i, x := range rates_arr {
		rates_num[i], _ = strconv.ParseFloat(x, 64)
	}
	project.SamplingRates = rates_num

	// Set up the data.
	numgroups := len(project.GroupNames)
	data0 := make([][][]float64, len(project.Variables))
	for j, va := range project.Variables {
		data0[j] = make([][]float64, len(va.Levels))
		for k, _ := range va.Levels {
			data0[j][k] = make([]float64, numgroups)
		}
	}
	project.Data = data0

	pkey := user.String() + "::" + project_name
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

	type TV struct {
		User     string
		LoggedIn bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil

	if err := tmpl.ExecuteTemplate(w, "create_project_step9.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

func validation_error_step8(w http.ResponseWriter,
	r *http.Request) {

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

	type TV struct {
		User          string
		LoggedIn      bool
		Name          string
		NumGroups     int
		Pkey          string
		GroupNames    string
		StoreRawData  bool
		Numvar        int
		SamplingRates string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Name = r.FormValue("project_name")
	template_values.GroupNames = r.FormValue("group_names")
	template_values.NumGroups, _ = strconv.Atoi(r.FormValue("numgroups"))
	template_values.Numvar, _ = strconv.Atoi(r.FormValue("numvar"))
	template_values.StoreRawData = r.FormValue("store_rawdata") == "true"
	template_values.SamplingRates = r.FormValue("rates")

	if err := tmpl.ExecuteTemplate(w, "validation_error_step8.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
