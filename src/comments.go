package randomization

import (
	"net/http"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// viewComments
func viewComments(w http.ResponseWriter, r *http.Request) {

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

	PR, _ := getProjectFromKey(ctx, pkey)
	PV := formatProject(PR)

	for _, c := range PV.Comments {
		c.Date = c.DateTime.Format("2006-1-2")
		c.Time = c.DateTime.Format("3:04pm")
	}

	type TV struct {
		User         string
		LoggedIn     bool
		PR           *Project
		PV           *ProjectView
		Pkey         string
		Any_comments bool
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.PR = PR
	template_values.PV = PV
	template_values.Any_comments = len(PR.Comments) > 0
	template_values.Pkey = pkey

	if err := tmpl.ExecuteTemplate(w, "view_comments.html",
		template_values); err != nil {
		log.Errorf(ctx, "View_comments: %v", err)
	}
}

// addComment
func addComment(w http.ResponseWriter, r *http.Request) {

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

	proj, _ := getProjectFromKey(ctx, pkey)
	PV := formatProject(proj)

	type TV struct {
		User     string
		LoggedIn bool
		PR       *Project
		PV       *ProjectView
		Pkey     string
	}

	template_values := new(TV)
	template_values.User = user.String()
	template_values.LoggedIn = user != nil
	template_values.PR = proj
	template_values.PV = PV
	template_values.Pkey = pkey

	if err := tmpl.ExecuteTemplate(w, "add_comment.html",
		template_values); err != nil {
		log.Errorf(ctx, "Add_comment: %v", err)
	}
}

// confirmAddComment
func confirmAddComment(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
		return
	}

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error, unable to add comment."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		log.Errorf(ctx, "Confirm_add_comment [1]: %v", err)
		return
	}

	comment_text := r.FormValue("comment_text")
	comment_text = strings.TrimSpace(comment_text)
	comment_lines := strings.Split(comment_text, "\n")

	if len(comment_text) == 0 {
		msg := "No comment was entered."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
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
	proj.Comments = append(proj.Comments, comment)

	storeProject(ctx, proj, pkey)

	msg := "Your comment has been added to the project."
	rmsg := "Return to project"
	messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
}
