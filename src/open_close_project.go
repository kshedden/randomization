package randomization

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// openClose_project
func openCloseProject(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
		msg := "You do not have access to this page."
		rmsg := "Return"
		messagePage(w, r, user, msg, rmsg, "/")
		return
	}

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		log.Errorf(ctx, "OpenClose_project [1]: %v", err)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can open or close a project for enrollment."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
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

	if err := tmpl.ExecuteTemplate(w, "openclose_project.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// openCloseCompleted
func openCloseCompleted(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
		msg := "You do not have access to this page."
		rmsg := "Return"
		messagePage(w, r, user, msg, rmsg, "/")
		return
	}

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can open or close a project for enrollment."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	status := r.FormValue("open")

	if status == "open" {
		msg := fmt.Sprintf("The project \"%s\" is now open for enrollment.", proj.Name)
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		proj.Open = true
	} else {
		msg := fmt.Sprintf("The project \"%s\" is now closed for enrollment.", proj.Name)
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		proj.Open = false
	}

	storeProject(ctx, proj, pkey)
}
