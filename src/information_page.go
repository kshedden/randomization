package randomization

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// informationPage displays a page of information about this application.
func informationPage(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)

	tvals := struct {
		User     string
		LoggedIn bool
	}{
		LoggedIn: user != nil,
	}

	if user != nil {
		tvals.User = user.String()
	}

	if err := tmpl.ExecuteTemplate(w, "information_page.html", tvals); err != nil {
		log.Errorf(ctx, "Execute template faile in informationPage: %v", err)
	}
}
