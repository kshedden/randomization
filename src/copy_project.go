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

	ok := checkAccess(ctx, user, pkey, &w, r)

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

	tvals := struct {
		User        string
		LoggedIn    bool
		Pkey        string
		ProjectName string
	}{
		User:        user.String(),
		LoggedIn:    user != nil,
		Pkey:        pkey,
		ProjectName: eproj.Name,
	}

	if err := tmpl.ExecuteTemplate(w, "copy_project.html", tvals); err != nil {
		log.Errorf(ctx, "copyProject failed to execute template: %v", err)
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

	ok := checkAccess(ctx, user, pkey, &w, r)

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

	eprojCopy := copyEncodedProject(&eproj)

	// Check if the name is valid (not blank)
	newName := r.FormValue("new_project_name")
	newName = strings.TrimSpace(newName)
	if len(newName) == 0 {
		msg := "A name for the new project must be provided."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}
	eprojCopy.Name = newName

	// The owner of the copied project is the current user
	eprojCopy.Owner = user.String()

	// Check if the project name has already been used.
	newPkey := user.String() + "::" + newName
	newKey := datastore.NewKey(ctx, "EncodedProject", newPkey, 0, nil)
	var pr EncodedProject
	err = datastore.Get(ctx, newKey, &pr)
	if err == nil {
		msg := fmt.Sprintf("A project named \"%s\" belonging to user %s already exists.", newName, user.String())
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	_, err = datastore.Put(ctx, newKey, eprojCopy)
	if err != nil {
		log.Errorf(ctx, "Copy_project: %v", err)
		msg := "Unknown error, the project was not copied."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	log.Infof(ctx, "Copied %s to %s", pkey, newPkey)
	msg := "The project has been successfully copied."
	rmsg := "Return to dashboard"
	messagePage(w, r, user, msg, rmsg, "/dashboard")
}
