package randomization

import (
	"appengine"
	"appengine/user"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// View_statistics
func View_complete_data(w http.ResponseWriter,
	r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	c := appengine.NewContext(r)
	user := user.Current(c)
	pkey := r.FormValue("pkey")

	if ok := Check_access(user, pkey, &c, &w, r); !ok {
		return
	}

	proj, _ := Get_project_from_key(pkey, &c)
	if !proj.StoreRawData {
		msg := "Complete data are not stored for this project."
		return_msg := "Return to dashboard"
		Message_page(w, r, user, msg, return_msg,
			fmt.Sprintf("/project_dashboard?pkey=%s", pkey))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Header line
	io.WriteString(w, "Subject id,Assignment date,Assignment time,")
	io.WriteString(w, "Assigned group,Final group,Included,Assigner")
	for _, va := range proj.Variables {
		io.WriteString(w, ",")
		io.WriteString(w, va.Name)
	}
	io.WriteString(w, "\n")

	for _, rec := range proj.RawData {
		io.WriteString(w, rec.SubjectId)
		io.WriteString(w, ",")
		t := rec.AssignedTime
		loc, _ := time.LoadLocation("America/New_York")
		t = t.In(loc)
		io.WriteString(w, t.Format("2006-1-2"))
		io.WriteString(w, ",")
		io.WriteString(w, t.Format("3:04 PM EST"))
		io.WriteString(w, ",")
		io.WriteString(w, rec.AssignedGroup)
		io.WriteString(w, ",")
		io.WriteString(w, rec.CurrentGroup)
		io.WriteString(w, ",")
		if rec.Included {
			io.WriteString(w, "Yes,")
		} else {
			io.WriteString(w, "No,")
		}
		io.WriteString(w, rec.Assigner+",")
		io.WriteString(w, strings.Join(rec.Data, ","))
		io.WriteString(w, "\n")
	}
}
