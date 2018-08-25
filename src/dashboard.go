package randomization

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	_, projlist, err := getProjects(ctx, user.String(), true)
	if err != nil {
		msg := "A datastore error occured, projects cannot be retrieved."
		log.Errorf(ctx, "Dashboard: %v", err)
		rmsg := "Return to dashboard"
		messagePage(w, r, nil, msg, rmsg, "/dashboard")
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
	template_values.PR = formatEncodedProjects(projlist)
	template_values.PRN = len(projlist) > 0
	template_values.LoggedIn = user != nil

	if err := tmpl.ExecuteTemplate(w, "dashboard.html", template_values); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}
