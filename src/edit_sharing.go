package randomization

import (
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// editSharing
func editSharing(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")
	shr := strings.Split(pkey, "::")
	owner := shr[0]
	projectName := shr[1]

	if strings.ToLower(owner) != strings.ToLower(user.String()) {
		msg := "Only the owner of a project can manage sharing."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	sharedUsers, err := getSharedUsers(ctx, pkey)
	if err != nil {
		sharedUsers = make([]string, 0)
		log.Infof(ctx, "editSharing failed to retrieve sharing: %v %v", projectName, owner)
	}

	tvals := struct {
		User           string
		LoggedIn       bool
		SharedUsers    []string
		AnySharedUsers bool
		ProjectName    string
		Pkey           string
	}{
		User:           user.String(),
		LoggedIn:       user != nil,
		SharedUsers:    sharedUsers,
		AnySharedUsers: len(sharedUsers) > 0,
		ProjectName:    projectName,
		Pkey:           pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "edit_sharing.html", tvals); err != nil {
		log.Errorf(ctx, "editSharing failed to execute template: %v", err)
	}
}

// editSharingConfirm
func editSharingConfirm(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	spkey := strings.Split(pkey, "::")
	projectName := spkey[1]

	ap := r.FormValue("additional_people")
	addUsers := []string{}
	addUsers = cleanSplit(ap, ",")
	for k, x := range addUsers {
		addUsers[k] = strings.ToLower(x)
	}

	// Gmail addresses don't use @gmail.com.
	invalidEmails := make([]string, 0)
	for k, x := range addUsers {
		uparts := strings.Split(x, "@")
		if len(uparts) != 2 {
			invalidEmails = append(invalidEmails, x)
		} else {
			if uparts[1] == "gmail.com" {
				addUsers[k] = uparts[0]
			}
		}
	}

	if len(invalidEmails) > 0 {
		msg := "The project was not shared because the following email addresses are not valid: "
		msg += strings.Join(invalidEmails, ", ") + "."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	var err error
	err = addSharing(ctx, pkey, addUsers)
	if err != nil {
		msg := "Datastore error: unable to update sharing information."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		log.Errorf(ctx, "editSharingConfirm [1]: %v", err)
		return
	}

	removeUsers := r.Form["remove_users"]
	err = removeSharing(ctx, pkey, removeUsers)
	if err != nil {
		msg := "Datastore error: unable to update sharing information."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		log.Errorf(ctx, "editSharingConfirm [2]: %v", err)
		return
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		ProjectName string
		Pkey        string
	}{
		User:        user.String(),
		LoggedIn:    user != nil,
		ProjectName: projectName,
		Pkey:        pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "edit_sharing_confirm.html", tvals); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
