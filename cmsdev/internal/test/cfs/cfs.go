// MIT License
//
// (C) Copyright 2019-2025 Hewlett Packard Enterprise Development LP
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
package cfs

/*
 * cfs.go
 *
 * cfs commons file
 *
 */

import (
	"regexp"
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

// This value indicates if the CFS test should fail
// because of a problem with the product catalog.
var prodCatOk = true

// This setter function is called by a subtest if it
// encounters an error getting product catalog data.
func SetProdCatOk(ok bool) {
	prodCatOk = ok
}

func IsCFSRunning(includeCLI bool) (passed bool) {
	passed = true
	// 2 pods minimum since we expect both an api and operator pod
	podNames, ok := test.GetPodNamesByPrefixKey("cfs", 2, -1)
	if !ok {
		passed = false
	}
	common.Infof("Found %d cfs pods", len(podNames))
	apiPodName, operatorPodName := "", ""
	for _, podName := range podNames {
		// CFS is running if there is operator and api pod
		// we can ignore the state of cfs-jobs or cfs-db pods
		re := regexp.MustCompile(common.PodServiceNamePrefixes["cfsServices"])
		if re.MatchString(podName) {
			if strings.HasPrefix(podName, common.PodServiceNamePrefixes["cfs-api"]) {
				apiPodName = podName
				common.Infof("Found apiPod=%s", podName)
			} else if strings.HasPrefix(podName, common.PodServiceNamePrefixes["cfs-operator"]) {
				operatorPodName = podName
				common.Infof("Found operatorPod=%s", podName)
			}
		}
		if re.MatchString(podName) {
			common.Infof("checking pod status for %s expecting %s", podName, "Running")
		} else {
			common.Infof("checking pod status for %s expecting %s", podName, "N/A")
		}
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			continue
		}
		common.Infof("Pod status is %s", status)
		if re.MatchString(podName) {
			if status == "Succeeded" {
				common.Warnf("Pod %s has status %s", podName, status)
			} else if status != "Running" {
				common.VerboseFailedf("expected status=Running, found status=%s for podName=%s", status, podName)
				passed = false
			} else {
				common.VerboseOkay()
			}
		}
	}
	if len(apiPodName) == 0 {
		common.Errorf("No apiPod found")
		passed = false
	}
	if len(operatorPodName) == 0 {
		common.Errorf("No operatorPod found")
		passed = false
	}
	if !testCFSAPI() {
		passed = false
	}
	if !TestCFSConfigurationsCRUDOperationWithTenantsUsingAPIVersions() {
		passed = false
	}
	if !TestCFSSourcesCRUDOperation() {
		passed = false
	}

	// CLI tests will be run only if requested using the include-cli flag
	if includeCLI {
		if !testCFSCLI() {
			passed = false
		}
		if !TestCFSConfigurationsCRUDOperationWithTenantsUsingCLI() {
			passed = false
		}
		if !TestCFSSourcesCRUDOperationUsingCLI() {
			passed = false
		}
	}

	// Fail if any subtest got an error trying to get product catalog data.
	// This is not covered by the subtest itself, because it may have been
	// able to run successfully even without that data.
	passed = passed && prodCatOk

	if !passed {
		common.ArtifactsKubernetes()
		if len(podNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, podNames)
		}
	}
	return
}
