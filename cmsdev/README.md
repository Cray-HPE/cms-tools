# CMSDEV Tool

cmsdev is a test utility for CMS services. The tool executes CT, smoke and API tests for all CMS services. Additionally, cmsdev has been padded with additional functionality to make CMS development and system information gathering easier.

## Installation

Building requires the installation of [Golang](https://golang.org/doc/install).

Be sure to copy the repo to a directory OUTSIDE of $GOPATH/src. In this example, assume we have copied it into /tmp/stash/cms-tools

```bash
cd /tmp/stash/cms-tools/cmsdev
GOOS=linux GOARCH=amd64 go build -mod vendor .
```
or download the binary here
* [linux amd64](./cmsdev)

# NOTE: Please rebuild and commit binary after code changes. Currently, there is no process in place to automatially rebuild after a repository commit.

### Nteworthy file/directory locations

For most CMS components, the stage4/5 CT tests consist of a basic check for kubernetes pods. IMS also includes some calls to the service itself. These tests should absolutely be expanded (while still keeping in mind that longer-running tests should happen after localization).

On an installed system, these are important files and directories:

| File or Directory | Description |
| ------------------|-------------|
| /usr/local/bin/cmsdev | The man himself |
| /opt/cray/tests/cmsdev.log | Detailed log file (not to be mistaken with the output from when it is run). [See an example file](examples/cmsdev.log) |
| /opt/cray/tests/crayctl-stage4/cms/ | Directory containing the CMS CT tests which are called at stage4 of an install. These are shell scripts which call cmsdev |
| /opt/cray/tests/crayctl-stage5/cms/ | Stage5 install CT test scripts |
| /tmp/cray/tests/ | Location of log directories from CT tests run during the various stages of the install and localization. These logs are generated externally from the cmsdev tool, but these are usually where you will be pointed to begin investigating a failure |

Noteworthy files in the repo:

| File or Directory | Description |
| ------------------|-------------|
| cmsdev/cmsdev | cmsdev binary |
| [cmsdev/internal/cmd/test.go](internal/cmd/test.go) | Main test driver |
| [cmsdev/internal/test/](internal/test/) | Every CMS component which is tested has a directory here that contains all test code |
| [cmsdev/internal/lib/](internal/lib/) | Library modules shared by the tests (e.g. kubernetes functions, test logging functions, API/CLI functions, etc) |

Note that this repo does NOT include the shell scripts that call the CT tests during the install. Those live in the repos for their respective components. For example, [/opt/cray/tests/crayctl-stage4/cms/ims_stage4_ct_tests.sh](https://stash.us.cray.com/projects/SCMS/repos/ims/browse/ct_tests/ims_stage4_ct_tests.sh) comes from [the IMS repo](https://stash.us.cray.com/projects/SCMS/repos/ims/browse).

## Example Command Usage
### CT Tests

```bash
cmsdev test cfs --ct
   # runs cfs ct tests at stage 4, default
cmsdev test conman --ct --crayctl-stage=1
   # runs conman ct tests at level crayctl stage 1
cmsdev test ims --ct --logs
   # runs ims ct tests with error logging enabled
     default=/opt/cray/tests/cmsdev.log, --output to override default
```

### API and Smoke Tests
```bash
cmsdev test ims --api
   # runs entire ims api test suite
cmsdev test ims --api --verbose
   # runs ims api tests with verbosity
cmsdev test bos --api sessions -v
   # runs bos api sessions endpoint tests with verbosity
cmsdev test ims --smoke --verbose
   # runs ims smoke tests with verbosity, includes ct tests
```


### Misc commands 
```bash
cmsdev ls services
   # lists the names of currently installed cms services
cmsdev ls bos --name
   # returns the service pod name of bos
cmsdev ls services --count
   # returms the number of currently installed cms services
cmsdev ls services --status
   # returns a list of cms service pods with status

cmsdev get bos endpoints
   # returns all bos endpoint descriptions
cmsdev get bos logs 
   # returns bos container logs 
cmsdev get bos sessiontemplate --endpoint
   # describe bos's sessiontemplate endpoint

cmsdev get k token --print
   # returns k8s access token
cmsdev get k client-secret --print
   # returns k8s client secret
```

## Contributing
Pull requests are welcome.
