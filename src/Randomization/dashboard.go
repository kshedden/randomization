package randomization

import (
	"appengine"
	"appengine/user"
	"html/template"
	"net/http"
)

func Dashboard(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)

	_, projlist, err := GetProjects(user.String(), true, &c)
	if err != nil {
		Msg := "A datastore error occured, projects cannot be retrieved."
		c.Errorf("Dashboard: %v", err)
		Return_msg := "Return to dashboard"
		Message_page(w, r, nil, Msg, Return_msg, "/dashboard")
		return
	}

	type TV struct {
		User     string
		LoggedIn bool
		PRN      bool
		PR       []*EncodedProjectView
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.PR = Format_encoded_projects(projlist)
	template_values.PRN = len(projlist) > 0
	template_values.LoggedIn = user != nil

	tmpl, err := template.ParseFiles("header.html", "dashboard.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "dashboard.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}
