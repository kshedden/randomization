package randomization

import (
	"appengine"
	"appengine/user"
	"html/template"
	"net/http"
)

// Information_page displays a page of information about this application.
func Information_page(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)

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

	tmpl, err := template.ParseFiles("header.html", "information_page.html")
	if err != nil {
		c.Errorf("Information_page: %v", err)
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "information_page.html",
		template_values); err != nil {
		c.Errorf("Information page: %v", err)
	}
}
