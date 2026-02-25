# cms-tools Issue Triage Guide

## Overview

This guide helps triage and debug issues with `cmsdev` automated tests, particularly focusing on CASMTRIAGE tickets generated from automated test runs.

## Identifying Test Failures

### Step 1: Locate the JIRA Ticket

Auto-triage tickets follow the pattern: `CASMTRIAGE-XXXX`

Example: [CASMTRIAGE-8823](https://jira-pro.it.hpe.com:8443/browse/CASMTRIAGE-8823)

### Step 2: Identify the Node

Check the ticket description or title to identify which NCN (Non-Compute Node) the test ran on.

Common node patterns:
- `ncn-w001` - Worker node 1
- `ncn-w002` - Worker node 2
- `ncn-m001` - Master node 1

### Step 3: Access Test Logs

#### For cmsdev >= 1.34.0 (Current Logging Structure)

1. **In JIRA ticket**: Look for hyperlink with text **"Test log output"**
2. **Click the link** to open the log viewer page
3. **Click "stdout" hyperlink** to view the actual test logs
4. **Identify the log directory** from the output:
   ```
   Starting main run, version: 1.34.0, log directory: /opt/cray/tests/install/logs/cmsdev/241012_050305_414367_990773
   ```
5. **Access logs on the node**:
   ```bash
   ssh ncn-w001  # or appropriate node
   cd /opt/cray/tests/install/logs/cmsdev/241012_050305_414367_990773
   less cmsdev.log
   ```
6. **Check for artifacts** (if test failed):
   ```bash
   # Artifacts are saved in the same directory
   ls -lh artifacts.tgz

   # Extract artifacts
   tar -xzf artifacts.tgz
   ls -R artifacts/
   ```

#### For cmsdev < 1.34.0 (Legacy Logging Structure)

1. **Access the single log file**:
   ```bash
   ssh ncn-w001
   less /opt/cray/tests/install/logs/cmsdev/cmsdev.log
   ```
2. **Find the run tag** from JIRA or log output:
   ```
   Creating temporary directory
   Starting sub-run, tag: 0P65L-bos
   ```
3. **Filter logs by run tag**:
   ```bash
   # Extract logs for specific run tag
   grep "run=0P65L" /opt/cray/tests/install/logs/cmsdev/cmsdev.log > /tmp/filtered-0P65L.log

   # For service-specific sub-run
   grep "run=0P65L-bos" /opt/cray/tests/install/logs/cmsdev/cmsdev.log > /tmp/bos-test.log
   ```

### Log Structure Comparison

| Aspect | Legacy (< 1.34.0) | Current (>= 1.34.0) |
|--------|-------------------|---------------------|
| **Log file** | Single shared `cmsdev.log` | Per-run directory with own `cmsdev.log` |
| **Directory** | `/opt/cray/tests/install/logs/cmsdev/` | `/opt/cray/tests/install/logs/cmsdev/<timestamp>/` |
| **Run isolation** | Run tag required to filter | Each run has its own directory |
| **Artifacts** | In `artifacts/` subdirectory per tag | `artifacts.tgz` in run directory |
| **Timestamp format** | N/A | `YYMMDD_HHMMSS_microseconds_PID` |

### Python Test Logs

Python tests write logs to a separate directory:

```
/opt/cray/tests/integration/logs/csm/cmstools/
+-- barebones_image_test/
|   +-- YYYYMMDD_HHMMSS.log
+-- cfs_sessions_rc_test/
    +-- YYYYMMDD_HHMMSS.log
```

```bash
# Find the latest barebones image test log
ls -lt /opt/cray/tests/integration/logs/csm/cmstools/barebones_image_test/ | head -5

# Find the latest CFS race condition test log
ls -lt /opt/cray/tests/integration/logs/csm/cmstools/cfs_sessions_rc_test/ | head -5
```

## Common Failure Patterns

### 1. Pod Not Running

**Log Pattern**:
```
ERROR: Expected 3 or more pods with prefix cray-bos, found 2
```
or
```
ERROR: Expected Running/Succeeded phase for pod cray-cfs-api-xxxx, found phase=CrashLoopBackOff
```

**Triage Steps**:
```bash
kubectl get pods -n services -l app.kubernetes.io/name=cray-bos
kubectl describe pod <failing-pod> -n services
kubectl logs <failing-pod> -n services --previous
```

### 2. API Endpoint Failures

**Log Pattern**:
```
ERROR: Unexpected status code 503 from GET https://api-gw-service-nmn.local/apis/bos/v2/healthz
```

**Triage Steps**:
```bash
# Check API gateway
curl -k https://api-gw-service-nmn.local/apis/

# Check service health directly
kubectl logs -n services -l app.kubernetes.io/name=cray-bos-api
```

### 3. PVC Not Bound

**Log Pattern**:
```
ERROR: Expected Bound status for pvc=cray-console-operator-data-claim, found status=Pending
```

**Triage Steps**:
```bash
kubectl get pvc -n services <pvc-name>
kubectl describe pvc -n services <pvc-name>
kubectl get pv | grep Available
kubectl get storageclass
```

### 4. Timeout Errors

**Log Pattern**:
```
ERROR: Test timeout after 300 seconds
```

Timeout values per service (from `test.go`):
- **BOS, CFS, conman, iPXE/TFTP, VCS/gitea**: 300 seconds
- **All other services (default)**: 120 seconds

**Triage Steps**:
- Check if retry was enabled: `--retry` flag
- Review service response times in logs
- Check for network issues or service overload
- Verify Kubernetes cluster health

### 5. CLI Test Failures

**Log Pattern**:
```
ERROR: Command failed: /usr/bin/cray bos v2 sessiontemplates list --format json
```

**Triage Steps**:
```bash
# Verify CLI is configured
cray init

# Check CLI configuration
cat /root/.config/cray/configurations/default

# Test CLI manually
cray bos v2 healthz list
```

### 6. Tenant Test Failures

**Log Pattern**:
```
ERROR: Unexpected status code from tenant-scoped request with Cray-Tenant-Name header
```

**Triage Steps**:
```bash
# Check tenants
kubectl get namespaces | grep tenant

# Verify tenant CRDs
kubectl get tenants -n tenants
```

## Service-Specific Triage

### BOS (Boot Orchestration Service)

**What the test checks**: At least 3 pods (`cray-bos` prefix), migration pod status (Succeeded), API CRUD (session templates, sessions, components), version/healthz endpoints.

**Common Issues**:
- Session template creation failures
- Component listing errors
- Migration pod not completing (should be `Succeeded`)

**Key Logs to Check**:
```bash
kubectl logs -n services -l app.kubernetes.io/name=cray-bos-api
kubectl logs -n services -l app.kubernetes.io/name=cray-bos-operator
```

**Related Files**: `cmsdev/internal/test/bos/`

### CFS (Configuration Framework Service)

**What the test checks**: At least 2 pods (`cray-cfs` prefix), API CRUD (configurations, sessions, sources, components, options), product catalog validation.

**Common Issues**:
- Configuration CRUD failures
- Product catalog access errors
- Tenant isolation issues

**Key Logs to Check**:
```bash
kubectl logs -n services -l app.kubernetes.io/name=cray-cfs-api
kubectl logs -n services -l app.kubernetes.io/name=cray-cfs-operator
```

**Related Files**: `cmsdev/internal/test/cfs/`

### IMS (Image Management Service)

**What the test checks**: 1 pod (`cray-ims` prefix), PVC bound status (`cray-ims-data-claim`), API CRUD (images, recipes, public-keys), remote image builds (S3 upload, job monitoring).

**Common Issues**:
- PVC not bound
- Image creation/deletion failures
- S3 upload errors
- Remote build job failures

**Key Logs to Check**:
```bash
kubectl logs -n services -l app.kubernetes.io/name=cray-ims
kubectl get pvc -n services cray-ims-data-claim
```

**Related Files**: `cmsdev/internal/test/ims/`

### Console Services (Conman)

**What the test checks**: Pod running, PVC bound (`cray-console-operator-data-claim`, `cray-console-node-0-data-claim`), console API access.

**Common Issues**:
- PVC not bound (frequent)
- Console operator pod not running
- Data claim issues

**Key Logs to Check**:
```bash
kubectl logs -n services -l app.kubernetes.io/name=cray-console-operator
kubectl get pvc -n services | grep console
```

**Related Files**: `cmsdev/internal/test/conman/`

### VCS (Version Control Service / Gitea)

**What the test checks**: At least 2 pods (`gitea-vcs` prefix), PVC bound status, VCS API (repository CRUD, file operations), authentication via K8s `vcs-user-credentials` secret.

**Common Issues**:
- Gitea database issues
- PVC problems
- Repository operations failing
- VCS credentials secret missing or invalid

**Key Logs to Check**:
```bash
kubectl logs -n services -l app.kubernetes.io/name=gitea-vcs
kubectl get secret -n services vcs-user-credentials
```

**Related Files**: `cmsdev/internal/test/vcs/`

### iPXE/TFTP

**What the test checks**: BSS iPXE pods (per-architecture: `aarch64` and `x86_64`), `cray-tftp` pod, TFTP file upload/download operations.

**Common Issues**:
- Architecture-specific pods missing
- TFTP transfer failures (only runs on worker NCNs)

**Key Logs to Check**:
```bash
kubectl get pods -n services | grep -E "ipxe|tftp"
```

**Note**: TFTP file transfer subtests only run on worker NCNs (`ncn-w*`). Failures on master nodes for this subtest are expected to be skipped.

**Related Files**: `cmsdev/internal/test/ipxe_tftp/`

## Barebones Image Boot Test Failures

**Script location**: `/opt/cray/tests/integration/csm/barebones_image_test`

**Test flow**: Product catalog lookup -> IMS image creation -> CFS image customization -> BOS session -> Node boot -> Validation -> Cleanup

**Common Failures**:

| Failure Point | Symptoms | Triage Steps |
|---------------|----------|--------------|
| Product catalog | "No CSM entry found" | `kubectl -n services get cm cray-product-catalog -o jsonpath='{.data.csm}'` |
| HSM node lookup | "No enabled compute nodes" | `cray hsm state components list --type Node --role Compute --enabled true` |
| CFS image customization | Session stuck or failed | Check CFS session status, Ansible logs |
| BOS boot | Session never completes | Check BOS operator logs, node console output |
| Cleanup | Resources left behind | Use `--no-cleanup` to inspect; manually delete |

**Key Arguments**: See [Barebones Image Boot Test -- Command-Line Arguments](cms-tools-test-suites.md#command-line-arguments) in the Test Suites Guide.

## CFS Race Condition Test Failures

**Script location**: `/opt/cray/tests/integration/csm/cfs_sessions_rc_test`

**Test flow**: Scale CFS operator to 0 -> create sessions -> run subtests (concurrent DELETE/GET) -> cleanup -> restore operator

**Critical recovery note**: If the test is interrupted, the CFS operator may remain scaled to 0:
```bash
# Check current replica count
kubectl get deployment -n services cray-cfs-operator

# Manually restore (typically 1 replica)
kubectl scale deployment -n services cray-cfs-operator --replicas=1
```

**Subtests**: See [CFS Sessions Race Condition Test -- Subtests](cms-tools-test-suites.md#subtests) in the Test Suites Guide.

**Key Arguments**: See [CFS Sessions Race Condition Test -- Command-Line Arguments](cms-tools-test-suites.md#command-line-arguments-1) in the Test Suites Guide.

**Common Failures**:
- CFS operator not scaling back up after test
- CFS v2 `page-size` option left modified (restore with `cray cfs v3 options update --default-page-size 1000`)
- Unexpected HTTP status codes from concurrent operations (indicates real race condition bugs)

## Artifact Collection Details

### Automatic Artifact Collection (cmsdev)

When a cmsdev service test fails, artifacts are collected automatically into `artifacts.tgz` in the log directory:

#### Kubernetes Resource Dumps

The following K8s resources are collected from the `services` namespace:
- nodes, namespaces, pods, pv, pvc, services
- daemonsets, statefulsets, deployments, etcd
- configmaps, secrets, endpoints, postgresqls
- cronjobs, jobs, sealedsecrets, etcdbackups

#### Per-Service Artifacts

For each failed service, detailed information is collected via `ArtifactDescribeNamespacePods()`:
- `kubectl describe pod -n services <pod> --show-events=true` -- for each service pod
- `kubectl logs -n services <pod> --all-containers=true --timestamps=true --prefix=true` -- for each service pod
- `kubectl describe pvc -n services <pvc> --show-events=true` -- for each service PVC

#### RPM Versions

RPM query output for: `craycli`, `docs-csm`, `csm-testing`, `goss-servers`

### Manual Artifact Collection

```bash
# Gather all pods status
kubectl get pods --all-namespaces -o wide > /tmp/all-pods.txt

# Gather PVC status
kubectl get pvc --all-namespaces > /tmp/all-pvcs.txt

# Gather recent events
kubectl get events --all-namespaces --sort-by='.lastTimestamp' > /tmp/events.txt

# Gather cmsdev logs and artifacts
cd /opt/cray/tests/install/logs/cmsdev/<timestamp>
tar -czf /tmp/cmsdev-evidence.tgz cmsdev.log artifacts.tgz

# Gather Python test logs
ls -lt /opt/cray/tests/integration/logs/csm/cmstools/*/
```

### Migrate Legacy Logs

If you need to convert legacy logs to new format:

```bash
# Using default directory
/usr/local/bin/convert_cmsdev_logs.sh

# Using custom directory
/usr/local/bin/convert_cmsdev_logs.sh /path/to/custom/logdir
```

This script (see `convert_cmsdev_logs.sh`):
1. Scans for unique run tags
2. Creates timestamped directories
3. Extracts logs per run tag
4. Moves associated artifacts
5. Removes original log file

## Resolution Workflow

### 1. Categorize the Issue

- **Infrastructure**: Pod failures, PVC issues, node problems
- **Service**: API errors, functionality failures
- **Test**: Test logic errors, false positives
- **Environment**: Configuration, network, permissions

### 2. Collect Evidence

```bash
# Gather system state
kubectl get pods --all-namespaces -o wide > /tmp/all-pods.txt
kubectl get pvc --all-namespaces > /tmp/all-pvcs.txt
kubectl get events --all-namespaces --sort-by='.lastTimestamp' > /tmp/events.txt

# Service-specific info
kubectl describe pod <failing-pod> -n services > /tmp/pod-describe.txt
kubectl logs <failing-pod> -n services > /tmp/pod-logs.txt
```

### 3. Determine Root Cause

| Category | Common Causes | Resolution Path |
|----------|---------------|-----------------|
| **Infrastructure** | Node issues, storage problems, network issues | System administration |
| **Service** | Service bugs, configuration errors, dependency failures | Service team escalation |
| **Test** | Test assumptions invalid, timing issues, environment changes | Update test code |
| **Environment** | Wrong versions, missing configuration, RBAC issues | Configuration management |

### 4. Verify Resolution

```bash
# Re-run the specific service test
/usr/local/bin/cmsdev test <service> --verbose

# Re-run Python test
/opt/cray/tests/integration/csm/barebones_image_test
/opt/cray/tests/integration/csm/cfs_sessions_rc_test
```

## Quick Reference Commands

### Service Health Checks

```bash
# Check API Gateway
curl -k https://api-gw-service-nmn.local/apis/

# BOS health
curl -k https://api-gw-service-nmn.local/apis/bos/v2/healthz

# CFS health
curl -k https://api-gw-service-nmn.local/apis/cfs/v3/healthz

# IMS version (requires auth)
cray ims versions list

# VCS health
curl -k https://api-gw-service-nmn.local/vcs/api/v1/version
```

### Kubernetes Debugging

```bash
kubectl get pods -n services -o wide              # All service pods
kubectl logs <pod-name> -n services                # Pod logs
kubectl logs <pod-name> -n services --previous     # Previous pod logs (if crashed)
kubectl describe pod <pod-name> -n services        # Pod details
kubectl get events -n services --sort-by='.lastTimestamp'  # Recent events
kubectl get pvc -n services                        # PVC status
kubectl exec -it <pod-name> -n services -- /bin/sh # Exec into pod
```

### CFS Operator Recovery (after interrupted race condition test)

```bash
# Check current state
kubectl get deployment -n services cray-cfs-operator

# Restore if scaled to 0
kubectl scale deployment -n services cray-cfs-operator --replicas=1

# If CFS v2 page-size was modified, restore via CLI
cray cfs v3 options update --default-page-size 1000
```

### Log Analysis

```bash
# Find latest cmsdev test run
ls -lt /opt/cray/tests/install/logs/cmsdev/ | head -5

# Search for errors in cmsdev logs
grep -i error /opt/cray/tests/install/logs/cmsdev/<timestamp>/cmsdev.log

# Find latest Python test log
ls -lt /opt/cray/tests/integration/logs/csm/cmstools/barebones_image_test/
ls -lt /opt/cray/tests/integration/logs/csm/cmstools/cfs_sessions_rc_test/
```

## Additional Resources

- [cmsdev Tests Documentation](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cmsdev_tests.md)
- [Barebones Image Boot Test](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cms_barebones_image_boot.md)
- [CFS Race Condition Test](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/cfs_sessions_race_condition_test.md)
- [SMS Health Check Known Issues](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/troubleshooting/known_issues/sms_health_check.md)
- [Validate CSM Health](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/operations/validate_csm_health.md)
- [Configure Cray CLI](https://github.com/Cray-HPE/docs-csm/blob/release/1.7/operations/configure_cray_cli.md)