package randomization

import (
	"fmt"
	"time"
	//	"strconv"
	"appengine/user"
	"io"
	"net/http"
	//	"math"
	//	"appengine/datastore"
	"appengine"
	"strings"
	//	"html/template"
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

	Pkey := r.FormValue("pkey")

	if ok := Check_access(user, Pkey, &c, &w, r); !ok {
		return
	}

	PR, _ := Get_project_from_key(Pkey, &c)

	if !PR.Store_RawData {
		Msg := "Complete data are not stored for this project."
		Return_msg := "Return to dashboard"
		Message_page(w, r, user, Msg, Return_msg,
			fmt.Sprintf("/project_dashboard?pkey=%s", Pkey))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Header line
	io.WriteString(w, "Subject id,Assignment date,Assignment time,")
	io.WriteString(w, "Assigned group,Final group,Included,Assigner")
	for _, va := range PR.Variables {
		io.WriteString(w, ",")
		io.WriteString(w, va.Name)
	}
	io.WriteString(w, "\n")

	for _, rec := range PR.RawData {
		io.WriteString(w, rec.Subject_id)
		io.WriteString(w, ",")
		t := rec.Assigned_time
		loc, _ := time.LoadLocation("America/New_York")
		t = t.In(loc)
		io.WriteString(w, t.Format("2006-1-2"))
		io.WriteString(w, ",")
		io.WriteString(w, t.Format("3:04 PM EST"))
		io.WriteString(w, ",")
		io.WriteString(w, rec.Assigned_group)
		io.WriteString(w, ",")
		io.WriteString(w, rec.Current_group)
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
