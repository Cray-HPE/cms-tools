package vcs

/*
 * vcs.go
 *
 * vcs commons file
 *
 * © Copyright 2019-2020 Hewlett Packard Enterprise Development LP
 */

import (
	"regexp"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

func IsVCSRunning(local, smoke, ct bool, crayctlStage string) bool {
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return true
	case "4", "5":
		// We expect to find 2 or more gitea-vcs pods
		podNames, ok := test.GetPodNamesByPrefixKey("vcs", 2, -1)
		if !ok {
			return false
		}
		common.Infof("Found %d vcs pods", len(podNames))

		// In addition to the main gitea-vcs pod (which should be in Running state), we expect:
		// - 1 or more gitea-vcs-postgres pod (in Running state)
		// - 0 or more gitea-vcs-wait-for-postgres-<num>- pods (in Succeeded state)
		mainPodCount := 0
		postgresPodCount := 0
		passed := true
		waitForPostgresRe := regexp.MustCompile("gitea-vcs-wait-for-postgres-[0-9][0-9]*-")
		for _, podName := range podNames {
			expectedStatus := "Running"
			if waitForPostgresRe.MatchString(podName) {
				expectedStatus = "Succeeded"
			} else if strings.HasPrefix(podName, "gitea-vcs-postgres-") {
				postgresPodCount += 1
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
			common.VerboseFailedf("Expected exactly 1 main gitea-vcs- pod but found %d", mainPodCount)
			passed = false
		}
		if postgresPodCount < 1 {
			common.VerboseFailedf("Expected at least 1 gitea-vcs-postgres pod but found %d", postgresPodCount)
			passed = false
		}

		// check vcs pvc pod status
		// We expect 1 for the main pod and 1 per postgres pod
		expectedPvcCount := postgresPodCount + mainPodCount
		if !test.CheckServicePVCStatusByPrefixKey("vcs", expectedPvcCount, expectedPvcCount) {
			passed = false
		}
		return passed
	default:
		common.Errorf("Invalid stage for this test")
		return false
	}
	common.Errorf("Programming logic error: this line should never be reached")
	return false
}
