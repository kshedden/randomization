package randomization

import (
	"appengine"
	"appengine/user"
	"html/template"
	"net/http"
	"strings"
)

// Create_project_step1 gets the project name from the user.
func Project_dashboard(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)
	pkey := r.FormValue("pkey")

	if ok := Check_access(user, pkey, &c, &w, r); !ok {
		return
	}

	splkey := strings.Split(pkey, "::")
	owner := splkey[0]

	project, _ := Get_project_from_key(pkey, &c)
	project_view := Format_project(project)

	type TV struct {
		User            string
		LoggedIn        bool
		ProjView        *ProjectView
		NumGroups       int
		Sharing         string
		SharedUsers     []string
		Pkey            string
		ShowEditSharing bool
		Owner           string
		StoreRawData    string
		Open            string
		AnyVars         bool
	}

	susers, _ := Get_shared_users(pkey, &c)

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.ProjView = project_view
	template_values.NumGroups = len(project.GroupNames)
	template_values.AnyVars = len(project.Variables) > 0
	if project.StoreRawData {
		template_values.StoreRawData = "Yes"
	} else {
		template_values.StoreRawData = "No"
	}
	if len(susers) > 0 {
		template_values.Sharing = strings.Join(susers, ", ")
	} else {
		template_values.Sharing = "Nobody"
	}
	template_values.Pkey = pkey
	template_values.ShowEditSharing = owner == user.String()
	template_values.Owner = owner

	if project_view.Open {
		template_values.Open = "Yes"
	} else {
		template_values.Open = "No"
	}

	tmpl, err := template.ParseFiles("header.html", "project_dashboard.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "project_dashboard.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}
