package randomization

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func copyProject(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	ok := checkAccess(user, pkey, ctx, &w, r)

	if !ok {
		msg := "Only the project owner can copy a project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	key := datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	var eproj EncodedProject
	err := datastore.Get(ctx, key, &eproj)
	if err != nil {
		log.Errorf(ctx, "Copy_project: %v", err)
		msg := "Unknown datastore error."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	type TV struct {
		User        string
		LoggedIn    bool
		Pkey        string
		ProjectName string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Pkey = pkey
	template_values.ProjectName = eproj.Name

	if err := tmpl.ExecuteTemplate(w, "copy_project.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

func copyProjectCompleted(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	ok := checkAccess(user, pkey, ctx, &w, r)

	if !ok {
		msg := "You do not have access to the requested project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	key := datastore.NewKey(ctx, "EncodedProject", pkey, 0, nil)
	var eproj EncodedProject
	err := datastore.Get(ctx, key, &eproj)
	if err != nil {
		msg := "Unknown error, the project was not copied."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		log.Errorf(ctx, "Copy_project: %v", err)
		return
	}

	eproj_copy := copyEncodedProject(&eproj)

	// Check if the name is valid (not blank)
	new_name := r.FormValue("new_project_name")
	new_name = strings.TrimSpace(new_name)
	if len(new_name) == 0 {
		msg := "A name for the new project must be provided."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}
	eproj_copy.Name = new_name

	// The owner of the copied project is the current user
	eproj_copy.Owner = user.String()

	// Check if the project name has already been used.
	new_pkey := user.String() + "::" + new_name
	new_key := datastore.NewKey(ctx, "EncodedProject", new_pkey, 0, nil)
	var pr EncodedProject
	err = datastore.Get(ctx, new_key, &pr)
	if err == nil {
		msg := fmt.Sprintf("A project named \"%s\" belonging to user %s already exists.", new_name, user.String())
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	_, err = datastore.Put(ctx, new_key, eproj_copy)
	if err != nil {
		log.Errorf(ctx, "Copy_project: %v", err)
		msg := "Unknown error, the project was not copied."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	log.Infof(ctx, "Copied %s to %s", pkey, new_pkey)
	msg := "The project has been successfully copied."
	rmsg := "Return to dashboard"
	messagePage(w, r, user, msg, rmsg, "/dashboard")
	return
}
