package randomization

import (
	"appengine"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
)

// View_statistics
func View_statistics(w http.ResponseWriter,
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

	var err error
	project, err := Get_project_from_key(pkey, &c)
	if err != nil {
		c.Errorf("View_statistics [1]: %v", err)
		msg := "Datastore error: unable to view statistics."
		return_msg := "Return to dashboard"
		Message_page(w, r, user, msg, return_msg, "/dashboard")
		c.Errorf("View_statistics [1]: %v", err)
		return
	}
	project_view := Format_project(project)

	// Treatment assignment.
	TAS := make([][]string, len(project.GroupNames))
	for k, v := range project.GroupNames {
		TAS[k] = []string{v, fmt.Sprintf("%d", project.Assignments[k])}
	}

	num_groups := len(project.GroupNames)
	data := project.Data

	m := 0
	for _, v := range project.Variables {
		m += len(v.Levels)
	}

	// Balance statistics.
	FVS := make([][]string, m)
	jj := 0
	for j, v := range project.Variables {
		num_levels := len(v.Levels)
		for k := 0; k < num_levels; k++ {
			S := make([]string, 1+num_groups)
			S[0] = v.Name + "=" + v.Levels[k]
			for q := 0; q < num_groups; q++ {
				u := data[j][k][q]
				S[q+1] = fmt.Sprintf("%.0f", u)
			}
			FVS[jj] = S
			jj++
		}
	}

	type TV struct {
		User        string
		LoggedIn    bool
		Project     *Project
		AnyVars     bool
		ProjectView *ProjectView
		TAS         [][]string
		FVS         [][]string
		Pkey        string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Project = project
	template_values.AnyVars = len(project.Variables) > 0
	template_values.ProjectView = project_view
	template_values.TAS = TAS
	template_values.Pkey = pkey
	template_values.FVS = FVS

	tmpl, err := template.ParseFiles("header.html", "view_statistics.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "view_statistics.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}
