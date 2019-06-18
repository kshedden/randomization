package randomization

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// removeSubject
func removeSubject(w http.ResponseWriter, r *http.Request) {

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

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can remove treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if !proj.StoreRawData {
		msg := "Subjects cannot be removed for a project in which the subject level data is not stored"
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	tvals := struct {
		User                 string
		LoggedIn             bool
		Pkey                 string
		ProjectName          string
		Any_removed_subjects bool
		RemovedSubjects      string
	}{
		User:        user.String(),
		LoggedIn:    user != nil,
		Pkey:        pkey,
		ProjectName: proj.Name,
	}

	if len(proj.RemovedSubjects) > 0 {
		tvals.Any_removed_subjects = true
		tvals.RemovedSubjects = strings.Join(proj.RemovedSubjects, ", ")
	} else {
		tvals.Any_removed_subjects = false
	}

	if err := tmpl.ExecuteTemplate(w, "remove_subject.html", tvals); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// removeSubjectConfirm
func removeSubjectConfirm(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")
	subjectId := r.FormValue("subject_id")

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can remove treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	// Check if the subject has already been removed
	for _, s := range proj.RemovedSubjects {
		if s == subjectId {
			msg := fmt.Sprintf("Subject '%s' has already been removed from the study.", subjectId)
			rmsg := "Return to project"
			messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
			return
		}
	}

	// Check if the subject exists
	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			found = true
			break
		}
	}
	if !found {
		msg := fmt.Sprintf("There is no subject with id '%s' in the project.", subjectId)
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Pkey        string
		SubjectId   string
		ProjectName string
	}{
		User:        user.String(),
		LoggedIn:    user != nil,
		SubjectId:   subjectId,
		Pkey:        pkey,
		ProjectName: proj.Name,
	}

	if err := tmpl.ExecuteTemplate(w, "remove_subject_confirm.html", tvals); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
	}
}

// removeSubjectCompleted
func removeSubjectCompleted(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		msg := "You do not have access to this page."
		rmsg := "Return"
		messagePage(w, r, user, msg, rmsg, "/")
		return
	}

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can remove treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if !proj.StoreRawData {
		msg := "Subjects cannot be removed for a project in which the subject level data is not stored"
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	subjectId := r.FormValue("subject_id")
	found := false
	var removeRec *DataRecord
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			rec.Included = false
			removeRec = rec
			found = true
		}
	}
	proj.RemovedSubjects = append(proj.RemovedSubjects, subjectId)

	comment := new(Comment)
	comment.Person = user.String()
	comment.DateTime = time.Now()
	comment.Comment = []string{fmt.Sprintf("Subject '%s' removed from the project.", subjectId)}
	proj.Comments = append(proj.Comments, comment)

	if !found {
		msg := fmt.Sprintf("Unable to remove subject '%s' from the project.", subjectId)
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	removeFromAggregate(removeRec, proj)
	proj.NumAssignments--
	err = storeProject(ctx, proj, pkey)
	if err != nil {
		msg := "Error, unable to save project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	msg := fmt.Sprintf("Subject '%s' has been removed from the study.", subjectId)
	rmsg := "Return to project dashboard"
	messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
}
