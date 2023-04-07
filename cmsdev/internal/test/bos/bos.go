// MIT License
//
// (C) Copyright 2019-2023 Hewlett Packard Enterprise Development LP
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

func IsBOSRunning() (passed bool) {
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
	// All pods are expected to be Running except for the following, which are expected to
	// be Succeeded if they exist:
	// - cray-bos-wait-for-etcd
	etcPodCount := 0
	waitForEtcdRe := regexp.MustCompile("cray-bos-wait-for-etcd-[0-9][0-9]*-")
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

		expectedStatus := "Running"
		if waitForEtcdRe.MatchString(podName) {
			expectedStatus = "Succeeded"
		}
		if strings.HasPrefix(podName, "cray-bos-etcd-") {
			etcPodCount += 1
			// There should be a corresponding pvc with the same name
			common.Infof("There should be a corresponding pvc with the same name as this pod (%s)", podName)
			pvcNames = append(pvcNames, podName)
			if !test.CheckPVCStatus(podName) {
				passed = false
			}
		}
		if status != "" && status != expectedStatus {
			common.VerboseFailedf("Pod %s has status %s, but we expect it to be %s", podName, status, expectedStatus)
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

	// Defined in bos_api.go
	if !apiTests() {
		passed = false
	}

	// Defined in bos_cli.go
	if !cliTests() {
		passed = false
	}

	return
}
