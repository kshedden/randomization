package randomization

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// viewStatistics
func viewStatistics(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
		return
	}

	var err error
	project, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		log.Errorf(ctx, "View_statistics [1]: %v", err)
		msg := "Datastore error: unable to view statistics."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, "/dashboard")
		log.Errorf(ctx, "View_statistics [1]: %v", err)
		return
	}
	project_view := formatProject(project)

	// Treatment assignment.
	tx_asgn := make([][]string, len(project.GroupNames))
	for k, v := range project.GroupNames {
		tx_asgn[k] = []string{v, fmt.Sprintf("%d", project.Assignments[k])}
	}

	num_groups := len(project.GroupNames)
	data := project.Data

	m := 0
	for _, v := range project.Variables {
		m += len(v.Levels)
	}

	// Balance statistics.
	bal_stat := make([][]string, m)
	jj := 0
	for j, v := range project.Variables {
		num_levels := len(v.Levels)
		for k := 0; k < num_levels; k++ {
			fstat := make([]string, 1+num_groups)
			fstat[0] = v.Name + "=" + v.Levels[k]
			for q := 0; q < num_groups; q++ {
				u := data[j][k][q]
				fstat[q+1] = fmt.Sprintf("%.0f", u)
			}
			bal_stat[jj] = fstat
			jj++
		}
	}

	type TV struct {
		User        string
		LoggedIn    bool
		Project     *Project
		AnyVars     bool
		ProjectView *ProjectView
		TxAsgn      [][]string
		BalStat     [][]string
		Pkey        string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.Project = project
	template_values.AnyVars = len(project.Variables) > 0
	template_values.ProjectView = project_view
	template_values.TxAsgn = tx_asgn
	template_values.Pkey = pkey
	template_values.BalStat = bal_stat

	if err := tmpl.ExecuteTemplate(w, "view_statistics.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
