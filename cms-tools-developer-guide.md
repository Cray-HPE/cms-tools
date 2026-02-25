# cms-tools Developer Guide

## Overview

The `cms-tools` repository contains test utilities for CMS (Cray Management Services) components. It includes two main components:

1. **`cmsdev`** (Go) -- A CLI tool that validates the health and functionality of CMS services including BOS, CFS, IMS, VCS, console services, and iPXE/TFTP.
2. **`cmstools`** (Python) -- A Python package providing integration tests including the barebones image boot test and the CFS sessions race condition test.

Both components are packaged into the `cray-cmstools-crayctldeploy` RPM and installed on NCNs (Non-Compute Nodes).

## Repository Structure

```
cms-tools/
+-- cmsdev/                          # Go-based cmsdev test tool
|   +-- internal/
|   |   +-- cmd/                     # Cobra command implementations (root, test, version)
|   |   +-- lib/                     # Shared libraries
|   |   |   +-- cms/                 # CMS service helpers (service data, S3 ops)
|   |   |   +-- common/              # Core utilities (constants, API client, logging, shell exec)
|   |   |   +-- k8s/                 # Kubernetes client operations (pods, PVCs, secrets, tokens)
|   |   |   +-- test/                # Test helper functions (API, CLI, pod/PVC validation)
|   |   |   +-- prod-catalog-utils/  # Product catalog parsing utilities
|   |   +-- test/                    # Service-specific test implementations
|   |       +-- bos/                 # BOS tests
|   |       +-- cfs/                 # CFS tests
|   |       +-- conman/              # Console tests
|   |       +-- ims/                 # IMS tests
|   |       +-- ipxe_tftp/           # iPXE/TFTP tests
|   |       +-- vcs/                 # VCS/Gitea tests
|   +-- vendor/                      # Vendored Go dependencies
+-- python-venv/                     # Python-based cmstools package
|   +-- cmstools/
|   |   +-- lib/                     # Shared Python libraries
|   |   |   +-- api/                 # API gateway client (auth, requests with retry)
|   |   |   +-- bos/                 # BOS API helpers
|   |   |   +-- cfs/                 # CFS API helpers
|   |   |   +-- hsm/                 # HSM API helpers
|   |   |   +-- ims/                 # IMS API helpers
|   |   |   +-- k8s/                 # Kubernetes operations (secrets, deployments)
|   |   |   +-- prodcat/             # Product catalog helpers
|   |   |   +-- s3/                  # S3 operations
|   |   +-- test/                    # Python test implementations
|   |       +-- barebones_image_test/ # Barebones image boot test
|   |       +-- cfs_sessions_rc_test/ # CFS sessions race condition test
+-- cms-tftp/                        # TFTP upload scripts
+-- tmp/                             # Reference documentation
+-- docs/                            # Additional documentation
```

## Technology Stack

### Go (cmsdev)

| Component | Details |
|-----------|---------|
| **Language** | Go 1.24.0 (toolchain go1.24.1) |
| **Module path** | `stash.us.cray.com/SCMS/cms-tools/cmsdev` |
| **CLI framework** | cobra v1.7.0 with viper v1.16.0 |
| **HTTP client** | resty v1.12.0 |
| **Kubernetes client** | client-go v0.24.17 |
| **Logging** | logrus v1.9.3 |
| **Colored output** | fatih/color v1.17.0 |
| **TFTP** | pin/tftp v2.1.0 |

### Python (cmstools)

| Component | Details |
|-----------|---------|
| **Language** | Python >= 3.11 |
| **Package name** | `cmstools` |
| **Build system** | `setuptools` via `pyproject.toml` |
| **HTTP client** | `requests` with `requests_retry_session` |
| **Kubernetes** | `kubernetes` Python client |
| **Entry points** | `barebones_image_test`, `cfs_sessions_rc_test` |

## Building the Project

### Prerequisites

- Go 1.24.0 or later
- Python 3.11 or later
- `make`
- Docker (for containerized/RPM builds)

### Makefile Targets

```bash
# Run build preparation (version stamping, etc.)
make runbuildprep

# Lint Go code
make lint

# Build the cmsdev binary (cross-compiled for Linux amd64)
# CGO_ENABLED=0 GO111MODULE=on GOARCH=amd64 GOOS=linux go build
make build_cmsdev

# Build the Python virtual environment
# PY_VERSION defaults to 3.11
make build_python_venv

# Package into RPM (expects cmsdev binary and Python venv to already be built)
make rpm
```

**CI Pipeline Order** (`Jenkinsfile.github`): `runbuildprep` → `lint` → (parallel: `build_cmsdev` + `build_python_venv`) → `rpm`

**Note**: The Go build produces a statically-linked Linux binary (`CGO_ENABLED=0`). The `@VERSION@` placeholder in `version.go` is replaced at build time by the build preparation step.

### Installation Paths

After RPM installation:
- `cmsdev` binary: `/usr/local/bin/cmsdev`
- Barebones image test: `/opt/cray/tests/integration/csm/barebones_image_test`
- CFS race condition test: `/opt/cray/tests/integration/csm/cfs_sessions_rc_test`
- Shell wrapper scripts: `/opt/cray/tests/install/ncn/scripts/`

