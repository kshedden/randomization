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

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		return
	}

	PR, _ := getProjectFromKey(ctx, pkey)
	PV := formatProject(PR)

	for _, c := range PV.Comments {
		c.Date = c.DateTime.Format("2006-1-2")
		c.Time = c.DateTime.Format("3:04pm")
	}

	tvals := struct {
		User         string
		LoggedIn     bool
		PR           *Project
		PV           *ProjectView
		Pkey         string
		Any_comments bool
	}{
		User:         user.String(),
		LoggedIn:     user != nil,
		PR:           PR,
		PV:           PV,
		Any_comments: len(PR.Comments) > 0,
		Pkey:         pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "view_comments.html", tvals); err != nil {
		log.Errorf(ctx, "ViewComments: %v", err)
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

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		return
	}

	proj, _ := getProjectFromKey(ctx, pkey)
	fproj := formatProject(proj)

	tvals := struct {
		User     string
		LoggedIn bool
		PR       *Project
		PV       *ProjectView
		Pkey     string
	}{
		User:     user.String(),
		LoggedIn: user != nil,
		PR:       proj,
		PV:       fproj,
		Pkey:     pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "add_comment.html", tvals); err != nil {
		log.Errorf(ctx, "addComment: %v", err)
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

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		return
	}

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error, unable to add comment."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		log.Errorf(ctx, "confirmAddComment [1]: %v", err)
		return
	}

	commentText := r.FormValue("comment_text")
	commentText = strings.TrimSpace(commentText)
	commentLines := strings.Split(commentText, "\n")

	if len(commentText) == 0 {
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
	comment.Comment = commentLines
	proj.Comments = append(proj.Comments, comment)

	storeProject(ctx, proj, pkey)

	msg := "Your comment has been added to the project."
	rmsg := "Return to project"
	messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
}
