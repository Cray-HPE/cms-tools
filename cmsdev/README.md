# CMSDEV Tool

cmsdev is a test utility for CMS services. The tool executes tests for all CMS services. Additionally, cmsdev has been padded with additional functionality to make CMS development and system information gathering easier.

## Installation

Building requires the installation of [Golang](https://golang.org/doc/install).

Be sure to copy the repo to a directory OUTSIDE of `$GOPATH/src`. In this example, assume we have copied it into `/tmp/stash/cms-tools`

```bash
cd /tmp/stash/cms-tools/cmsdev
GOOS=linux GOARCH=amd64 go build -mod vendor .
```

### Noteworthy file/directory locations

For all CMS components, the test includes a basic check for Kubernetes pods. Most also include some calls to the service itself via
API and CLI. These tests should absolutely be expanded, while keeping in mind that these are intended to be quick, non-destructive
tests.

On an installed system, these are important files and directories:

| File or Directory | Description |
| ------------------|-------------|
| `/usr/local/bin/cmsdev` | The man himself |
| `/opt/cray/tests/install/logs/cmsdev/cmsdev.log` | Detailed log file (not to be mistaken with the output from when it is run). [See an example file](examples/cmsdev.log) |

Noteworthy files in the repository:

| File or Directory | Description |
| ------------------|-------------|
| [`cmsdev/internal/cmd/test.go`](internal/cmd/test.go) | Main test driver |
| [`cmsdev/internal/test`/](internal/test/) | Every CMS component which is tested has a directory here that contains all test code |
| [`cmsdev/internal/lib/`](internal/lib/) | Library modules shared by the tests (e.g. Kubernetes functions, test logging functions, API/CLI functions, etc) |

## Command Usage

Run the command with the `-h` flag for a usage statement.

```bash
cmsdev test -h
```

## Contributing

Pull requests are welcome.
