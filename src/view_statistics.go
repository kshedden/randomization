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

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
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
	projectView := formatProject(project)

	// Treatment assignment.
	txAsgn := make([][]string, len(project.GroupNames))
	for k, v := range project.GroupNames {
		txAsgn[k] = []string{v, fmt.Sprintf("%d", project.Assignments[k])}
	}

	numGroups := len(project.GroupNames)
	data := project.Data

	m := 0
	for _, v := range project.Variables {
		m += len(v.Levels)
	}

	// Balance statistics
	balStat := make([][]string, m)
	jj := 0
	for j, v := range project.Variables {
		numLevels := len(v.Levels)
		for k := 0; k < numLevels; k++ {
			fstat := make([]string, 1+numGroups)
			fstat[0] = v.Name + "=" + v.Levels[k]
			for q := 0; q < numGroups; q++ {
				u := data[j][k][q]
				fstat[q+1] = fmt.Sprintf("%.0f", u)
			}
			balStat[jj] = fstat
			jj++
		}
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Project     *Project
		AnyVars     bool
		ProjectView *ProjectView
		TxAsgn      [][]string
		BalStat     [][]string
		Pkey        string
	}{
		User:        user.String(),
		LoggedIn:    user != nil,
		Project:     project,
		AnyVars:     len(project.Variables) > 0,
		ProjectView: projectView,
		TxAsgn:      txAsgn,
		Pkey:        pkey,
		BalStat:     balStat,
	}

	if err := tmpl.ExecuteTemplate(w, "view_statistics.html", tvals); err != nil {
		log.Errorf(ctx, "viewStatistics failed to execute template: %v", err)
	}
}
