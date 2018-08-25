package randomization

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func dashboard(w http.ResponseWriter, r *http.Request) {

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

	tvals := struct {
		User     string
		LoggedIn bool
		PRN      bool
		PR       []*EncodedProjectView
	}{
		User:     user.String(),
		PR:       formatEncodedProjects(projlist),
		PRN:      len(projlist) > 0,
		LoggedIn: user != nil,
	}

	if err := tmpl.ExecuteTemplate(w, "dashboard.html", tvals); err != nil {
		log.Errorf(ctx, "Dashboard failed to execute template: %v", err)
	}
}
