package randomization

import (
	"appengine"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// Edit_assignment
func Edit_assignment(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)
	pkey := r.FormValue("pkey")

	if ok := Check_access(user, pkey, &c, &w, r); !ok {
		Msg := "Only the project owner can edit treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	proj, _ := Get_project_from_key(pkey, &c)
	if proj.Owner != user.String() {
		Msg := "Only the project owner can edit treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	if proj.NumAssignments == 0 {
		Msg := "There are no assignments to edit."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	if proj.StoreRawData == false {
		Msg := "Group assignments cannot be edited for a project in which the subject level data is not stored"
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	type TV struct {
		User        string
		LoggedIn    bool
		Pkey        string
		ProjectName string
		GroupNames  []string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Pkey = pkey
	template_values.ProjectName = proj.Name
	template_values.GroupNames = proj.GroupNames

	tmpl, err := template.ParseFiles("header.html",
		"edit_assignment.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "edit_assignment.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// Edit_assignment_confirm
func Edit_assignment_confirm(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	pkey := r.FormValue("pkey")

	if ok := Check_access(user, pkey, &c, &w, r); !ok {
		return
	}

	proj, _ := Get_project_from_key(pkey, &c)

	if proj.Owner != user.String() {
		Msg := "Only the project owner can edit treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	if proj.StoreRawData == false {
		Msg := "Assignments cannot be edited in a project in which the subject level data is not stored"
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	subject_id := r.FormValue("subject_id")

	type TV struct {
		User             string
		LoggedIn         bool
		Pkey             string
		ProjectName      string
		CurrentGroupName string
		NewGroupName     string
		SubjectId        string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Pkey = pkey
	template_values.ProjectName = proj.Name
	template_values.NewGroupName = r.FormValue("new_group_name")
	template_values.SubjectId = subject_id

	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subject_id {
			template_values.CurrentGroupName = rec.CurrentGroup
			found = true
		}
	}
	if !found {
		Msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subject_id)
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	if template_values.CurrentGroupName == template_values.NewGroupName {
		Msg := fmt.Sprintf("You have requested to change the treatment group of subject '%s' to '%s', but the subject is already in this treatment group.", subject_id, template_values.NewGroupName)
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	tmpl, err := template.ParseFiles("header.html",
		"edit_assignment_confirm.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "edit_assignment_confirm.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// Edit_assignment_completed
func Edit_assignment_completed(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	pkey := r.FormValue("pkey")

	if ok := Check_access(user, pkey, &c, &w, r); !ok {
		return
	}

	proj, _ := Get_project_from_key(pkey, &c)

	if proj.Owner != user.String() {
		Msg := "Only the project owner can edit treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	if proj.StoreRawData == false {
		Msg := "Group assignments cannot be edited in a project in which the subject level data is not stored."
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	new_group_name := r.FormValue("new_group_name")
	subject_id := r.FormValue("subject_id")

	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subject_id {
			Remove_from_aggregate(rec, proj)
			old_group_name := rec.CurrentGroup
			rec.CurrentGroup = new_group_name
			Add_to_aggregate(rec, proj)

			comment := new(Comment)
			comment.Person = user.String()
			comment.DateTime = time.Now()
			comment.Comment = []string{
				fmt.Sprintf("Group assignment for subject '%s' changed from '%s' to '%s'",
					subject_id, old_group_name, new_group_name)}
			proj.Comments = append(proj.Comments, comment)

			found = true
		}
	}
	if !found {
		Msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subject_id)
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	Store_project(proj, pkey, &c)

	Msg := "The assignment has been changed."
	Return_msg := "Return to project"
	Message_page(w, r, user, Msg, Return_msg,
		"/project_dashboard?pkey="+pkey)
}
