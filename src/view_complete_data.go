package randomization

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

// viewCompleteData
func viewCompleteData(w http.ResponseWriter, r *http.Request) {

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
	if !proj.StoreRawData {
		msg := "Complete data are not stored for this project."
		rmsg := "Return to dashboard"
		messagePage(w, r, user, msg, rmsg, fmt.Sprintf("/project_dashboard?pkey=%s", pkey))
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
