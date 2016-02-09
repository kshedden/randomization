package randomization

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"net/http"
	"strings"
)

func init() {

	http.HandleFunc("/", Information_page)
	http.HandleFunc("/dashboard", require_login(Dashboard))

	// Project creation pages
	http.HandleFunc("/create_project_step1",
		require_login(Create_project_step1))
	http.HandleFunc("/create_project_step2",
		require_login(Create_project_step2))
	http.HandleFunc("/create_project_step3",
		require_login(Create_project_step3))
	http.HandleFunc("/create_project_step4",
		require_login(Create_project_step4))
	http.HandleFunc("/create_project_step5",
		require_login(Create_project_step5))
	http.HandleFunc("/create_project_step6",
		require_login(Create_project_step6))
	http.HandleFunc("/create_project_step7",
		require_login(Create_project_step7))
	http.HandleFunc("/create_project_step8",
		require_login(Create_project_step8))
	http.HandleFunc("/create_project_step9",
		require_login(Create_project_step9))

	// Copy project pages
	http.HandleFunc("/copy_project",
		require_login(Copy_project))
	http.HandleFunc("/copy_project_completed",
		require_login(Copy_project_completed))

	// Project deletion pages
	http.HandleFunc("/delete_project_step1",
		require_login(Delete_project_step1))
	http.HandleFunc("/delete_project_step2",
		require_login(Delete_project_step2))
	http.HandleFunc("/delete_project_step3",
		require_login(Delete_project_step3))

	http.HandleFunc("/project_dashboard",
		require_login(Project_dashboard))
	http.HandleFunc("/edit_sharing",
		require_login(Edit_sharing))
	http.HandleFunc("/edit_sharing_confirm",
		require_login(Edit_sharing_confirm))

	// Treatment assignment pages
	http.HandleFunc("/assign_treatment_input",
		require_login(Assign_treatment_input))
	http.HandleFunc("/assign_treatment_confirm",
		require_login(Assign_treatment_confirm))
	http.HandleFunc("/assign_treatment",
		require_login(Assign_treatment))

	http.HandleFunc("/view_statistics",
		require_login(View_statistics))
	http.HandleFunc("/view_comments",
		require_login(View_comments))
	http.HandleFunc("/add_comment",
		require_login(Add_comment))
	http.HandleFunc("/confirm_add_comment",
		require_login(Confirm_add_comment))
	http.HandleFunc("/view_complete_data",
		require_login(View_complete_data))

	// Remove subject pages
	http.HandleFunc("/remove_subject",
		require_login(Remove_subject))
	http.HandleFunc("/remove_subject_confirm",
		require_login(Remove_subject_confirm))
	http.HandleFunc("/remove_subject_completed",
		require_login(Remove_subject_completed))

	// Edit assignment pages
	http.HandleFunc("/edit_assignment",
		require_login(Edit_assignment))
	http.HandleFunc("/edit_assignment_confirm",
		require_login(Edit_assignment_confirm))
	http.HandleFunc("/edit_assignment_completed",
		require_login(Edit_assignment_completed))

	// Close or open project for enrollment pages
	http.HandleFunc("/openclose_project",
		require_login(OpenClose_project))
	http.HandleFunc("/openclose_completed",
		require_login(OpenClose_completed))
}

// Check_access determines whether the given user has permission to
// access the given project.
func Check_access(user *user.User,
	pkey string,
	c *appengine.Context,
	w *http.ResponseWriter,
	r *http.Request) bool {

	user_name := strings.ToLower(user.String())

	A := strings.Split(pkey, "::")
	owner := A[0]

	// A user can always access his or her own projects.
	if user_name == strings.ToLower(owner) {
		return true
	}

	// Otherwise, check if the project is shared with the user.
	Key := datastore.NewKey(*c, "SharingByUser", user_name, 0, nil)
	var SBU SharingByUser
	err := datastore.Get(*c, Key, &SBU)
	if err == datastore.ErrNoSuchEntity {
		check_access_failed(nil, c, w, r, user)
		return false
	} else if err != nil {
		check_access_failed(&err, c, w, r, user)
		return false
	}
	L := Clean_split(SBU.Projects, ",")
	for _, x := range L {
		if pkey == x {
			return true
		}
	}
	check_access_failed(nil, c, w, r, user)
	return false
}

// check_access_failed displays an error message when a project cannot be accessed.
func check_access_failed(err *error,
	c *appengine.Context,
	w *http.ResponseWriter,
	r *http.Request,
	user *user.User) {

	if err != nil {
		Msg := "A datastore error occured.  Ask the administrator to check the log for error details."
		(*c).Errorf("check_access_failed: %v", err)
		Return_msg := "Return to dashboard"
		Message_page(*w, r, user, Msg, Return_msg, "/dashboard")
		return
	}
	Msg := "You don't have access to this project."
	Return_msg := "Return to dashboard"
	(*c).Infof(fmt.Sprintf("Failed access: %v\n", user))
	Message_page(*w, r, user, Msg, Return_msg, "/dashboard")
}

type handler func(http.ResponseWriter, *http.Request)

// require_login is a wrapper for a function that serves web pages.
// By wrapping the function in require_login, the user is forced to
// log in to their Google account in order to access the system.
func require_login(H handler) handler {

	return func(w http.ResponseWriter, r *http.Request) {

		c := appengine.NewContext(r)

		// Force the person to log in.
		client := user.Current(c)
		if client == nil {

			U := strings.Split(r.URL.String(), "/")
			V := U[len(U)-1]
			url, err := user.LoginURL(c, V)
			if err != nil {
				http.Error(w, "Error",
					http.StatusInternalServerError)
				return
			}
			c.Infof(fmt.Sprintf("Login url: %s", url))
			msg := "To use this site, you must be logged into your Google account."
			return_msg := "Continue to login page"
			Message_page(w, r, nil, msg, return_msg, url)
			return
		}

		H(w, r)
	}
}
