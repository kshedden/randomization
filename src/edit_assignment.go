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

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
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

	type TV struct {
		User        string
		LoggedIn    bool
		Pkey        string
		ProjectName string
		GroupNames  []string
	}

	templateValues := new(TV)
	templateValues.User = user.String()
	templateValues.LoggedIn = user != nil
	templateValues.Pkey = pkey
	templateValues.ProjectName = proj.Name
	templateValues.GroupNames = proj.GroupNames

	if err := tmpl.ExecuteTemplate(w, "edit_assignment.html", templateValues); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
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

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
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

	type TV struct {
		User             string
		LoggedIn         bool
		Pkey             string
		ProjectName      string
		CurrentGroupName string
		NewGroupName     string
		SubjectId        string
	}

	templateValues := new(TV)
	templateValues.User = user.String()
	templateValues.LoggedIn = user != nil
	templateValues.Pkey = pkey
	templateValues.ProjectName = proj.Name
	templateValues.NewGroupName = r.FormValue("NewGroupName")
	templateValues.SubjectId = subjectId

	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			templateValues.CurrentGroupName = rec.CurrentGroup
			found = true
		}
	}
	if !found {
		msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subjectId)
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if templateValues.CurrentGroupName == templateValues.NewGroupName {
		msg := fmt.Sprintf("You have requested to change the treatment group of subject '%s' to '%s', but the subject is already in this treatment group.", subjectId, templateValues.NewGroupName)
		rmsg := "Return to project"
		messagePage(w, r, user, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "edit_assignment_confirm.html", templateValues); err != nil {
		log.Errorf(ctx, "Failed to execute template: %v", err)
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

	if ok := checkAccess(user, pkey, ctx, &w, r); !ok {
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
