package randomization

import (
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

	_, projlist, err := GetProjects(user.String(), false, &c)
	if err != nil {
		msg := "A datastore error occured, your projects cannot be retrieved."
		c.Errorf("Delete_project_step1: %v", err)
		return_msg := "Return to dashboard"
		Message_page(w, r, user, msg, return_msg, "/dashboard")
		return
	}

	if len(projlist) == 0 {
		msg := "You are not the owner of any projects.  A project can only be deleted by its owner."
		return_msg := "Return to dashboard"
		Message_page(w, r, user, msg, return_msg, "/dashboard")
		return
	}

	type TV struct {
		User     string
		LoggedIn bool
		Proj     []*EncodedProjectView
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Proj = Format_encoded_projects(projlist)
	template_values.LoggedIn = user != nil

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

	pkey := r.FormValue("project_list")
	svec := strings.Split(pkey, "::")

	type TV struct {
		User        string
		LoggedIn    bool
		ProjectName string
		Pkey        string
		Nokey       bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Pkey = pkey
	if len(svec) >= 2 {
		template_values.ProjectName = svec[1]
	}
	template_values.Nokey = len(pkey) == 0

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
	pkey := r.FormValue("Pkey")

	if err := r.ParseForm(); err != nil {
		c.Errorf("Delete_project_step3 [1]: %v", err)
		ServeError(&c, w, err)
		return
	}

	// Delete the SharingByProject object, but first read the
	// users list from it so we can delete the project from their
	// SharingByUsers records.
	key := datastore.NewKey(c, "SharingByProject", pkey, 0, nil)
	var sbproj SharingByProject
	Shared_with := make([]string, 0)
	err := datastore.Get(c, key, &sbproj)
	if err == datastore.ErrNoSuchEntity {
		c.Errorf("Delete_project_step3 [2]: %v", err)
	} else if err != nil {
		c.Errorf("Delete_project_step3 [3] %v", err)
	} else {
		Shared_with = Clean_split(sbproj.Users, ",")
		err = datastore.Delete(c, key)
		if err != nil {
			c.Errorf("Delete_project_step3 [4] %v", err)
		}
	}

	// Delete the project.
	key = datastore.NewKey(c, "EncodedProject", pkey, 0, nil)
	err = datastore.Delete(c, key)
	if err != nil {
		c.Errorf("Delete_project_step3 [5]: %v", err)
	}

	// Delete from each user's SharingByUser record.
	for _, user1 := range Shared_with {
		var sbuser SharingByUser
		key := datastore.NewKey(c, "SharingByUser", strings.ToLower(user1), 0, nil)
		err := datastore.Get(c, key, &sbuser)
		if err != nil {
			c.Errorf("Delete_project_step3 [6]: %v", err)
		}
		Projects := Clean_split(sbuser.Projects, ",")

		// Get the unique project keys, except for pkey.
		mp := make(map[string]bool)
		for _, x := range Projects {
			if x != pkey {
				mp[x] = true
			}
		}
		vec := make([]string, len(mp))
		jj := 0
		for k, _ := range mp {
			vec[jj] = k
			jj += 1
		}
		sbuser.Projects = strings.Join(vec, ",")

		_, err = datastore.Put(c, key, &sbuser)
		if err != nil {
			c.Errorf("Delete_project_step3 [7]: %v", err)
		}
	}

	type TV struct {
		User     string
		LoggedIn bool
		Success  bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Success = err == nil
	template_values.LoggedIn = user != nil

	tmpl, err := template.ParseFiles("header.html",
		"delete_project_step3.html")
	if err != nil {
		c.Errorf("Delete_project_step3 [8]: %v", err)
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "delete_project_step3.html",
		template_values); err != nil {
		c.Errorf("Delete_project_step3 [9]: %v", err)
	}
}