## Key Components -- Go (cmsdev)

### Core Constants (`common.go`)

```go
const BASEHOST = "api-gw-service-nmn.local"
const BASEURL = "https://" + BASEHOST
const NAMESPACE string = "services"
const API_TIMEOUT_SECONDS = 120 * time.Second
const API_RETRY_COUNT = 3
const API_RETRY_WAIT_SECONDS = 5
```

### CMS Services List

```go
var CMSServices = []string{
    "bos", "cfs", "conman", "gitea", "ims", "ipxe", "tftp", "vcs",
}

// Aliases (gitea->vcs, ipxe->tftp)
var CMSServicesDuplicates = []string{"gitea", "ipxe"}
```

### Test Execution Flow

The main test driver is in `cmsdev/internal/cmd/test.go`:

```go
func RunTest(service string, includeCLI, includeTenant bool) bool {
    switch service {
    case "bos":
        return bos.IsBOSRunning(includeCLI, includeTenant)
    case "cfs":
        return cfs.IsCFSRunning(includeCLI, includeTenant)
    case "conman":
        return con.IsConmanRunning()
    case "ims":
        return ims.IsIMSRunning(includeCLI)
    case "ipxe", "tftp":
        return ipxe_tftp.AreTheyRunning()
    case "vcs", "gitea":
        return vcs.IsVCSRunning()
    }
    // ...
}
```

### Test Timeouts

All listed services use 300 seconds; the default for unlisted services is 120 seconds:

```go
var TestTimeouts = map[string]int64{
    "bos": 300, "cfs": 300, "conman": 300,
    "ipxe": 300, "tftp": 300, "vcs": 300, "gitea": 300,
}
const DefaultTestTimeout int64 = 120
```

### Retry Logic

When `--retry` is specified, `DoTestWithRetry` implements exponential backoff:
- Initial sleep: 5 seconds
- Increases by 5 seconds each attempt
- Capped at 1/6 of the test timeout
- Truncated as the timeout approaches

### Service Test Signatures

| Service | Function | Parameters |
|---------|----------|------------|
| BOS | `bos.IsBOSRunning` | `(includeCLI, includeTenant bool)` |
| CFS | `cfs.IsCFSRunning` | `(includeCLI, includeTenant bool)` |
| Conman | `con.IsConmanRunning` | `()` -- no optional test phases |
| IMS | `ims.IsIMSRunning` | `(includeCLI bool)` -- no tenant tests |
| iPXE/TFTP | `ipxe_tftp.AreTheyRunning` | `()` -- no optional test phases |
| VCS | `vcs.IsVCSRunning` | `()` -- no optional test phases |

### Logging System (`printlog.go`)

```go
const DEFAULT_LOG_FILE_DIR string = "/opt/cray/tests/install/logs/cmsdev"
```

Log files are organized in timestamped directories:

```
/opt/cray/tests/install/logs/cmsdev/
+-- YYMMDD_HHMMSS_microseconds_PID/
    +-- cmsdev.log          # Main test log
    +-- artifacts.tgz       # Failure artifacts (if test failed)
```

The `CreateLogFile` function creates a timestamped subdirectory using the format `YYMMDD_HHMMSS_microseconds_PID`.

