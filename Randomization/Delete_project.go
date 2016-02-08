package randomization

import (
	//	"fmt"
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"html/template"
	"net/http"
	"strings"
)

// Create_project_step1 gets the project name from the user.
func Delete_project_step1(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	_, PR, err := Get_projects(user.String(), false, &c)
	if err != nil {
		Msg := "A datastore error occured, your projects cannot be retrieved."
		c.Errorf("Delete_project_step1: %v", err)
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	if len(PR) == 0 {
		Msg := "You are not the owner of any projects.  A project can only be deleted by its owner."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	type TV struct {
		User      string
		Logged_in bool
		PRN       bool
		PR        []*Encoded_Project_view
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.PR = Format_encoded_projects(PR)
	template_values.PRN = len(PR) > 0
	template_values.Logged_in = user != nil

	tmpl, err := template.ParseFiles("header.html",
		"delete_project_step1.html")
	if err != nil {
		c.Errorf("Delete_project_step1: %v", err)
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "delete_project_step1.html",
		template_values); err != nil {
		c.Errorf("Delete_project_step1: %v", err)
	}
}

// Delete_project_step2 confirms that a project should be deleted.
func Delete_project_step2(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	if err := r.ParseForm(); err != nil {
		ServeError(&c, w, err)
		return
	}

	Pkey := r.FormValue("project_list")
	S := strings.Split(Pkey, "::")

	type TV struct {
		User         string
		Logged_in    bool
		Project_name string
		Pkey         string
		Nokey        bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Logged_in = user != nil
	template_values.Pkey = Pkey
	if len(S) >= 2 {
		template_values.Project_name = S[1]
	}
	template_values.Nokey = len(Pkey) == 0

	tmpl, err := template.ParseFiles("header.html",
		"delete_project_step2.html")
	if err != nil {
		c.Errorf("Delete_project_step2: %v", err)
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "delete_project_step2.html",
		template_values); err != nil {
		c.Errorf("Delete_project_step2: %v", err)
	}
}

// Delete_project_step3 deletes a project.
func Delete_project_step3(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	Pkey := r.FormValue("Pkey")

	if err := r.ParseForm(); err != nil {
		c.Errorf("Delete_proect_step3 (1): %v", err)
		ServeError(&c, w, err)
		return
	}

	// Delete the Sharing_by_project object, but first read the
	// users list from it so we can delete the project from their
	// Sharing_by_users records.
	key := datastore.NewKey(c, "Sharing_by_project", Pkey, 0, nil)
	var SBP Sharing_by_project
	err := datastore.Get(c, key, &SBP)
	if err != nil && err != datastore.ErrNoSuchEntity {
		Msg := "A datastore error occured, the project may not have been deleted."
		c.Errorf("Delete_project_step3 (2): %v", err)
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}
	Shared_with := make([]string, 0)
	if err != datastore.ErrNoSuchEntity {
		Shared_with = Clean_split(SBP.Users, ",")
		err = datastore.Delete(c, key)
		if err != nil {
			Msg := "A datastore error occured, the project may not have been deleted."
			c.Errorf("Delete_project_step3 (3): %v", err)
			Return_msg := "Return to dashboard"
			Message_page(w, r, user, Msg, Return_msg, "/dashboard")
			return
		}
	}

	// Delete the project.
	key = datastore.NewKey(c, "Encoded_Project", Pkey, 0, nil)
	err = datastore.Delete(c, key)
	if err != nil {
		Msg := "A datastore error occured, the project may not have been deleted."
		c.Errorf("Delete_project_step3 (4): %v", err)
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	// Delete from each user's Sharing_by_user record.
	for _, user1 := range Shared_with {
		var SBU Sharing_by_user
		key := datastore.NewKey(c, "Sharing_by_user",
			strings.ToLower(user1), 0, nil)
		err := datastore.Get(c, key, &SBU)
		if err != nil {
			Msg := "A datastore error occured, the project may not have been deleted."
			c.Errorf("Delete_project_step3 (5): %v", err)
			Return_msg := "Return to dashboard"
			Message_page(w, r, user, Msg, Return_msg, "/dashboard")
			return
		}

		Projects := Clean_split(SBU.Projects, ",")

		// Get the unique project keys, except for Pkey.
		H := make(map[string]bool)
		for _, x := range Projects {
			if x != Pkey {
				H[x] = true
			}
		}
		V := make([]string, len(H))
		jj := 0
		for k, _ := range H {
			V[jj] = k
			jj += 1
		}
		SBU.Projects = strings.Join(V, ",")

		_, err = datastore.Put(c, key, &SBU)
		if err != nil {
			Msg := "A datastore error occured, the project may not have been deleted."
			c.Errorf("Delete_project_step3 (6): %v", err)
			Return_msg := "Return to dashboard"
			Message_page(w, r, user, Msg, Return_msg, "/dashboard")
			return
		}
	}

	type TV struct {
		User      string
		Logged_in bool
		Success   bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Success = err == nil
	template_values.Logged_in = user != nil

	tmpl, err := template.ParseFiles("header.html",
		"delete_project_step3.html")
	if err != nil {
		c.Errorf("Delete_project_step3 (7): %v", err)
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "delete_project_step3.html",
		template_values); err != nil {
		c.Errorf("Delete_project_step3 (8): %v", err)
	}
}
