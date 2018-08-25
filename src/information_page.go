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

	type TV struct {
		User     string
		LoggedIn bool
	}

	template_values := new(TV)
	if user != nil {
		template_values.User = user.String()
	} else {
		template_values.User = ""
	}
	template_values.LoggedIn = user != nil

	if err := tmpl.ExecuteTemplate(w, "information_page.html", template_values); err != nil {
		log.Errorf(ctx, "Information page: %v", err)
	}
}
