## Randomization tool

The randomization tool is a web application that can be used to
support research trials that involve sequential enrollment of
subjects.  Subjects are assigned to treatment arms using the Pocock
and Simon minimization algorithm (<em>Biometrics</em> 31, 1975).
Subjects are assigned randomly to treatment groups, but the assignment
probabilities are adjusted to promote balance with respect to
confounding factors.

Some features of the randomization tool are:

* Hosted using Google Appengine, typical usage on a standard Google
  user account does not incur any charges.

* Cloud-based data storage in the Google Cloud Datastore, data are
  replicated on multiple servers.

* Login and authentication using Google accounts.

* Unlimited projects per user.  Distinct roles for project leaders and
  study managers.

* Supports multi-center trials.

* Option to disable online storage of disaggregated data.

* Customization of post-randomization data editing.


### Authentication and data security

Users are authenticated through their Google account.  This is the
only form of authentication that is currently supported.

The randomization tool must store aggregated data to apply the
minimization algorithm ("aggregated data" are the marginal totals per
treatment arm within each level of each covariate).  Disaggregated
data are not stored by default, but the project creator can opt-in to
have disaggregated data stored.

The data are stored in a cloud database on Google servers.  If the
Google appengine account on which the randomization tool was installed
is compromised, data could be lost or viewed by others.  Google
supports [two factor
authentication](https://www.google.com/landing/2step/) for greater
security.  If the project owner's account is compromised the project
could be deleted or viewed by others.  If a study manager's account is
compromised, the data could be viewed by others, but any changes could
be reverted (reverting is only possible with opt-in for storing
disaggregated data).

### Installation on Google Appengine

The project is written in the [Go](golang.org) programming language,
but no knowledge of Go is needed to install or use the application.

1. Download and unpack the [latest
version](https://github.com/kshedden/randomization/archive/master.zip)
of the application.

2. Download the appropriate version of the Go Appengine Software
development kit for your platform from
[here](https://cloud.google.com/appengine/downloads#Google_App_Engine_SDK_for_Go).
Unpack and install as appropriate for your system.

3. Navigate to https://console.cloud.google.com to create a project.
More detailed instructions are
[here](https://cloud.google.com/appengine/docs/go/gettingstarted/uploading).
Note that you will need to edit the "app.yaml" file -- replace
"randomization" in the first line with the application id that you
select when creating your Appengine project on-line.

4. Deploy your application using `goapp` or `appcfg.py`.  Note that
you must be in the `src` directory of the randomization project when
deplying.

5. If you call your application "my_randomization", then after
deployment it will be available at the URL
"my_randomization.appspot.com".

### Development and bug reports

The project is hosted at https://github.com/kshedden/randomization.
Bug reports, feature requests, and pull requests can be made through
the Github site.