package randomization

import (
	"appengine"
	"appengine/user"
	"html/template"
	"net/http"
	"strings"
)

// Edit_sharing
func Edit_sharing(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)
	pkey := r.FormValue("pkey")
	shr := strings.Split(pkey, "::")
	owner := shr[0]
	project_name := shr[1]

	if strings.ToLower(owner) != strings.ToLower(user.String()) {
		msg := "Only the owner of a project can manage sharing."
		return_msg := "Return to dashboard"
		Message_page(w, r, user, msg, return_msg, "/dashboard")
		return
	}

	shared_users, err := Get_shared_users(pkey, &c)
	if err != nil {
		shared_users = make([]string, 0)
		c.Infof("Failed to retrieve sharing: %v %v", project_name, owner)
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

	tmpl, err := template.ParseFiles("header.html",
		"edit_sharing.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "edit_sharing.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}

// Edit_sharing_confirm
func Edit_sharing_confirm(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)
	pkey := r.FormValue("pkey")

	spkey := strings.Split(pkey, "::")
	project_name := spkey[1]

	ap := r.FormValue("additional_people")
	add_users := []string{}
	add_users = Clean_split(ap, ",")
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
		return_msg := "Return to project"
		Message_page(w, r, user, msg, return_msg, "/project_dashboard?pkey="+pkey)
		return
	}

	var err error
	err = Add_sharing(pkey, add_users, &c)
	if err != nil {
		msg := "Datastore error: unable to update sharing information."
		return_msg := "Return to dashboard"
		Message_page(w, r, user, msg, return_msg, "/dashboard")
		c.Errorf("Edit_sharing_confirm [1]: %v", err)
		return
	}

	remove_users := r.Form["remove_users"]
	err = Remove_sharing(pkey, remove_users, &c)
	if err != nil {
		msg := "Datastore error: unable to update sharing information."
		return_msg := "Return to dashboard"
		Message_page(w, r, user, msg, return_msg, "/dashboard")
		c.Errorf("Edit_sharing_confirm [2]: %v", err)
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

	tmpl, err := template.ParseFiles("header.html",
		"edit_sharing_confirm.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "edit_sharing_confirm.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}
