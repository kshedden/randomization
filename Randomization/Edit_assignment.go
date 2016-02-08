package randomization

import (
	"fmt"
	"time"
	//	"strconv"
	"appengine/user"
	"net/http"
	//	"appengine/datastore"
	"appengine"
	"html/template"
	//"strings"
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

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can edit treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	if PR.Store_RawData == false {
		Msg := "Group assignments cannot be edited for a project in which the subject level data is not stored"
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	type TV struct {
		User         string
		Logged_in    bool
		Pkey         string
		Project_name string
		Group_names  []string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Logged_in = user != nil
	template_values.Pkey = Pkey
	template_values.Project_name = PR.Name
	template_values.Group_names = PR.Group_names

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

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can edit treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	if PR.Store_RawData == false {
		Msg := "Assignments cannot be edited in a project in which the subject level data is not stored"
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	subject_id := r.FormValue("subject_id")

	type TV struct {
		User               string
		Logged_in          bool
		Pkey               string
		Project_name       string
		Current_group_name string
		New_group_name     string
		Subject_id         string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Logged_in = user != nil
	template_values.Pkey = Pkey
	template_values.Project_name = PR.Name
	template_values.New_group_name = r.FormValue("new_group_name")
	template_values.Subject_id = subject_id

	found := false
	for _, rec := range PR.RawData {
		if rec.Subject_id == subject_id {
			template_values.Current_group_name = rec.Current_group
			found = true
		}
	}
	if !found {
		Msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subject_id)
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	if template_values.Current_group_name == template_values.New_group_name {
		Msg := fmt.Sprintf("You have requested to change the treatment group of subject '%s' to '%s', but the subject is already in this treatment group.", subject_id, template_values.New_group_name)
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
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

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can edit treatment group assignments that have already been made."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	if PR.Store_RawData == false {
		Msg := "Group assignments cannot be edited in a project in which the subject level data is not stored."
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	new_group_name := r.FormValue("new_group_name")
	subject_id := r.FormValue("subject_id")

	found := false
	for _, rec := range PR.RawData {
		if rec.Subject_id == subject_id {
			Remove_from_aggregate(rec, PR)
			old_group_name := rec.Current_group
			rec.Current_group = new_group_name
			Add_to_aggregate(rec, PR)

			comment := new(Comment)
			comment.Person = user.String()
			comment.DateTime = time.Now()
			comment.Comment = []string{
				fmt.Sprintf("Group assignment for subject '%s' changed from '%s' to '%s'",
					subject_id, old_group_name, new_group_name)}
			PR.Comments = append(PR.Comments, comment)

			found = true
		}
	}
	if !found {
		Msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subject_id)
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	Store_project(PR, Pkey, &c)

	Msg := "The assignment has been changed."
	Return_msg := "Return to project"
	Message_page(w, r, user, Msg, Return_msg,
		"/project_dashboard?pkey="+Pkey)
}
