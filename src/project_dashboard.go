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

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
		return
	}

	splkey := strings.Split(pkey, "::")
	owner := splkey[0]

	project, _ := getProjectFromKey(ctx, pkey)
	project_view := formatProject(project)

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

	susers, _ := getSharedUsers(ctx, pkey)

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

	if err := tmpl.ExecuteTemplate(w, "project_dashboard.html",
		template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
