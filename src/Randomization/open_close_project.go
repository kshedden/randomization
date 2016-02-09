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

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can open or close a project for enrollment."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
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
	template_values.Pkey = Pkey
	template_values.ProjectName = PR.Name
	template_values.GroupNames = PR.GroupNames
	template_values.Open = PR.Open

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

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if PR.Owner != user.String() {
		Msg := "Only the project owner can open or close a project for enrollment."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	status := r.FormValue("open")

	if status == "open" {
		Msg := fmt.Sprintf("The project \"%s\" is now open for enrollment.", PR.Name)
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		PR.Open = true
	} else {
		Msg := fmt.Sprintf("The project \"%s\" is now closed for enrollment.", PR.Name)
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		PR.Open = false
	}

	Store_project(PR, Pkey, &c)
}
