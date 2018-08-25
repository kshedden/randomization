package randomization

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// assignTreatmentInput
func assignTreatmentInput(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	pkey := r.FormValue("pkey")

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		return
	}

	PR, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		log.Errorf(ctx, "Assign_treatment_input: %v", err)
		msg := "A datastore error occured, the project could not be loaded."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	if PR.Open == false {
		msg := "This project is currently not open for new enrollments.  The project owner can change this by following the \"Open/close enrollment\" link on the project dashboard."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	PV := formatProject(PR)

	tvals := struct {
		User      string
		LoggedIn  bool
		PR        *Project
		PV        *ProjectView
		NumGroups int
		Fields    string
		Pkey      string
	}{
		User:      user.String(),
		LoggedIn:  user != nil,
		PR:        PR,
		PV:        PV,
		NumGroups: len(PR.GroupNames),
		Pkey:      pkey,
	}

	S := make([]string, len(PR.Variables))
	for i, v := range PR.Variables {
		S[i] = v.Name
	}
	tvals.Fields = strings.Join(S, ",")

	if err := tmpl.ExecuteTemplate(w, "assign_treatment_input.html", tvals); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

func checkBeforeAssigning(proj *Project, pkey string, subjectId string, user *user.User, w http.ResponseWriter, r *http.Request) bool {

	if proj.Open == false {
		msg := "This project is currently not open for new enrollments.  The project owner can change this by following the \"Open/close enrollment\" link on the project dashboard."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return false
	}

	// Check the subject id
	if proj.StoreRawData {

		if len(subjectId) == 0 {
			msg := fmt.Sprintf("The subject id may not be blank.")
			rmsg := "Return to project"
			messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
			return false
		}

		for _, rec := range proj.RawData {
			if subjectId == rec.SubjectId {
				msg := fmt.Sprintf("Subject '%s' has already been assigned to a treatment group.  Please use a different subject id.", subjectId)
				rmsg := "Return to project"
				messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
				return false
			}
		}
	}

	return true
}

// assignTreatmentConfirm
func assignTreatmentConfirm(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	pkey := r.FormValue("pkey")

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	subjectId := r.FormValue("subject_id")
	subjectId = strings.TrimSpace(subjectId)

	project, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		log.Errorf(ctx, "Assign_treatment_confirm: %v", err)
		msg := "A datastore error occured, the project could not be loaded."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	ok := checkBeforeAssigning(project, pkey, subjectId, user, w, r)
	if !ok {
		return
	}

	projView := formatProject(project)

	Fields := strings.Split(r.FormValue("fields"), ",")
	FV := make([][]string, len(Fields)+1)
	Values := make([]string, len(Fields))

	FV[0] = []string{"Subject id", subjectId}
	for i, v := range Fields {
		x := r.FormValue(v)
		FV[i+1] = []string{v, x}
		Values[i] = x
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Pkey        string
		Project     *Project
		ProjectView *ProjectView
		NumGroups   int
		Fields      string
		FV          [][]string
		Values      string
		SubjectId   string
		AnyVars     bool
	}{
		User:        user.String(),
		LoggedIn:    user != nil,
		Pkey:        pkey,
		Project:     project,
		ProjectView: projView,
		NumGroups:   len(project.GroupNames),
		Fields:      strings.Join(Fields, ","),
		FV:          FV,
		Values:      strings.Join(Values, ","),
		SubjectId:   subjectId,
		AnyVars:     len(project.Variables) > 0,
	}

	if err := tmpl.ExecuteTemplate(w, "assign_treatment_confirm.html", tvals); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// assignTreatment
func assignTreatment(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		log.Errorf(ctx, "Assign_treatment %v", err)
		msg := "A datastore error occured, the project could not be loaded."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	subjectId := r.FormValue("subject_id")

	// Check this a second time in case someone lands on this page
	// without going through the previous checks
	// (e.g. inappropriate use of back button on browser).
	ok := checkBeforeAssigning(proj, pkey, subjectId, user, w, r)
	if !ok {
		return
	}

	pview := formatProject(proj)

	fields := strings.Split(r.FormValue("fields"), ",")
	values := strings.Split(r.FormValue("values"), ",")

	// mpv maps variable names to values for the unit that is about
	// to be randomized to a treatment group.
	mpv := make(map[string]string)
	for i, x := range fields {
		mpv[x] = values[i]
	}

	ax, err := doAssignment(&mpv, proj, subjectId, user.String())
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}

	proj.Modified = time.Now()

	// Update the project in the database.
	eproj, _ := encodeProject(proj)
	key := datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	_, err = datastore.Put(ctx, key, eproj)
	if err != nil {
		log.Errorf(ctx, "Assign_treatment: %v", err)
		msg := "A datastore error occured, the project could not be updated."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	tvals := struct {
		User      string
		LoggedIn  bool
		PR        *Project
		PV        *ProjectView
		NumGroups int
		Ax        string
		Pkey      string
	}{
		User:      user.String(),
		LoggedIn:  user != nil,
		Ax:        ax,
		PR:        proj,
		PV:        pview,
		NumGroups: len(proj.GroupNames),
		Pkey:      pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "assign_treatment.html", tvals); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
