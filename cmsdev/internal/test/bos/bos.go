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
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
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
	// section of the check has been simplified.

	// Check for:
	// All pods are expected to be Running except the migration pods, where
	// (if there are any) we require one of them to be Succeeded
	numMigrationPodsSucceeded := 0
	numMigrationPodsNotSucceeded := 0
	for _, podName := range podNames {
		common.Infof("Getting pod status for %s", podName)
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			continue
		} else {
			common.Infof("Pod %s status is %s", podName, status)
		}

		if strings.HasPrefix(podName, "cray-bos-migration-") {
			if status == "Succeeded" {
				numMigrationPodsSucceeded += 1
			} else {
				numMigrationPodsNotSucceeded += 1
			}
		} else if status != "Running" {
			common.VerboseFailedf("Pod %s has status %s, but we expect it to be Running", podName, status)
			passed = false
		}
	}

	if numMigrationPodsNotSucceeded > 0 && numMigrationPodsSucceeded == 0 {
		common.VerboseFailedf("There are %d cray-bos-migration pods, and none of them have status Succeeded", numMigrationPodsNotSucceeded)
		passed = false
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

	if !passed {
		common.ArtifactsKubernetes()
		if len(podNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, podNames)
		}
	}

	return
}
