package randomization

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

var tmpl = template.Must(template.ParseGlob("html_templates/*.html"))

func init() {

	http.HandleFunc("/", informationPage)
	http.HandleFunc("/dashboard", requireLogin(Dashboard))

	// Project creation pages
	http.HandleFunc("/create_project_step1", requireLogin(CreateProjectStep1))
	http.HandleFunc("/create_project_step2", requireLogin(CreateProjectStep2))
	http.HandleFunc("/create_project_step3", requireLogin(CreateProjectStep3))
	http.HandleFunc("/create_project_step4", requireLogin(CreateProjectStep4))
	http.HandleFunc("/create_project_step5", requireLogin(CreateProjectStep5))
	http.HandleFunc("/create_project_step6", requireLogin(CreateProjectStep6))
	http.HandleFunc("/create_project_step7", requireLogin(CreateProjectStep7))
	http.HandleFunc("/create_project_step8", requireLogin(CreateProjectStep8))
	http.HandleFunc("/create_project_step9", requireLogin(CreateProjectStep9))

	// Copy project pages
	http.HandleFunc("/copy_project", requireLogin(copyProject))
	http.HandleFunc("/copy_project_completed", requireLogin(copyProjectCompleted))

	// Project deletion pages
	http.HandleFunc("/delete_project_step1", requireLogin(deleteProjectStep1))
	http.HandleFunc("/delete_project_step2", requireLogin(deleteProjectStep2))
	http.HandleFunc("/delete_project_step3", requireLogin(deleteProjectStep3))

	http.HandleFunc("/project_dashboard", requireLogin(projectDashboard))
	http.HandleFunc("/edit_sharing", requireLogin(editSharing))
	http.HandleFunc("/edit_sharing_confirm", requireLogin(editSharingConfirm))

	// Treatment assignment pages
	http.HandleFunc("/assign_treatment_input", requireLogin(assignTreatmentInput))
	http.HandleFunc("/assign_treatment_confirm", requireLogin(assignTreatmentConfirm))
	http.HandleFunc("/assign_treatment", requireLogin(assignTreatment))

	http.HandleFunc("/view_statistics", requireLogin(viewStatistics))
	http.HandleFunc("/view_comments", requireLogin(viewComments))
	http.HandleFunc("/add_comment", requireLogin(addComment))
	http.HandleFunc("/confirm_add_comment", requireLogin(confirmAddComment))
	http.HandleFunc("/view_complete_data", requireLogin(viewCompleteData))

	// Remove subject pages
	http.HandleFunc("/remove_subject", requireLogin(removeSubject))
	http.HandleFunc("/remove_subject_confirm", requireLogin(removeSubjectConfirm))
	http.HandleFunc("/remove_subject_completed", requireLogin(removeSubjectCompleted))

	// Edit assignment pages
	http.HandleFunc("/edit_assignment", requireLogin(editAssignment))
	http.HandleFunc("/edit_assignment_confirm", requireLogin(editAssignmentConfirm))
	http.HandleFunc("/edit_assignment_completed", requireLogin(editAssignmentCompleted))

	// Close or open project for enrollment pages
	http.HandleFunc("/openclose_project", requireLogin(openCloseProject))
	http.HandleFunc("/openclose_completed", requireLogin(openCloseCompleted))
}

// checkAccess determines whether the given user has permission to
// access the given project.
func checkAccess(user *user.User, pkey string, ctx context.Context, w *http.ResponseWriter, r *http.Request) bool {

	user_name := strings.ToLower(user.String())

	keyparts := strings.Split(pkey, "::")
	owner := keyparts[0]

	// A user can always access his or her own projects.
	if user_name == strings.ToLower(owner) {
		return true
	}

	// Otherwise, check if the project is shared with the user.
	key := datastore.NewKey(ctx, "SharingByUser", user_name, 0, nil)
	var sbuser SharingByUser
	err := datastore.Get(ctx, key, &sbuser)
	if err == datastore.ErrNoSuchEntity {
		check_access_failed(nil, ctx, w, r, user)
		return false
	} else if err != nil {
		check_access_failed(&err, ctx, w, r, user)
		return false
	}
	L := cleanSplit(sbuser.Projects, ",")
	for _, x := range L {
		if pkey == x {
			return true
		}
	}
	check_access_failed(nil, ctx, w, r, user)
	return false
}

// check_access_failed displays an error message when a project cannot be accessed.
func check_access_failed(err *error, ctx context.Context, w *http.ResponseWriter, r *http.Request, user *user.User) {

	if err != nil {
		msg := "A datastore error occured.  Ask the administrator to check the log for error details."
		log.Errorf(ctx, "check_access_failed: %v", err)
		rmsg := "Return to dashboard"
		messagePage(*w, r, user, msg, rmsg, "/dashboard")
		return
	}
	msg := "You don't have access to this project."
	rmsg := "Return to dashboard"
	log.Infof(ctx, fmt.Sprintf("Failed access: %v\n", user))
	messagePage(*w, r, user, msg, rmsg, "/dashboard")
}

type handler func(http.ResponseWriter, *http.Request)

// requireLogin is a wrapper for a function that serves web pages.
// By wrapping the function in require_login, the user is forced to
// log in to their Google account in order to access the system.
func requireLogin(H handler) handler {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := appengine.NewContext(r)

		// Force the person to log in.
		client := user.Current(ctx)
		if client == nil {

			U := strings.Split(r.URL.String(), "/")
			V := U[len(U)-1]
			url, err := user.LoginURL(ctx, V)
			if err != nil {
				http.Error(w, "Error",
					http.StatusInternalServerError)
				return
			}
			log.Infof(ctx, fmt.Sprintf("Login url: %s", url))
			msg := "To use this site, you must be logged into your Google account."
			rmsg := "Continue to login page"
			messagePage(w, r, nil, msg, rmsg, url)
			return
		}

		H(w, r)
	}
}
