package randomization

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// Assign_treatment_input
func Assign_treatment_input(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	if err := r.ParseForm(); err != nil {
		ServeError(&c, w, err)
		return
	}

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, err := Get_project_from_key(Pkey, &c)
	if err != nil {
		c.Errorf("Assign_treatment_input: %v", err)
		Msg := "A datastore error occured, the project could not be loaded."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	if PR.Open == false {
		Msg := "This project is currently not open for new enrollments.  The project owner can change this by following the \"Open/close enrollment\" link on the project dashboard."
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	PV := Format_project(PR)

	type TV struct {
		User      string
		LoggedIn  bool
		PR        *Project
		PV        *Project_view
		NumGroups int
		Fields    string
		Pkey      string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.PR = PR
	template_values.PV = PV
	template_values.NumGroups = len(PR.GroupNames)
	template_values.Pkey = Pkey

	S := make([]string, len(PR.Variables))
	for i, v := range PR.Variables {
		S[i] = v.Name
	}
	template_values.Fields = strings.Join(S, ",")

	tmpl, err := template.ParseFiles("header.html",
		"assign_treatment_input.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "assign_treatment_input.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

func check_before_assigning(proj *Project,
	pkey string,
	subject_id string,
	user *user.User,
	w http.ResponseWriter,
	r *http.Request) bool {

	if proj.Open == false {
		Msg := "This project is currently not open for new enrollments.  The project owner can change this by following the \"Open/close enrollment\" link on the project dashboard."
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return false
	}

	// Check the subject id
	if proj.StoreRawData {

		if len(subject_id) == 0 {
			Msg := fmt.Sprintf("The subject id may not be blank.")
			Return_msg := "Return to project"
			Message_page(w, r, user, Msg, Return_msg,
				"/project_dashboard?pkey="+pkey)
			return false
		}

		for _, rec := range proj.RawData {
			if subject_id == rec.SubjectId {
				Msg := fmt.Sprintf("Subject '%s' has already been assigned to a treatment group.  Please use a different subject id.", subject_id)
				Return_msg := "Return to project"
				Message_page(w, r, user, Msg, Return_msg,
					"/project_dashboard?pkey="+pkey)
				return false
			}
		}
	}

	return true
}

// Assign_treatment_confirm
func Assign_treatment_confirm(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		ServeError(&c, w, err)
		return
	}

	subject_id := r.FormValue("subject_id")
	subject_id = strings.TrimSpace(subject_id)

	project, err := Get_project_from_key(Pkey, &c)
	if err != nil {
		c.Errorf("Assign_treatment_confirm: %v", err)
		Msg := "A datastore error occured, the project could not be loaded."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	ok := check_before_assigning(project, Pkey, subject_id, user, w, r)
	if !ok {
		return
	}

	project_view := Format_project(project)

	Fields := strings.Split(r.FormValue("fields"), ",")
	FV := make([][]string, len(Fields)+1)
	Values := make([]string, len(Fields))

	FV[0] = []string{"Subject id", subject_id}
	for i, v := range Fields {
		x := r.FormValue(v)
		FV[i+1] = []string{v, x}
		Values[i] = x
	}

	type TV struct {
		User         string
		LoggedIn     bool
		Pkey         string
		Project      *Project
		Project_view *Project_view
		NumGroups    int
		Fields       string
		FV           [][]string
		Values       string
		SubjectId    string
		Any_vars     bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Pkey = Pkey
	template_values.Project = project
	template_values.Project_view = project_view
	template_values.NumGroups = len(project.GroupNames)
	template_values.Fields = strings.Join(Fields, ",")
	template_values.FV = FV
	template_values.Values = strings.Join(Values, ",")
	template_values.SubjectId = subject_id
	template_values.Any_vars = len(project.Variables) > 0

	tmpl, err := template.ParseFiles("header.html",
		"assign_treatment_confirm.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "assign_treatment_confirm.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// Assign_treatment
func Assign_treatment(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		ServeError(&c, w, err)
		return
	}

	PR, err := Get_project_from_key(Pkey, &c)
	if err != nil {
		c.Errorf("Assign_treatment %v", err)
		Msg := "A datastore error occured, the project could not be loaded."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	subject_id := r.FormValue("subject_id")

	// Check this a second time in case someone lands on this page
	// without going through the previous checks
	// (e.g. inappropriate use of back button on browser).
	ok := check_before_assigning(PR, Pkey, subject_id, user, w, r)
	if !ok {
		return
	}

	PV := Format_project(PR)

	Fields := strings.Split(r.FormValue("fields"), ",")
	Values := strings.Split(r.FormValue("values"), ",")

	// M maps variable names to values for the unit that is about
	// to be randomized to a treatment group.
	M := make(map[string]string)
	for i, x := range Fields {
		M[x] = Values[i]
	}

	ax, msg, err := Do_assignment(&M, PR, subject_id, user.String())
	if err != nil {
		c.Errorf("%v", err)
	}
	c.Infof("%v", msg)

	PR.Modified = time.Now()

	// Update the project in the database.
	EP, _ := Encode_Project(PR)
	Key := datastore.NewKey(c, "EncodedProject", Pkey, 0, nil)
	_, err = datastore.Put(c, Key, EP)
	if err != nil {
		c.Errorf("Assign_treatment: %v", err)
		Msg := "A datastore error occured, the project could not be updated."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	type TV struct {
		User      string
		LoggedIn  bool
		PR        *Project
		PV        *Project_view
		NumGroups int
		Ax        string
		Pkey      string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Ax = ax
	template_values.PR = PR
	template_values.PV = PV
	template_values.NumGroups = len(PR.GroupNames)
	template_values.Pkey = Pkey

	tmpl, err := template.ParseFiles("header.html",
		"assign_treatment.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "assign_treatment.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}
