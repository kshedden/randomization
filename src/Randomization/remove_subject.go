package randomization

import (
	"appengine"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// Remove_subject
func Remove_subject(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)
	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can remove treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	if PR.StoreRawData == false {
		Msg := "Subjects cannot be removed for a project in which the subject level data is not stored"
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	type TV struct {
		User                 string
		LoggedIn             bool
		Pkey                 string
		ProjectName          string
		Any_removed_subjects bool
		RemovedSubjects      string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Pkey = Pkey
	template_values.ProjectName = PR.Name

	if len(PR.RemovedSubjects) > 0 {
		template_values.Any_removed_subjects = true
		template_values.RemovedSubjects = strings.Join(PR.RemovedSubjects, ", ")
	} else {
		template_values.Any_removed_subjects = false
	}

	tmpl, err := template.ParseFiles("header.html",
		"remove_subject.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "remove_subject.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// Remove_subject_confirm
func Remove_subject_confirm(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	Pkey := r.FormValue("pkey")
	subject_id := r.FormValue("subject_id")

	type TV struct {
		User        string
		LoggedIn    bool
		Pkey        string
		SubjectId   string
		ProjectName string
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can remove treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	// Check if the subject has already been removed
	for _, s := range PR.RemovedSubjects {
		if s == subject_id {
			Msg := fmt.Sprintf("Subject '%s' has already been removed from the study.", subject_id)
			Return_msg := "Return to project"
			Message_page(w, r, user, Msg, Return_msg,
				"/project_dashboard?pkey="+Pkey)
			return
		}
	}

	// Check if the subject exists
	found := false
	for _, rec := range PR.RawData {
		if rec.SubjectId == subject_id {
			found = true
			break
		}
	}
	if found == false {
		Msg := fmt.Sprintf("There is no subject with id '%s' in the project.", subject_id)
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.SubjectId = subject_id
	template_values.Pkey = Pkey
	template_values.ProjectName = PR.Name

	tmpl, err := template.ParseFiles("header.html",
		"remove_subject_confirm.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "remove_subject_confirm.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// Remove_subject_completed
func Remove_subject_completed(w http.ResponseWriter,
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

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can remove treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	if PR.StoreRawData == false {
		Msg := "Subjects cannot be removed for a project in which the subject level data is not stored"
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			fmt.Sprintf("/project_dashboard?pkey=", Pkey))
		return
	}

	subject_id := r.FormValue("subject_id")
	found := false
	var remove_rec *DataRecord
	for _, rec := range PR.RawData {
		if rec.SubjectId == subject_id {
			rec.Included = false
			remove_rec = rec
			found = true
		}
	}
	PR.RemovedSubjects = append(PR.RemovedSubjects, subject_id)

	comment := new(Comment)
	comment.Person = user.String()
	comment.DateTime = time.Now()
	comment.Comment = []string{fmt.Sprintf("Subject '%s' removed from the project.", subject_id)}
	PR.Comments = append(PR.Comments, comment)

	if found == false {
		Msg := fmt.Sprintf("Unable to remove subject '%s' from the project.", subject_id)
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	Remove_from_aggregate(remove_rec, PR)
	PR.NumAssignments -= 1
	Store_project(PR, Pkey, &c)

	Msg := fmt.Sprintf("Subject '%s' has been removed from the study.", subject_id)
	Return_msg := "Return to project dashboard"
	Message_page(w, r, user, Msg, Return_msg,
		"/project_dashboard?pkey="+Pkey)
}
