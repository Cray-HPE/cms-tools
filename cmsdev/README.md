# CMSDEV Tool

cmsdev is a test utility for CMS services. The tool executes tests for all CMS services. Additionally, cmsdev has been padded with additional functionality to make CMS development and system information gathering easier.

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

For all CMS components, the test includes a basic check for kubernetes pods. Most also include some calls to the service itself via API and CLI. These tests should absolutely be expanded, while keeping in mind that these are intended to be quick, non-destructive tests.

On an installed system, these are important files and directories:

| File or Directory | Description |
| ------------------|-------------|
| /usr/local/bin/cmsdev | The man himself |
| /opt/cray/tests/cmsdev.log | Detailed log file (not to be mistaken with the output from when it is run). [See an example file](examples/cmsdev.log) |
| /usr/local/bin/cmslogs | Utility to collect most of these files to help debug test failures. [See the README here](../cmslogs) for more details |

Noteworthy files in the repo:

| File or Directory | Description |
| ------------------|-------------|
| cmsdev/cmsdev | cmsdev binary |
| [cmsdev/internal/cmd/test.go](internal/cmd/test.go) | Main test driver |
| [cmsdev/internal/test/](internal/test/) | Every CMS component which is tested has a directory here that contains all test code |
| [cmsdev/internal/lib/](internal/lib/) | Library modules shared by the tests (e.g. kubernetes functions, test logging functions, API/CLI functions, etc) |

## Example Command Usage
### Tests

```bash
cmsdev test cfs
   # runs cfs tests
cmsdev test conman -q -r
   # runs conman tests quietly, retrying on failure
cmsdev test ims -v --no-log
   # runs ims tests verbosely with logging disabled
     default=/opt/cray/tests/cmsdev.log, --output to override default
IMS_RECIPE_NAME=uan-recipe cmsdev test ims --no-log -v
   # same as previous, but also verifies that an IMS recipe with the specified name exists
   # (with distro type of sles15, by default)
IMS_RECIPE_NAME=uan-recipe IMS_RECIPE_DISTRO=centos cmsdev test ims --no-log -v
   # same as previous, but with the distro type specified as well
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
cmsdev get ims logs 
   # returns ims container logs 
cmsdev get bos sessiontemplate --endpoint
   # describe bos's sessiontemplate endpoint

cmsdev get k token --print
   # returns k8s access token
cmsdev get k client-secret --print
   # returns k8s client secret
```

## Contributing
Pull requests are welcome.
