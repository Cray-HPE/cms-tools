# cms-tools Test Suites

## Overview

The `cms-tools` repository contains automated test suites for validating CMS (Cray Management Services) functionality. These tests are implemented in both Python and Go, covering critical workflows including image boot operations, configuration management, and service health validation.

## Test Suite Categories

### Python Tests

Python-based tests focus on end-to-end workflows and integration testing for specific CMS operations.

#### 1. Barebones Image Boot Test

**Purpose**: Validates the complete boot workflow using a minimal barebones image, testing the integration of BOS (Boot Orchestration Service), IMS (Image Management Service), and BSS (Boot Script Service).

**Entry point**: `cmstools.test.barebones_image_test.__main__:main`
**Script location**: `/opt/cray/tests/integration/csm/barebones_image_test`
**Source**: `python-venv/cmstools/test/barebones_image_test/__main__.py` (~450 lines)

#### 2. CFS Sessions Race Condition Test

**Purpose**: Tests CFS API behavior under concurrent access patterns, verifying proper handling of simultaneous DELETE and GET operations against CFS session endpoints.

**Entry point**: `cmstools.test.cfs_sessions_rc_test.__main__:main`
**Script location**: `/opt/cray/tests/integration/csm/cfs_sessions_rc_test`
**Source**: `python-venv/cmstools/test/cfs_sessions_rc_test/`

### Go Tests

Go-based tests perform comprehensive health checks on individual CMS services.

#### 3. cmsdev Test Suite

**Purpose**: Validates the health and functionality of all CMS services including pod status, API endpoints, CLI operations, and multi-tenancy support.

**Binary location**: `/usr/local/bin/cmsdev`
**Source**: `cmsdev/`

---

## 1. Barebones Image Boot Test -- Detailed

### Test Workflow

The test executes the following steps in order:

1. Parse arguments and setup logging
2. Query product catalog for CSM data
3. Create/select IMS base image
4. Create CFS configuration
5. Build customized image via IMS+CFS
6. Create BOS session template
7. Boot target compute node via BOS
8. Wait for node boot and validate
9. Cleanup resources (unless --no-cleanup)

### Step Details

1. **Product Catalog Lookup**: Reads the `cray-product-catalog` ConfigMap from the `services` namespace to find the CSM version, base image ID, git commit, and VCS clone URL.

2. **IMS Image Creation**: Either uses a provided `--base-id` or looks up the base image from the product catalog. If `--id` is provided, image customization is skipped entirely.

3. **CFS Configuration**: Creates a CFS configuration with the specified playbook (`compute_nodes.yml` by default) pointing to the VCS repository. If `--cfs-config` is provided, uses an existing configuration.

4. **Image Customization**: Creates a CFS session targeting the base image to produce a customized image with the specified Ansible playbook applied.

5. **BOS Session Template**: Creates a BOS v2 session template referencing the customized image and CFS configuration.

6. **Node Boot**: Creates a BOS session to boot the target compute node. The node is selected automatically from HSM (Hardware State Manager) unless `--xname` is specified.

7. **Validation**: Waits for the node to boot and validates it reaches the expected state.

8. **Cleanup**: Removes all created resources (BOS session template, CFS configuration, IMS images, etc.) unless `--no-cleanup` is specified.

### Command-Line Arguments

```bash
barebones_image_test [OPTIONS]
```

| Argument | Description | Default |
|----------|-------------|---------|
| `--arch` | Architecture: `x86` or `arm` | `x86` |
| `--csm-version` | CSM product catalog version to use | Latest found |
| `--base-id` | IMS image ID for the base image | From product catalog |
| `--id` | Pre-customized IMS image ID (skips CFS customization) | None |
| `--cfs-config` | Name of existing CFS configuration | None (creates new) |
| `--vcs-url` | Git clone URL for Ansible content | From product catalog |
| `--git-commit` | Git commit hash for VCS content | From product catalog |
| `--playbook` | Ansible playbook filename | `compute_nodes.yml` |
| `--xname` | Target compute node xname | Auto-selected from HSM |
| `--no-cleanup` | Do not delete created resources after test | `false` |

