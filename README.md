## Randomization tool

The randomization tool is a web application that can be used to
support research trials that involve sequential enrollment of
subjects.  Subjects are assigned to treatment arms using the Pocock
and Simon minimization algorithm (<em>Biometrics</em> 31, 1975).
Subjects are assigned randomly to treatment groups, but the assignment
probabilities are adjusted to promote balance with respect to
confounding factors.

__Example__: Suppose we are conducting a trial aiming to compare three
approaches to math tutoring.  The subjects take a pre-test before
beginning the study, where the possible scores are low, medium, and
high.  The following treatment group assignments would be considered
*unbalanced* with respect to the pre-test scores:

| Treatment | Low   | Medium | High |
|-----------|-------|--------|------|
| A         |  11   | 23     |  17  |
| B         |  15   | 16     |  10  |
| C         |  13   | 12     |  27  |

In the situation illustrated by this table, a simple comparison of
post-test scores could be biased.  For example, we might expect
treatment C to do better than treatment A, since stronger students
were enrolled in treatment C compared to treatment A.  This apparent
treatment effect of method C may not have anything to do with the
treatment itself.

Since the treatment assignments are partially random, we do not obtain
perfect balance when using the minimization algorithm.  However the
results will be approximately balanced with high probability.  For
example, we might obtain the following:

| Treatment | Low   | Medium | High |
|-----------|-------|--------|------|
| A         |  14   | 17     |  16  |
| B         |  18   | 16     |  18  |
| C         |  19   | 17     |  20  |


The randomization tool will ensure that within each score band on the
pre-test, the relative numbers of people in each treatment arm are
similar.

### Features

Some features of the randomization tool are:

* Hosted using Google Appengine, typical usage on a standard Google
  user account does not incur any charges.

* Cloud-based data storage in the Google Cloud Datastore, data are
  replicated on multiple servers.

* Login and authentication using Google accounts.

* Unlimited projects per user.

* Distinct roles for project leaders and study managers.

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
deploying the application.

5. If you call your application "my_randomization", then after
deployment it will be available at the URL
"my_randomization.appspot.com".  Note that the first time you install
the application it will take a few minutes for Appengine to build the
database indices.  While this is happening, the application will not
be usable.

### Customization

You can perform any of these simple customizations:

* Replace the file `stylesheets/logo.png` with a PNG containing your
  organization's logo, it will then be displayed on the landing page.

* Add a `stylesheets/favicon.ico` [icon
  file](https://en.wikipedia.org/wiki/ICO_(file_format)) which will be
  displayed on the browser tab.

* Edit the text in the `information_page.html` file however you like.

### Development and bug reports

The project is hosted at https://github.com/kshedden/randomization.
Bug reports, feature requests, and pull requests can be made through
the Github site.