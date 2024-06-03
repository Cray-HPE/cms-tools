// MIT License
//
// (C) Copyright 2019-2024 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
package bos

/*
 * bos.go
 *
 * bos commons file
 *
 */

import (
	"regexp"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

// For tests which need a tenant name, if they can work even with a nonexistent tenant, use the
// following name if no actual tenants exist
const defaultTenantName = "cmsdev-tenant"

// Return a random tenant from the list
// If the list is empty, return defaultTenantName
func getAnyTenant(tenantList []string) string {
	if len(tenantList) > 0 {
		if tenantName, err := common.GetRandomStringFromList(tenantList); err == nil {
			common.Debugf("Randomly selected tenant name '%s'", tenantName)
			return tenantName
		}
	}
	common.Debugf("Returning default tenant name '%s'", defaultTenantName)
	return defaultTenantName
}

func IsBOSRunning() (passed bool) {
	var err error
	var tenantList []string

	passed = true

	// Look for at least 3 bos pods, although we know there are more.
	podNames, ok := test.GetPodNamesByPrefixKey("bos", 3, -1)
	if !ok {
		passed = false
	}
	common.Infof("Found %d bos pods", len(podNames))
	// The BOS pod zoo has become rather unruly. Parts of this test were
	// passing more by accident than anything else. Given that, this
	// section of the check has been simplified to just make sure that there
	// are 3 etcd pods, and that any transient pods, if they exist, have succeeded.

	// Check for:
	// Exactly 3 BOS etcd pods
	etcPodCount := 0
	etcdRe := regexp.MustCompile("cray-bos-bitnami-etcd-[0-9]")
	pvcNames := make([]string, 0, 3)
	for _, podName := range podNames {
		common.Infof("Getting pod status for %s", podName)
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			status = ""
		} else {
			common.Infof("Pod %s status is %s", podName, status)
		}

		if etcdRe.MatchString(podName) {
			etcPodCount += 1
			// There should be a corresponding pvc with the same name prepended with "data-"
			pvcName := "data-" + podName
			common.Infof("There should be a corresponding pvc for this pod (%s) with the name '%s'", podName, pvcName)
			pvcNames = append(pvcNames, pvcName)
			if !test.CheckPVCStatus(pvcName) {
				passed = false
			}
		}

		if status != "" && !podHasExpectedStatus(podName, status, etcdRe) {
			passed = false
		}
	}
	if etcPodCount != 3 {
		common.VerboseFailedf("Found %d cray-bos-etcd- pod(s), but expect to find 3", etcPodCount)
		passed = false
	}

	if !passed {
		common.ArtifactsKubernetes()
		if len(podNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, podNames)
		}
		if len(pvcNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, pvcNames)
		}
	}

	// Get list of defined tenants on the system (if any)
	tenantList, err = k8s.GetTenants()
	if err != nil {
		common.VerboseFailedf(err.Error())
		passed = false
		// Set tenantList to an empty list -- some tenant tests can still run even if no tenants are known
		tenantList = []string{}
	}

	// Defined in bos_api.go
	if !apiTests(tenantList) {
		passed = false
	}

	// Defined in bos_cli.go
	if !cliTests(tenantList) {
		passed = false
	}

	return
}

var optionalCompletedPodNames = []string{
	"cray-bos-wait-for-etcd-",
	"cray-bos-etcd-post-install-",
	"cray-bos-pre-upgrade-etcd-backup-",
}

const succeededStatus = "Succeeded"
const pendingStatus = "Pending"
const runningStatus = "Running"

func podHasExpectedStatus(podName, status string, etcdRe *regexp.Regexp) (passed bool) {
	// All pods are expected to be Running except...
	//
	// the following, which are expected to be Succeeded if they exist:
	// - cray-bos-wait-for-etcd
	// - cray-bos-etcd-post-install
	// - cray-bos-pre-upgrade-etcd-backup
	//
	// and the following, which may be Pending, Running, or Succeeded, if it exists:
	// - cray-bos-bitnami-etcd-snapshotter

	expectedStatus := ""
	if etcdRe.MatchString(podName) {
		expectedStatus = runningStatus
	} else {
		for _, podPrefix := range optionalCompletedPodNames {
			if strings.HasPrefix(podName, podPrefix) {
				// These pods, if they exist, are expected to be Succeeded
				expectedStatus = succeededStatus
				break
			}
		}
		if expectedStatus == "" && !strings.HasPrefix(podName, "cray-bos-bitnami-etcd-snapshotter-") {
			expectedStatus = runningStatus
		}
	}

	// expectedStatus will be set except in the case where this is a cray-bos-bitnami-etcd-snapshotter- pod.
	if expectedStatus != "" {
		if status != expectedStatus {
			common.VerboseFailedf("Pod %s has status %s, but we expect it to be %s", podName, status, expectedStatus)
			return false
		}
		return true
	}

	// Now it only remains to check the cray-bos-bitnami-etcd-snapshotter-, which we allow to be Pending, Running, or Succeeded,
	// since they run periodically, and it is possible for us to be executing this test while one is starting or underway
	if status != runningStatus && status != succeededStatus && status != pendingStatus {
		common.VerboseFailedf("Pod %s has status %s, but we expect it to be Pending, Running, or Succeeded", podName, status)
		return false
	}
	return true
}