### Usage Examples

```bash
# Basic run (auto-detect everything from product catalog)
/opt/cray/tests/integration/csm/barebones_image_test

# Specify architecture
/opt/cray/tests/integration/csm/barebones_image_test --arch arm

# Use a pre-existing customized image (skip CFS step)
/opt/cray/tests/integration/csm/barebones_image_test --id <image-id>

# Use specific node and keep resources for debugging
/opt/cray/tests/integration/csm/barebones_image_test --xname x3000c0s19b1n0 --no-cleanup

# Use specific CSM version
/opt/cray/tests/integration/csm/barebones_image_test --csm-version 1.6.0
```

### Log Output

Logs are written to:
```
/opt/cray/tests/integration/logs/csm/cmstools/barebones_image_test/YYYYMMDD_HHMMSS.log
```

Console log level defaults to `INFO`; file log level defaults to `DEBUG`. These can be changed via environment variables:
```bash
export CONSOLE_LOG_LEVEL=DEBUG
export FILE_LOG_LEVEL=INFO
```

### Common Failure Modes

| Failure | Cause | Resolution |
|---------|-------|------------|
| "No CSM entry found in product catalog" | CSM not installed or catalog corrupt | `kubectl -n services get cm cray-product-catalog -o jsonpath='{.data.csm}'` |
| "No enabled compute nodes found" | No nodes available for boot test | `cray hsm state components list --type Node --role Compute --enabled true` |
| CFS session stuck | Ansible playbook failure or git access issue | Check CFS session logs, verify VCS access |
| BOS session never completes | Node hardware or networking issues | Check BOS operator logs, node console output |
| Cleanup failure | API errors during resource deletion | Use `--no-cleanup` to inspect; manually delete resources |

### Key Source Files

| File | Purpose |
|------|---------|
| `python-venv/cmstools/test/barebones_image_test/__main__.py` | Main test workflow (~450 lines) |
| `python-venv/cmstools/lib/api/api.py` | API gateway client with auth |
| `python-venv/cmstools/lib/bos/` | BOS API helpers |
| `python-venv/cmstools/lib/cfs/` | CFS API helpers |
| `python-venv/cmstools/lib/ims/` | IMS API helpers |
| `python-venv/cmstools/lib/hsm/` | HSM API helpers |
| `python-venv/cmstools/lib/prodcat/` | Product catalog helpers |
| `python-venv/cmstools/lib/log.py` | Log file setup |

---

## 2. CFS Sessions Race Condition Test -- Detailed

### Test Workflow

1. Parse arguments and setup logging
2. Get API access token
3. Scale CFS operator to 0 replicas
4. For each subtest:
   a. Create CFS sessions
   b. Send concurrent API requests
   c. Validate responses
   d. Clean up sessions
5. Restore CFS operator to 1 replica
6. Restore CFS options (if modified)

