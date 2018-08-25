package randomization

import (
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// deleteProjectStep1 gets the project name from the user.
func deleteProjectStep1(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	_, projlist, err := getProjects(ctx, user.String(), false)
	if err != nil {
		msg := "A datastore error occured, your projects cannot be retrieved."
		log.Errorf(ctx, "Delete_project_step1: %v", err)
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	if len(projlist) == 0 {
		msg := "You are not the owner of any projects.  A project can only be deleted by its owner."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	type TV struct {
		User     string
		LoggedIn bool
		Proj     []*EncodedProjectView
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Proj = formatEncodedProjects(projlist)
	template_values.LoggedIn = user != nil

	if err := tmpl.ExecuteTemplate(w, "delete_project_step1.html", template_values); err != nil {
		log.Errorf(ctx, "Delete_project_step1: %v", err)
	}
}

// deleteProjectStep2 confirms that a project should be deleted.
func deleteProjectStep2(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)

	user := user.Current(ctx)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
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

	if err := tmpl.ExecuteTemplate(w, "delete_project_step2.html",
		template_values); err != nil {
		log.Errorf(ctx, "Delete_project_step2: %v", err)
	}
}

// deleteProjectStep3 deletes a project.
func deleteProjectStep3(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("Pkey")

	if err := r.ParseForm(); err != nil {
		log.Errorf(ctx, "Delete_project_step3 [1]: %v", err)
		ServeError(ctx, w, err)
		return
	}

	// Delete the SharingByProject object, but first read the
	// users list from it so we can delete the project from their
	// SharingByUsers records.
	key := datastore.NewKey(ctx, "SharingByProject", pkey, 0, nil)
	var sbproj SharingByProject
	Shared_with := make([]string, 0)
	err := datastore.Get(ctx, key, &sbproj)
	if err == datastore.ErrNoSuchEntity {
		log.Errorf(ctx, "Delete_project_step3 [2]: %v", err)
	} else if err != nil {
		log.Errorf(ctx, "Delete_project_step3 [3] %v", err)
	} else {
		Shared_with = cleanSplit(sbproj.Users, ",")
		err = datastore.Delete(ctx, key)
		if err != nil {
			log.Errorf(ctx, "Delete_project_step3 [4] %v", err)
		}
	}

	// Delete the project.
	key = datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	err = datastore.Delete(ctx, key)
	if err != nil {
		log.Errorf(ctx, "Delete_project_step3 [5]: %v", err)
	}

	// Delete from each user's SharingByUser record.
	for _, user1 := range Shared_with {
		var sbuser SharingByUser
		key := datastore.NewKey(ctx, "SharingByUser", strings.ToLower(user1), 0, nil)
		err := datastore.Get(ctx, key, &sbuser)
		if err != nil {
			log.Errorf(ctx, "Delete_project_step3 [6]: %v", err)
		}
		Projects := cleanSplit(sbuser.Projects, ",")

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

		_, err = datastore.Put(ctx, key, &sbuser)
		if err != nil {
			log.Errorf(ctx, "Delete_project_step3 [7]: %v", err)
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

	if err := tmpl.ExecuteTemplate(w, "delete_project_step3.html", template_values); err != nil {
		log.Errorf(ctx, "Delete_project_step3 [9]: %v", err)
	}
}
