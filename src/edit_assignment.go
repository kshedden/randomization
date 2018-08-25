package randomization

import (
	"fmt"
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// editAssignment
func editAssignment(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := appengine.NewContext(r)
	user := user.Current(ctx)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(ctx, user, pkey, &w, r); !ok {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	proj, err := getProjectFromKey(ctx, pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		log.Errorf(ctx, "Edit_assignment_confirm [1]: %v", err)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.NumAssignments == 0 {
		msg := "There are no assignments to edit."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.StoreRawData == false {
		msg := "Group assignments cannot be edited for a project in which the subject level data is not stored"
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Pkey        string
		ProjectName string
		GroupNames  []string
	}{
		User:        user.String(),
		LoggedIn:    user != nil,
		Pkey:        pkey,
		ProjectName: proj.Name,
		GroupNames:  proj.GroupNames,
	}

	if err := tmpl.ExecuteTemplate(w, "edit_assignment.html", tvals); err != nil {
		log.Errorf(ctx, "editAssignment failed to execute template: %v", err)
	}
}

// editAssignmentConfirm
func editAssignmentConfirm(w http.ResponseWriter, r *http.Request) {

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
		log.Errorf(ctx, "Edit_assignment_confirm [1]: %v", err)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.StoreRawData == false {
		msg := "Assignments cannot be edited in a project in which the subject level data is not stored"
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	subjectId := r.FormValue("SubjectId")

	tvals := struct {
		User             string
		LoggedIn         bool
		Pkey             string
		ProjectName      string
		CurrentGroupName string
		NewGroupName     string
		SubjectId        string
	}{
		User:         user.String(),
		LoggedIn:     user != nil,
		Pkey:         pkey,
		ProjectName:  proj.Name,
		NewGroupName: r.FormValue("NewGroupName"),
		SubjectId:    subjectId,
	}

	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			tvals.CurrentGroupName = rec.CurrentGroup
			found = true
		}
	}
	if !found {
		msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subjectId)
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if tvals.CurrentGroupName == tvals.NewGroupName {
		msg := fmt.Sprintf("You have requested to change the treatment group of subject '%s' to '%s', but the subject is already in this treatment group.", subjectId, tvals.NewGroupName)
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "edit_assignment_confirm.html", tvals); err != nil {
		log.Errorf(ctx, "editAssignmentConfirm failed to execute template: %v", err)
	}
}

// editAssignmentCompleted
func editAssignmentCompleted(w http.ResponseWriter, r *http.Request) {

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
		log.Errorf(ctx, "Edit_assignment_completed [1]: %v", err)
		return
	}

	if proj.Owner != user.String() {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.StoreRawData == false {
		msg := "Group assignments cannot be edited in a project in which the subject level data is not stored."
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	newGroupName := r.FormValue("new_group_name")
	subjectId := r.FormValue("subject_id")

	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			removeFromAggregate(rec, proj)
			oldGroupName := rec.CurrentGroup
			rec.CurrentGroup = newGroupName
			addToAggregate(rec, proj)

			comment := new(Comment)
			comment.Person = user.String()
			comment.DateTime = time.Now()
			comment.Comment = []string{
				fmt.Sprintf("Group assignment for subject '%s' changed from '%s' to '%s'",
					subjectId, oldGroupName, newGroupName)}
			proj.Comments = append(proj.Comments, comment)

			found = true
		}
	}
	if !found {
		msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subjectId)
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	storeProject(ctx, proj, pkey)

	msg := "The assignment has been changed."
	rmsg := "Return to project"
	messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
}
