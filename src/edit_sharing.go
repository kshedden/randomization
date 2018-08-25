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
	project_name := shr[1]

	if strings.ToLower(owner) != strings.ToLower(user.String()) {
		msg := "Only the owner of a project can manage sharing."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		return
	}

	shared_users, err := getSharedUsers(ctx, pkey)
	if err != nil {
		shared_users = make([]string, 0)
		log.Infof(ctx, "Failed to retrieve sharing: %v %v", project_name, owner)
	}

	type TV struct {
		User           string
		LoggedIn       bool
		SharedUsers    []string
		AnySharedUsers bool
		ProjectName    string
		Pkey           string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.SharedUsers = shared_users
	template_values.AnySharedUsers = len(shared_users) > 0
	template_values.ProjectName = project_name
	template_values.Pkey = pkey

	if err := tmpl.ExecuteTemplate(w, "edit_sharing.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
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
	project_name := spkey[1]

	ap := r.FormValue("additional_people")
	add_users := []string{}
	add_users = cleanSplit(ap, ",")
	for k, x := range add_users {
		add_users[k] = strings.ToLower(x)
	}

	// Gmail addresses don't use @gmail.com.
	invalid_emails := make([]string, 0)
	for k, x := range add_users {
		uparts := strings.Split(x, "@")
		if len(uparts) != 2 {
			invalid_emails = append(invalid_emails, x)
		} else {
			if uparts[1] == "gmail.com" {
				add_users[k] = uparts[0]
			}
		}
	}

	if len(invalid_emails) > 0 {
		msg := "The project was not shared because the following email addresses are not valid: "
		msg += strings.Join(invalid_emails, ", ") + "."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	var err error
	err = addSharing(ctx, pkey, add_users)
	if err != nil {
		msg := "Datastore error: unable to update sharing information."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		log.Errorf(ctx, "Edit_sharing_confirm [1]: %v", err)
		return
	}

	remove_users := r.Form["remove_users"]
	err = removeSharing(ctx, pkey, remove_users)
	if err != nil {
		msg := "Datastore error: unable to update sharing information."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		log.Errorf(ctx, "Edit_sharing_confirm [2]: %v", err)
		return
	}

	type TV struct {
		User        string
		LoggedIn    bool
		ProjectName string
		Pkey        string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.ProjectName = project_name
	template_values.Pkey = pkey

	if err := tmpl.ExecuteTemplate(w, "edit_sharing_confirm.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
