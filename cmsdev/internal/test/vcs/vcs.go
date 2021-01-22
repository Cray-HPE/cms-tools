package vcs

/*
 * vcs.go
 *
 * vcs commons file
 *
 * Copyright 2019-2020 Hewlett Packard Enterprise Development LP
 */

import (
	"regexp"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

func IsVCSRunning(local, smoke, ct bool, crayctlStage string) (passed bool) {
	passed = true
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return
	case "4", "5":
		// We expect to find 2 or more gitea-vcs pods
		podNames, ok := test.GetPodNamesByPrefixKey("vcs", 2, -1)
		if !ok {
			passed = false
		}
		common.Infof("Found %d vcs pods: %s", len(podNames), strings.Join(podNames, ", "))

		pvcNames := make([]string, 0, 4)
		pvcNames = append(pvcNames, "gitea-vcs-data-claim")
		if !test.CheckPVCStatus(pvcNames[0]) {
			passed = false
		}
		// In addition to the main gitea-vcs pod (which should be in Running state), we expect:
		// - 1 or more gitea-vcs-postgres pod (in Running state)
		// - 0 or more gitea-vcs-wait-for-postgres-<num>- pods (in Succeeded state)
		listedPodsOnError := false
		mainPodCount := 0
		postgresPodCount := 0
		waitForPostgresRe := regexp.MustCompile("gitea-vcs-wait-for-postgres-[0-9][0-9]*-")
		for _, podName := range podNames {
			expectedStatus := "Running"
			if waitForPostgresRe.MatchString(podName) {
				expectedStatus = "Succeeded"
			} else if strings.HasPrefix(podName, "gitea-vcs-postgres-") {
				postgresPodCount += 1
				// There should be a corresponding pvc
				pvcName := "pgdata-" + podName
				common.Infof("There should be a corresponding pvc with name %s", pvcName)
				pvcNames = append(pvcNames, pvcName)
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
				if !listedPodsOnError {
					common.Printf("Found %d vcs pods: %s", len(podNames), strings.Join(podNames, ", "))
					listedPodsOnError = true
				}
				common.VerboseFailedf("expected status=%s, found status=%s for podName=%s", expectedStatus, status, podName)
				passed = false
			} else {
				common.VerboseOkay()
			}
		}
		if mainPodCount != 1 {
			if !listedPodsOnError {
				common.Printf("Found %d vcs pods: %s", len(podNames), strings.Join(podNames, ", "))
				listedPodsOnError = true
			}
			common.VerboseFailedf("Expected exactly 1 main gitea-vcs- pod but found %d", mainPodCount)
			passed = false
		}
		if postgresPodCount < 1 {
			if !listedPodsOnError {
				common.Printf("Found %d vcs pods: %s", len(podNames), strings.Join(podNames, ", "))
				listedPodsOnError = true
			}
			common.VerboseFailedf("Expected at least 1 gitea-vcs-postgres pod but found %d", postgresPodCount)
			passed = false
		}

		if !passed {
			common.ArtifactsPodsPvcs(podNames, pvcNames)
		}

		if !repoTest() {
			passed = false
		}

		return
	default:
		common.Errorf("Invalid stage for this test")
		passed = false
		return
	}
	common.Errorf("Programming logic error: this line should never be reached")
	passed = false
	return
}
