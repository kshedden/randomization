package randomization

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	//"strconv"
	"strings"
	//"time"
)

func Copy_project(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	pkey := r.FormValue("pkey")

	ok := Check_access(user, pkey, &c, &w, r)

	if !ok {
		Msg := "Only the project owner can copy a project."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	key := datastore.NewKey(c, "Encoded_Project", pkey, 0, nil)
	var eproj Encoded_Project
	err := datastore.Get(c, key, &eproj)
	if err != nil {
		c.Errorf("Copy_project: %v", err)
		Msg := "Unknown datastore error."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	type TV struct {
		User         string
		Logged_in    bool
		Pkey         string
		Project_name string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Logged_in = user != nil
	template_values.Pkey = pkey
	template_values.Project_name = eproj.Name

	tmpl, err := template.ParseFiles("header.html", "copy_project.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "copy_project.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

func Copy_project_completed(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	pkey := r.FormValue("pkey")

	ok := Check_access(user, pkey, &c, &w, r)

	if !ok {
		Msg := "You do not have access to the requested project."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}

	key := datastore.NewKey(c, "Encoded_Project", pkey, 0, nil)
	var eproj Encoded_Project
	err := datastore.Get(c, key, &eproj)
	if err != nil {
		Msg := "Unknown error, the project was not copied."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		c.Errorf("Copy_project: %v", err)
		return
	}

	eproj_copy := Copy_encoded_project(&eproj)

	// Check if the name is valid (not blank)
	new_name := r.FormValue("new_project_name")
	new_name = strings.TrimSpace(new_name)
	if len(new_name) == 0 {
		Msg := "A name for the new project must be provided."
		Return_msg := "Return to project dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+pkey)
		return
	}
	eproj_copy.Name = new_name

	// The owner of the copied project is the current user
	eproj_copy.Owner = user.String()

	// Check if the project name has already been used.
	new_pkey := user.String() + "::" + new_name
	new_key := datastore.NewKey(c, "Encoded_Project", new_pkey, 0, nil)
	var pr Encoded_Project
	err = datastore.Get(c, new_key, &pr)
	if err == nil {
		Msg := fmt.Sprintf("A project named \"%s\" belonging to user %s already exists.", new_name,
			user.String())
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	_, err = datastore.Put(c, new_key, eproj_copy)
	if err != nil {
		c.Errorf("Copy_project: %v", err)
		Msg := "Unknown error, the project was not copied."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	c.Infof("Copied %s to %s", pkey, new_pkey)
	Msg := "The project has been successfully copied."
	Return_msg := "Return to dashboard"
	Message_page(w, r, user, Msg, Return_msg, "/dashboard")
	return
}
