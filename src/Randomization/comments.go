package randomization

import (
	//	"fmt"
	"time"
	//	"strconv"
	"appengine/user"
	"net/http"
	//	"appengine/datastore"
	"appengine"
	"html/template"
	"strings"
)

// View_comment
func View_comments(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)
	PV := Format_project(PR)

	for _, c := range PV.Comments {
		c.Date = c.DateTime.Format("2006-1-2")
		c.Time = c.DateTime.Format("3:04pm")
	}

	type TV struct {
		User         string
		LoggedIn     bool
		PR           *Project
		PV           *Project_view
		Pkey         string
		Any_comments bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.PR = PR
	template_values.PV = PV
	template_values.Any_comments = len(PR.Comments) > 0
	template_values.Pkey = Pkey

	tmpl, err := template.ParseFiles("header.html",
		"view_comments.html")
	if err != nil {
		c.Errorf("View_comments: %v", err)
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "view_comments.html",
		template_values); err != nil {
		c.Errorf("View_comments: %v", err)
	}
}

// Add_comment
func Add_comment(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)
	PV := Format_project(PR)

	type TV struct {
		User     string
		LoggedIn bool
		PR       *Project
		PV       *Project_view
		Pkey     string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.PR = PR
	template_values.PV = PV
	template_values.Pkey = Pkey

	tmpl, err := template.ParseFiles("header.html",
		"add_comment.html")
	if err != nil {
		c.Errorf("Add_comment: %v", err)
		ServeError(&c, w, err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "add_comment.html",
		template_values); err != nil {
		c.Errorf("Add_comment: %v", err)
	}
}

// Confirm_add_comment
func Confirm_add_comment(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)

	user := user.Current(c)

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	comment_text := r.FormValue("comment_text")
	comment_text = strings.TrimSpace(comment_text)
	comment_lines := strings.Split(comment_text, "\n")

	if len(comment_text) == 0 {
		Msg := "No comment was entered."
		Return_msg := "Return to project"
		Message_page(w, r, user, Msg, Return_msg,
			"/project_dashboard?pkey="+Pkey)
		return
	}

	comment := new(Comment)
	comment.Person = user.String()
	comment.DateTime = time.Now()
	loc, _ := time.LoadLocation("America/New_York")
	t := comment.DateTime.In(loc)
	comment.Date = t.Format("2006-1-2")
	comment.Time = t.Format("3:04pm")
	comment.Comment = comment_lines
	PR.Comments = append(PR.Comments, comment)

	Store_project(PR, Pkey, &c)

	Msg := "Your comment has been added to the project."
	Return_msg := "Return to project"
	Message_page(w, r, user, Msg, Return_msg,
		"/project_dashboard?pkey="+Pkey)
}