For detailed design information, see [CASMCMS-9544](https://jira-pro.it.hpe.com:8443/browse/CASMCMS-9544).

### Why Scale CFS Operator to 0?

The CFS operator normally processes sessions as they are created, which would interfere with race condition testing. Scaling it to 0 ensures that sessions remain in their `pending` state during the concurrent API operations.

**Critical**: If the test is interrupted before cleanup, the CFS operator will remain at 0 replicas. Recovery:
```bash
kubectl get deployment -n services cray-cfs-operator
kubectl scale deployment -n services cray-cfs-operator --replicas=1
```

### Subtests

The test includes 6 subtests that exercise different concurrent access patterns:

| Subtest | Description | Sessions Created |
|---------|-------------|-----------------|
| `single_delete` | Multiple parallel DELETE requests targeting the same session | 1 |
| `multi_delete` | Multiple parallel DELETE requests, each targeting different sessions | `--max-sessions` (default 20) |
| `single_delete_single_get` | Parallel single DELETE + single GET on same session | 1 |
| `single_delete_multi_get` | Parallel single DELETE + batch GET on same session | 1 |
| `multi_delete_single_get` | Parallel batch DELETE + single GET across sessions | `--max-sessions` (default 20) |
| `multi_delete_multi_get` | Parallel batch DELETE + batch GET across sessions | `--max-sessions` (default 20) |

Each subtest uses Python's `concurrent.futures.ThreadPoolExecutor` with `--max-parallel-requests` threads (default 4).

### Command-Line Arguments

```bash
cfs_sessions_rc_test [OPTIONS]
```

| Argument | Description | Default |
|----------|-------------|---------|
| `--cfs-version` | CFS API version (`v2` or `v3`) | `v3` |
| `--max-sessions` | Maximum number of CFS sessions to create | `20` |
| `--max-multi-delete-reqs` | Maximum number of parallel multi-delete requests | `4` |
| `--max-multi-get-reqs` | Maximum number of parallel multi-get requests | `4` |
| `--max-single-delete-reqs` | Maximum number of parallel single-delete requests | `4` |
| `--max-single-get-reqs` | Maximum number of parallel single-get requests | `4` |
| `--name` | Prefix for all session names | `cfs-race-condition-test` |
| `--page-size` | Page size for multi-get requests (default: 10 × max-sessions for v3; min=max-sessions for v2, min=1 for v3) | `None` (auto-calculated) |
| `--delete-previous-sessions` | Delete any existing pending sessions with the specified name prefix | `false` |
| `--run-subtests` | Comma-separated list of subtests to run (mutually exclusive with `--skip-subtests`) | All 6 subtests |
| `--skip-subtests` | Comma-separated list of subtests to skip (mutually exclusive with `--run-subtests`) | None |

### Usage Examples

```bash
# Run all subtests with defaults
/opt/cray/tests/integration/csm/cfs_sessions_rc_test

# Run specific subtests
/opt/cray/tests/integration/csm/cfs_sessions_rc_test --subtests single_delete,multi_delete

# Use CFS v2 API
/opt/cray/tests/integration/csm/cfs_sessions_rc_test --cfs-version v2

# Increase parallelism
/opt/cray/tests/integration/csm/cfs_sessions_rc_test --max-parallel-requests 8 --max-sessions 50
```

### Log Output

Logs are written to:
```
/opt/cray/tests/integration/logs/csm/cmstools/cfs_sessions_rc_test/YYYYMMDD_HHMMSS.log
```

### Common Failure Modes

| Failure | Cause | Resolution |
|---------|-------|------------|
| CFS operator not scaled back | Test interrupted before cleanup | `kubectl scale deployment -n services cray-cfs-operator --replicas=1` |
| Unexpected HTTP status codes | Real race condition bug in CFS API | Report to CFS team with exact status codes and request patterns |
| CFS v2 page-size modified | Test modifies `default-page-size` for v2 | `cray cfs v3 options update --default-page-size 1000` |
| Session creation failures | CFS API overloaded or down | Check CFS API pod status and logs |
| Timeout during concurrent ops | Network or CFS performance issues | Review network stability, CFS resource allocation |

### Key Source Files

| File | Purpose |
|------|---------|
| `python-venv/cmstools/test/cfs_sessions_rc_test/__main__.py` | Main orchestration (~167 lines) |
| `python-venv/cmstools/test/cfs_sessions_rc_test/argument_parser.py` | Argparse definitions (~140 lines) |
| `python-venv/cmstools/test/cfs_sessions_rc_test/defs.py` | ScriptArgs NamedTuple, defaults |
| `python-venv/cmstools/test/cfs_sessions_rc_test/subtests/` | 6 subtest modules + base class |

---

## 3. cmsdev Test Suite -- Detailed

### Architecture

The cmsdev binary dispatches tests to service-specific test functions via the `RunTest` switch in `cmd/test.go`.

### Test Driver (`cmd/test.go`)

The `RunTest` function dispatches to service-specific test functions:

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

### Service Test Signatures

| Service | Function | Parameters | Notes |
|---------|----------|------------|-------|
| BOS | `bos.IsBOSRunning` | `(includeCLI, includeTenant bool)` | `includeCLI` controls whether CLI tests run; `includeTenant` controls whether tenant tests run |
| CFS | `cfs.IsCFSRunning` | `(includeCLI, includeTenant bool)` | `includeCLI` controls whether CLI tests run; `includeTenant` controls whether tenant tests run |
| Conman | `con.IsConmanRunning` | `()` | No optional test phases |
| IMS | `ims.IsIMSRunning` | `(includeCLI bool)` | `includeCLI` controls whether CLI tests run; no tenant tests |
| iPXE/TFTP | `ipxe_tftp.AreTheyRunning` | `()` | No optional test phases |
| VCS | `vcs.IsVCSRunning` | `()` | No optional test phases |

### Test Timeouts

```go
var TestTimeouts = map[string]int64{
    "bos": 300, "cfs": 300, "conman": 300,
    "ipxe": 300, "tftp": 300, "vcs": 300, "gitea": 300,
}
const DefaultTestTimeout int64 = 120
```

### Retry Logic

When `--retry` is enabled, `DoTestWithRetry()` implements exponential backoff:
- Initial sleep: 5 seconds
- Increases by 5 seconds each attempt
- Capped at 1/6 of the test timeout
- Truncated as timeout approaches

### Command-Line Interface

```bash
cmsdev test <service> [flags]
```

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

**Note**: `--quiet` and `--verbose` are mutually exclusive.

### Available Services

```bash
# List all services
cmsdev test --list

# Output:
# bos cfs conman gitea ims ipxe tftp vcs

# List without aliases (gitea, ipxe)
cmsdev test --list --exclude-aliases

# Output:
# bos cfs conman ims tftp vcs
```

Service aliases: `gitea` -> `vcs`, `ipxe` -> `tftp`

### Usage Examples

```bash
# Test BOS with verbose output
/usr/local/bin/cmsdev test bos --verbose

# Test CFS with CLI and tenant tests
/usr/local/bin/cmsdev test cfs --include-cli --include-tenant

# Test all CMS services with retry
for svc in bos cfs conman ims tftp vcs; do
    /usr/local/bin/cmsdev test $svc --retry
done

# Test with custom log directory
/usr/local/bin/cmsdev test ims --log-dir /tmp/my-logs

# Test without writing log files
/usr/local/bin/cmsdev test bos --no-log --verbose
```

---

## Service Test Details

### BOS (Boot Orchestration Service)

**Function**: `bos.IsBOSRunning(includeCLI, includeTenant bool) bool`
**Source**: `cmsdev/internal/test/bos/bos.go` (~146 lines)
**Pod prefix**: `cray-bos`
**Minimum pods**: 3

**Test Steps**:
1. Verify at least 3 pods with prefix `cray-bos` are running
2. Check migration pod status is `Succeeded`
3. Test API endpoints:
   - `GET /apis/bos/v2/healthz` -- Health check
   - `GET /apis/bos/v2/version` -- Version endpoint
   - CRUD operations on session templates
   - CRUD operations on sessions
   - CRUD operations on components
4. (Optional) CLI tests via `cray bos v2 ...` commands
5. (Optional) Tenant-scoped API tests

### CFS (Configuration Framework Service)

**Function**: `cfs.IsCFSRunning(includeCLI, includeTenant bool) bool`
**Source**: `cmsdev/internal/test/cfs/cfs.go` (~157 lines)
**Pod prefix**: `cray-cfs`
**Minimum pods**: 2

**Test Steps**:
1. Verify at least 2 pods with prefix `cray-cfs` are running
2. Test API endpoints:
   - `GET /apis/cfs/v3/healthz` -- Health check
   - `GET /apis/cfs/v3/version` -- Version endpoint
   - CRUD on configurations
   - CRUD on sessions
   - CRUD on sources
   - CRUD on components
   - GET/PUT on options
3. Validate product catalog integration
4. (Optional) CLI tests
5. (Optional) Tenant-scoped API tests

### IMS (Image Management Service)

**Function**: `ims.IsIMSRunning(includeCLI bool) bool`
**Source**: `cmsdev/internal/test/ims/ims.go` (~419 lines)
**Pod prefix**: `cray-ims`
**Minimum pods**: 1
**PVC**: `cray-ims-data-claim`

**Test Steps**:
1. Verify 1 pod with prefix `cray-ims` is running
2. Verify PVC `cray-ims-data-claim` is Bound
3. Test API endpoints:
   - `GET /apis/ims/v3/version` -- Version endpoint
   - CRUD on images (create, list, get, patch, delete)
   - CRUD on recipes (create, list, get, patch, delete)
   - CRUD on public-keys (create, list, get, delete)
4. Test remote image builds:
   - Upload recipe to S3
   - Create IMS job for remote build
   - Monitor job completion
   - Verify build artifacts
5. (Optional) CLI tests
6. Optionally validates IMS recipe name via `IMS_RECIPE_NAME` environment variable

### Console Services (Conman)

**Function**: `con.IsConmanRunning() bool`
**Source**: `cmsdev/internal/test/conman/conman.go` (~222 lines)

**Test Steps**:
1. Verify `cray-console-operator` pods are running
2. Verify `cray-console-node` pods are running
3. Verify PVCs are Bound:
   - `cray-console-operator-data-claim`
   - `cray-console-node-0-data-claim`
4. Test console API access

### VCS (Version Control Service / Gitea)

**Function**: `vcs.IsVCSRunning() bool`
**Source**: `cmsdev/internal/test/vcs/vcs.go` (~198 lines)
**Pod prefix**: `gitea-vcs`
**Minimum pods**: 2

**Test Steps**:
1. Verify at least 2 pods with prefix `gitea-vcs` are running
2. Get VCS credentials from `vcs-user-credentials` K8s secret
3. Test VCS API:
   - Create test repository
   - Create file in repository
   - List repositories
   - Delete test repository
4. Verify PVC bound status

### iPXE/TFTP

**Function**: `ipxe_tftp.AreTheyRunning() bool`
**Source**: `cmsdev/internal/test/ipxe_tftp/ipxe_tftp.go` (~138 lines)

**Test Steps**:
1. Verify BSS iPXE pods per architecture:
   - `cray-bss-ipxe-aarch64` pods
   - `cray-bss-ipxe-x86-64` pods
2. Verify `cray-tftp` pod is running
3. Test TFTP file upload and download operations

**Note**: TFTP file transfer tests only run on worker NCNs (`ncn-w*`). On master nodes, this subtest is skipped.

---

## Shared Infrastructure

### Go Test Helpers (`lib/test/`)

| Module | Functions | Purpose |
|--------|-----------|---------|
| `api.go` | `GetAccessToken()`, `RestfulVerifyStatus()`, `TenantRestfulVerifyStatus()` | API auth and request validation |
| `cli.go` | `RunCLICommand()`, `RunCLICommandJSON()` | Cray CLI execution (path: `/usr/bin/cray`) |
| `test.go` | `GetPodNamesInNamespace()`, `CheckPodStats()`, `CheckPVCStatus()` | K8s resource validation |

### Go HTTP Client (`lib/common/common.go`)

```go
const API_TIMEOUT_SECONDS = 120 * time.Second
const API_RETRY_COUNT = 3
const API_RETRY_WAIT_SECONDS = 5
```

The `Restful()` function auto-retries on HTTP 503. The `RestfulTenant()` variant adds the `Cray-Tenant-Name` header for multi-tenancy.

### Python API Client (`lib/api/api.py`)

```python
API_GW_DNSNAME = "api-gw-service-nmn.local"
API_GW_SECURE = f"https://{API_GW_DNSNAME}"
API_BASE_URL = f"{API_GW_SECURE}/apis"
```

Auth is obtained from the `admin-client-auth` K8s secret using Keycloak token exchange.

### Logging

| Component | Default Log Directory | Format |
|-----------|----------------------|--------|
| cmsdev (Go) | `/opt/cray/tests/install/logs/cmsdev/` | `YYMMDD_HHMMSS_microseconds_PID/cmsdev.log` |
| Python tests | `/opt/cray/tests/integration/logs/csm/cmstools/<test_name>/` | `YYYYMMDD_HHMMSS.log` |

> **Note:** For detailed logging structure differences between cmsdev >= 1.34.0 (per-run directories) and legacy versions (single shared log file), see [Log Structure Comparison](cms-tools-issue-triage-guide.md#log-structure-comparison) in the Issue Triage Guide.

### Artifact Collection (cmsdev)

On test failure, cmsdev automatically collects:

**Kubernetes resources**: nodes, namespaces, pods, pv, pvc, services, daemonsets, statefulsets, deployments, etcd, configmaps, secrets, endpoints, postgresqls, cronjobs, jobs, sealedsecrets, etcdbackups

**Per-service artifacts**:
- Pod descriptions with events
- Pod logs (all containers, timestamped)
- PVC descriptions with events

**RPM versions**: craycli, docs-csm, csm-testing, goss-servers

Artifacts are packaged as `artifacts.tgz` in the log directory.

---

## Running Tests

### Prerequisites

- NCN (master or worker) with `cray-cmstools-crayctldeploy` RPM installed
- Configured Cray CLI (`cray init`)
- Kubernetes access (KUBECONFIG)
- For barebones test: at least one enabled compute node

### Shell Wrapper Scripts

The RPM installs wrapper scripts at `/opt/cray/tests/install/ncn/scripts/`:

```bash
# Run cmsdev tests (wrapper)
/opt/cray/tests/install/ncn/scripts/run_cmstools_test.sh

# Run barebones image boot test
/opt/cray/tests/install/ncn/scripts/barebones_image_test.sh

# Run CFS race condition test
/opt/cray/tests/install/ncn/scripts/cfs_sessions_rc_test.sh
```

### Direct Execution

```bash
# cmsdev -- test a single service
/usr/local/bin/cmsdev test bos --verbose --retry

# cmsdev -- test all services
for svc in bos cfs conman ims tftp vcs; do
    /usr/local/bin/cmsdev test $svc --retry --verbose
done

# Barebones image boot
/opt/cray/tests/integration/csm/barebones_image_test

# CFS race condition
/opt/cray/tests/integration/csm/cfs_sessions_rc_test
```

### Environment Variables

| Variable | Description | Default | Applies To |
|----------|-------------|---------|------------|
| `KUBECONFIG` | Path to kubeconfig file | `$HOME/.kube/config` | All tests |
| `IMS_RECIPE_NAME` | IMS recipe name to verify | Unset (skips) | cmsdev IMS test |
| `IMS_RECIPE_DISTRO` | IMS recipe distribution | `sles15` | cmsdev IMS test |
| `CONSOLE_LOG_LEVEL` | Console output log level | `INFO` | Python tests |
| `FILE_LOG_LEVEL` | File output log level | `DEBUG` | Python tests |

## Related Documentation

- [cms-tools Developer Guide](cms-tools-developer-guide.md) -- Repository structure, build process, code walkthrough
- [cms-tools Issue Triage Guide](cms-tools-issue-triage-guide.md) -- Debugging failures, log analysis, escalation
- [cmsdev Tests (docs-csm)](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cmsdev_tests.md)
- [Barebones Image Boot (docs-csm)](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cms_barebones_image_boot.md)
- [CFS Race Condition Test (docs-csm)](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cfs_sessions_race_condition_test.md)