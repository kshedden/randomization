package randomization

import (
	//	"fmt"
	//	"time"
	//	"strconv"
	"appengine/user"
	"net/http"
	//	"appengine/datastore"
	"appengine"
	"html/template"
	"math/rand"
	"strings"
)

// Create_project_step1 gets the project name from the user.
func Project_dashboard(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	// debugging to get a sense of how rand performs without initialization
	c.Infof("%v\n", rand.Float64())

	user := user.Current(c)

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	A := strings.Split(Pkey, "::")
	owner := A[0]

	project, _ := Get_project_from_key(Pkey, &c)
	project_view := Format_project(project)

	type TV struct {
		User              string
		Logged_in         bool
		Proj_view         *Project_view
		Numgroups         int
		Sharing           string
		Shared_users      []string
		Pkey              string
		Show_edit_sharing bool
		Owner             string
		Store_RawData     string
		Open              string
		Any_vars          bool
	}

	SU, _ := Get_shared_users(Pkey, &c)

	template_values := new(TV)
	template_values.User = user.String()
	template_values.Logged_in = user != nil
	template_values.Proj_view = project_view
	template_values.Numgroups = len(project.Group_names)
	template_values.Any_vars = len(project.Variables) > 0
	if project.Store_RawData {
		template_values.Store_RawData = "Yes"
	} else {
		template_values.Store_RawData = "No"
	}
	if len(SU) > 0 {
		template_values.Sharing = strings.Join(SU, ", ")
	} else {
		template_values.Sharing = "Nobody"
	}
	template_values.Pkey = Pkey
	template_values.Show_edit_sharing = owner == user.String()
	template_values.Owner = owner

	if project_view.Open {
		template_values.Open = "Yes"
	} else {
		template_values.Open = "No"
	}

	tmpl, err := template.ParseFiles("header.html",
		"project_dashboard.html")
	if err != nil {
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "project_dashboard.html",
		template_values); err != nil {
		c.Errorf("Failed to execute template: %v", err)
	}
}
