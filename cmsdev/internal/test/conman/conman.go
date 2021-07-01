package conman

/*
 * conman.go
 *
 * conman commons file
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 *
 * (MIT License)
 */

import (
	"regexp"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

var allPvcNames = []string{
	"cray-console-operator-data-claim",
	"cray-console-node-agg-data-claim",
}

var allPodNames = []string{}

func IsConmanRunning() (passed bool) {
	passed = true

	// check conman pvc status
	for _, pvcName := range allPvcNames {
		if !test.CheckPVCStatus(pvcName) {
			passed = false
		}
	}

	if !verifyConsoleDataPods() {
		passed = false
	}

	if !verifyConsoleNodePods() {
		passed = false
	}

	if !verifyConsoleOperatorPods() {
		passed = false
	}

	if !passed {
		common.ArtifactsPodsPvcs(allPodNames, allPvcNames)
	}
	return
}

// We want to verify that there are:
//      - exactly 1 main cray-conole-data- pod
//      - exactly 3 console-data-postgres-# pods (we allow just 1 or 2 below because that's what
//			we do in the VCS test, which makes me think there may be cases when we see fewer)
//      - 0 or more cray-console-data-wait-for-postgres-# pods
//      - at least 2 cray-console-node-# pods
//      - exactly 1 cray-console-operator pod
// All should be in Running state except for wait-for-postgres, which sould be Succeeded

func verifyConsoleDataPods() (passed bool) {
	passed = true
	// We expect 4 or more of these (one main pod, 3 postgres pods, and possibly 1+ wait-for-postgres pods)
	podNames, ok := test.GetPodNamesByPrefixKey("console-data", 4, -1)
	if !ok {
		passed = false
	}
	allPodNames = append(allPodNames, podNames...)
	common.Infof("Found %d console-data pods: %s", len(podNames), strings.Join(podNames, ", "))
	postgresRe := regexp.MustCompile("^cray-console-data-postgres-[0-9][0-9]*$")
	waitForPostgresRe := regexp.MustCompile("^cray-console-data-wait-for-postgres-[0-9][0-9]*-")
	mainPodCount := 0
	postgresPodCount := 0
	for _, podName := range podNames {
		expectedStatus := "Running"
		if waitForPostgresRe.MatchString(podName) {
			expectedStatus = "Succeeded"
		} else if postgresRe.MatchString(podName) {
			postgresPodCount += 1
			// There should be a corresponding pvc
			pvcName := "pgdata-" + podName
			common.Infof("There should be a corresponding pvc with name %s", pvcName)
			allPvcNames = append(allPvcNames, pvcName)
			if !test.CheckPVCStatus(pvcName) {
				passed = false
			}
		} else {
			mainPodCount += 1
		}
		common.Infof("checking pod status for %s expecting %s", podName, expectedStatus)
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			continue
		}
		common.Infof("Pod status is %s", status)
		if status != expectedStatus {
			common.VerboseFailedf("expected status=%s, found status=%s for podName=%s", expectedStatus, status, podName)
			passed = false
		} else {
			common.VerboseOkay()
		}
	}
	if mainPodCount != 1 {
		common.VerboseFailedf("Expected exactly 1 main console-data pod but found %d", mainPodCount)
		passed = false
	}
	if postgresPodCount < 1 || postgresPodCount > 3 {
		common.VerboseFailedf("Expected between 1 and 3 console-data-postgres pods but found %d", postgresPodCount)
		passed = false
	}
	return
}

func verifyConsoleNodePods() (passed bool) {
	passed = true
	// We expect at least 2 of these
	podNames, ok := test.GetPodNamesByPrefixKey("console-node", 2, -1)
	if !ok {
		passed = false
	}
	allPodNames = append(allPodNames, podNames...)
	common.Infof("Found %d console-node pods: %s", len(podNames), strings.Join(podNames, ", "))
	consoleNodeRe := regexp.MustCompile("^cray-console-node-[0-9][0-9]*$")
	mainPodCount := 0
	for _, podName := range podNames {
		expectedStatus := "Running"
		if consoleNodeRe.MatchString(podName) {
			mainPodCount += 1
		} else {
			common.VerboseFailedf("Unexpected pod name format for console-nod pod: %s", podName)
			passed = false
		}
		common.Infof("checking pod status for %s expecting %s", podName, expectedStatus)
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			continue
		}
		common.Infof("Pod status is %s", status)
		if status != expectedStatus {
			common.VerboseFailedf("expected status=%s, found status=%s for podName=%s", expectedStatus, status, podName)
			passed = false
		} else {
			common.VerboseOkay()
		}
	}
	if mainPodCount < 2 {
		common.VerboseFailedf("Expected at least 2 main console-node pods but found %d", mainPodCount)
		passed = false
	}
	return
}

func verifyConsoleOperatorPods() (passed bool) {
	passed = true
	// We expect exactly one of these
	podNames, ok := test.GetPodNamesByPrefixKey("console-operator", 1, 1)
	if !ok {
		passed = false
	}
	allPodNames = append(allPodNames, podNames...)
	common.Infof("Found %d console-operator pods: %s", len(podNames), strings.Join(podNames, ", "))
	for _, podName := range podNames {
		expectedStatus := "Running"
		common.Infof("checking pod status for %s expecting %s", podName, expectedStatus)
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			continue
		}
		common.Infof("Pod status is %s", status)
		if status != expectedStatus {
			common.VerboseFailedf("expected status=%s, found status=%s for podName=%s", expectedStatus, status, podName)
			passed = false
		} else {
			common.VerboseOkay()
		}
	}
	// No need to keep or check a count like for the other two types, because there is only the one type of console-operator-
	// pod expected.
	return
}