> **Note:** For a comparison of the current (>= 1.34.0) vs legacy (< 1.34.0) logging structure and how to access logs, see [Log Structure Comparison](cms-tools-issue-triage-guide.md#log-structure-comparison) in the Issue Triage Guide.

#### Log Level Functions

```go
common.Debugf("Debug message: %s", value)     // DEBUG level
common.Infof("Info message: %s", value)        // INFO level
common.Warnf("Warning message: %s", value)     // WARN level
common.Errorf("Error message: %s", value)       // ERROR level

// Test outcome messages
common.Successf("Test passed")                  // Green SUCCESS prefix
common.Failuref("Test failed: %s", reason)      // Red FAILURE prefix
common.VerboseOkayf("check passed")             // Green OK (verbose mode only)
common.VerboseFailedf()                         // Red FAILED (verbose mode only)
```

### Artifact Collection on Failure

When a service test fails, artifacts are collected automatically. The `kubernetesThingsToCollect` list defines what K8s resources are captured:

```go
var kubernetesThingsToCollect = []string{
    "nodes", "namespaces", "pods", "pv", "pvc", "services",
    "daemonsets", "statefulsets", "deployments", "etcd",
    "configmaps", "secrets", "endpoints", "postgresqls",
    "cronjobs", "jobs", "sealedsecrets", "etcdbackups",
}
```

Additionally, RPM versions are captured for: `craycli`, `docs-csm`, `csm-testing`, `goss-servers`.

### Kubernetes Operations (`k8s/k8s.go`)

Key functions in the `k8s` package:
- `GetClientset()` -- Creates K8s client from KUBECONFIG
- `GetPodNames(namespace, prefix)` -- Lists pods matching a name prefix
- `GetPodStats(namespace, podName)` -- Gets pod phase and container states
- `GetPVCNames(namespace, prefix)` / `GetPVCStatus(namespace, pvcName)` -- PVC operations
- `GetAccessToken()` -- Gets OAuth2 token via Keycloak (`admin-client` credentials)
- `GetVcsUsernamePassword()` -- Gets VCS credentials from `vcs-user-credentials` K8s secret
- `RunCommandInContainer()` -- Executes commands inside pods via `kubectl exec`
- `GetTenants()` -- Lists tenants from the `tenants` namespace

K8s retry constants: `MaxRetries = 45`, `RetryIntervalSeconds = 2`.

### HTTP Client (`common.go`)

```go
func Restful(method, url string, params Params) (*resty.Response, error) {
    client := resty.New()
    client.SetTimeout(API_TIMEOUT_SECONDS)       // 120 seconds
    client.SetRetryCount(API_RETRY_COUNT)         // 3 retries
    client.SetRetryWaitTime(API_RETRY_WAIT_SECONDS * time.Second) // 5 seconds between retries
    // Auto-retry on HTTP 503
    // ...
}
```

A `RestfulTenant()` variant adds the `Cray-Tenant-Name` header for multi-tenancy.

## Key Components -- Python (cmstools)

### Package Configuration (`pyproject.toml`)

```toml
[project]
name = "cmstools"
version = "@RPM_VERSION@"
requires-python = ">=3.11"

[project.scripts]
barebones_image_test = "cmstools.test.barebones_image_test.__main__:main"
cfs_sessions_rc_test = "cmstools.test.cfs_sessions_rc_test.__main__:main"
```

### Python Logging (`lib/log.py`)

```python
DEFAULT_LOG_DIR = "/opt/cray/tests/integration/logs/csm/cmstools"
```

### API Client (`lib/api/api.py`)

```python
API_GW_DNSNAME = "api-gw-service-nmn.local"
API_GW_SECURE = f"https://{API_GW_DNSNAME}"
API_BASE_URL = f"{API_GW_SECURE}/apis"
```

## Adding a New Go Service Test

1. **Create test directory**: `cmsdev/internal/test/<service_name>/`

2. **Implement test function** following existing patterns (see BOS or CFS test for reference)

3. **Add pod name prefix** to `PodServiceNamePrefixes` in `common.go`

4. **Update the `RunTest` switch** in `test.go`

5. **Add to `CMSServices`** in `common.go`

6. **Add timeout** (if different from default 120s) in `test.go`

## Configuration

### Command-Line Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--log-dir` | | Custom base log directory | `/opt/cray/tests/install/logs/cmsdev` |
| `--no-cleanup` | | Do not remove temporary test files | `false` |
| `--no-log` | | Disable logging to file | `false` |
| `--retry` | `-r` | Retry on failure with backoff | `false` |
| `--quiet` | `-q` | Quiet mode (minimal output) | `false` |
| `--verbose` | `-v` | Verbose mode (detailed output) | `false` |
| `--list` | `-l` | List available service tests | |
| `--exclude-aliases` | | Exclude aliases from `--list` output | `false` |
| `--include-cli` | | Include CLI tests | `false` |
| `--include-tenant` | | Include tenant tests | `false` |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Path to kubeconfig file | `$HOME/.kube/config` |
| `IMS_RECIPE_NAME` | IMS recipe name to verify during IMS tests | Unset (skips verification) |
| `IMS_RECIPE_DISTRO` | Distribution of the IMS recipe being validated | `sles15` |
| `CONSOLE_LOG_LEVEL` | Python test console output level | `INFO` |
| `FILE_LOG_LEVEL` | Python test file output level | `DEBUG` |

## Debugging

### Enable Verbose Logging

```bash
/usr/local/bin/cmsdev test <service> --verbose
```

### Preserve Test Artifacts

```bash
/usr/local/bin/cmsdev test <service> --no-cleanup
```

### Review Log Files

```bash
# Find latest log directory
ls -lt /opt/cray/tests/install/logs/cmsdev/ | head -2

# View logs
less /opt/cray/tests/install/logs/cmsdev/<timestamp>/cmsdev.log
```

### Common Issues

1. **Kubernetes permissions**: Ensure proper RBAC roles for the test user
2. **Cray CLI not configured**: Run `cray init` on the NCN
3. **Service pods not ready**: Check with `kubectl get pods -n services`
4. **API endpoint unreachable**: Verify gateway at `https://api-gw-service-nmn.local`
5. **TFTP test skipped on masters**: This is expected; TFTP file transfer only runs on worker NCNs

## Related Documentation

- [cmsdev Tests](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cmsdev_tests.md)
- [Barebones Image Boot Test](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cms_barebones_image_boot.md)
- [CFS Race Condition Test](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cfs_sessions_race_condition_test.md)
- [Configure Cray CLI](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/operations/configure_cray_cli.md)

## Contributing

1. Follow existing code structure and patterns in each test package
2. Add comprehensive logging using `common.Infof/Debugf/Errorf` (Go) or `logger` (Python)
3. Include error handling and resource cleanup
4. Add artifact collection on failure
5. Test with `--retry`, `--verbose`, and `--no-cleanup` flags
6. Update this documentation and the test suites doc for new features