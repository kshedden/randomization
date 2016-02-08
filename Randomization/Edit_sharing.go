package randomization

import (
	"fmt"
	//	"time"
	//	"strconv"
	"appengine/user"
	"net/http"
	//	"appengine/datastore"
	"appengine"
	"html/template"
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

	Pkey := r.FormValue("pkey")

	A := strings.Split(Pkey, "::")
	owner := A[0]
	project_name := A[1]

	if owner != user.String() {
		Msg := "Only the owner of a project can manage sharing."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	Shared_users, err := Get_shared_users(Pkey, &c)
	if err != nil {
		Msg := "Datastore error: unable to retrieve sharing information."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	type TV struct {
		User             string
		Logged_in        bool
		Shared_users     []string
		Any_shared_users bool
		Project_name     string
		Pkey             string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Logged_in = user != nil
	template_values.Shared_users = Shared_users
	template_values.Any_shared_users = len(Shared_users) > 0
	template_values.Project_name = project_name
	template_values.Pkey = Pkey

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

	Pkey := r.FormValue("pkey")

	A := strings.Split(Pkey, "::")
	project_name := A[1]

	Shared_users, err := Get_shared_users(Pkey, &c)
	if err != nil {
		Msg := "Datastore error: unable to retrieve sharing information."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	ap := r.FormValue("additional_people")
	AP := []string{}
	AP = Clean_split(ap, ",")
	for k, x := range AP {
		AP[k] = strings.ToLower(x)
	}

	// Gmail addresses don't use @gmail.com.
	Invalid_emails := ""
	for k, x := range AP {
		A := strings.Split(x, "@")

		if len(A) != 2 {
			if Invalid_emails == "" {
				Invalid_emails = x
			} else {
				Invalid_emails = Invalid_emails + ";" + x
			}

		} else {
			if A[1] == "gmail.com" {
				AP[k] = A[0]
			}
		}
	}

	if Invalid_emails != "" {
		Msg := "The project was not shared because the following email addresses are not valid: "
		CL := Clean_split(Invalid_emails, ";")
		Msg += strings.Join(CL, ", ") + "."
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	err = Add_sharing(Pkey, AP, &c)
	if err != nil {
		Msg := "Datastore error: unable to update sharing information."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	remove_users := r.Form["remove_users"]
	err = Remove_sharing(Pkey, remove_users, &c)
	if err != nil {
		Msg := "Datastore error: unable to update sharing information."
		Return_msg := "Return to dashboard"
		fmt.Printf("\n\n2: %v\n\n", err)
		Message_page(w, r, user, Msg, Return_msg, "/dashboard")
		return
	}

	type TV struct {
		User             string
		Logged_in        bool
		Project_name     string
		Shared_users     []string
		Any_shared_users bool
		Pkey             string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Logged_in = user != nil
	template_values.Shared_users = Shared_users
	template_values.Any_shared_users = len(Shared_users) > 0
	template_values.Project_name = project_name
	template_values.Pkey = Pkey

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
