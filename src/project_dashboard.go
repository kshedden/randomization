package randomization

import (
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// projectDashboard gets the project name from the user.
func projectDashboard(w http.ResponseWriter, r *http.Request) {

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

	splkey := strings.Split(pkey, "::")
	owner := splkey[0]

	proj, _ := getProjectFromKey(ctx, pkey)
	projView := formatProject(proj)

	susers, _ := getSharedUsers(ctx, pkey)

	tvals := struct {
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
	}{
		User:            user.String(),
		LoggedIn:        user != nil,
		ProjView:        projView,
		NumGroups:       len(proj.GroupNames),
		AnyVars:         len(proj.Variables) > 0,
		Pkey:            pkey,
		ShowEditSharing: owner == user.String(),
		Owner:           owner,
	}

	if proj.StoreRawData {
		tvals.StoreRawData = "Yes"
	} else {
		tvals.StoreRawData = "No"
	}

	if len(susers) > 0 {
		tvals.Sharing = strings.Join(susers, ", ")
	} else {
		tvals.Sharing = "Nobody"
	}

	if projView.Open {
		tvals.Open = "Yes"
	} else {
		tvals.Open = "No"
	}

	if err := tmpl.ExecuteTemplate(w, "project_dashboard.html", tvals); err != nil {
		log.Errorf(ctx, "projectDashbord failed to execute template: %v", err)
	}
}
