package randomization

import (
	"appengine"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
)

// OpenClose_project
func OpenClose_project(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)
	pkey := r.FormValue("pkey")

	if ok := Check_access(user, pkey, &c, &w, r); !ok {
		return
	}

	proj, err := Get_project_from_key(pkey, &c)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		return_msg := "Return to project dashboard"
		Message_page(w, r, user, msg, return_msg,
			"/project_dashboard?pkey="+pkey)
		c.Errorf("OpenClose_project [1]: %v", err)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can open or close a project for enrollment."
		return_msg := "Return to project dashboard"
		Message_page(w, r, user, msg, return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	type TV struct {
		User        string
		LoggedIn    bool
		Pkey        string
		ProjectName string
		GroupNames  []string
		Open        bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Pkey = pkey
	template_values.ProjectName = proj.Name
	template_values.GroupNames = proj.GroupNames
	template_values.Open = proj.Open

	tmpl, err := template.ParseFiles("header.html",
		"openclose_project.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "openclose_project.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// OpenClose_completed
func OpenClose_completed(w http.ResponseWriter,
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
		msg := "Only the project owner can open or close a project for enrollment."
		return_msg := "Return to project dashboard"
		Message_page(w, r, user, msg, return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	status := r.FormValue("open")

	if status == "open" {
		msg := fmt.Sprintf("The project \"%s\" is now open for enrollment.", proj.Name)
		return_msg := "Return to project dashboard"
		Message_page(w, r, user, msg, return_msg,
			"/project_dashboard?pkey="+pkey)
		proj.Open = true
	} else {
		msg := fmt.Sprintf("The project \"%s\" is now closed for enrollment.", proj.Name)
		return_msg := "Return to project dashboard"
		Message_page(w, r, user, msg, return_msg,
			"/project_dashboard?pkey="+pkey)
		proj.Open = false
	}

	Store_project(proj, pkey, &c)
}
