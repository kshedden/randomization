package randomization

import (
	"appengine/user"
	//"fmt"
	"net/http"
	//	"appengine/datastore"
	"appengine"
	//	"strings"
	"html/template"
)

func Dashboard(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	_, PR, err := Get_projects(user.String(), true, &c)
	if err != nil {
		Msg := "A datastore error occured, projects cannot be retrieved."
		c.Errorf("Dashboard: %v", err)
		Return_msg := "Return to dashboard"
		Message_page(w, r, nil, Msg, Return_msg, "/dashboard")
		return
	}

	type TV struct {
		User      string
		Logged_in bool
		PRN       bool
		PR        []*Encoded_Project_view
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.PR = Format_encoded_projects(PR)
	template_values.PRN = len(PR) > 0
	template_values.Logged_in = user != nil

	tmpl, err := template.ParseFiles("header.html", "dashboard.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "dashboard.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}
