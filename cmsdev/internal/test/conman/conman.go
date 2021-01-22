package conman

/*
 * conman.go
 *
 * conman commons file
 *
 * Copyright 2019-2020 Hewlett Packard Enterprise Development LP
 */

import (
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

var pvcNames = []string{
	"cray-conman-data-claim",
}

func IsConmanRunning(local, smoke, ct bool, crayctlStage string) (passed bool) {
	passed = true
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return
	case "4", "5":
		var expectedContainerStatus string
		if crayctlStage == "4" {
			expectedContainerStatus = "CrashLoopBackOff"
		} else {
			expectedContainerStatus = "Running"
		}

		podNames, ok := test.GetPodNamesByPrefixKey("conman", 1, 1)
		if !ok {
			passed = false
		}
		for _, podName := range podNames {
			if !test.CheckPodStats(podName, "cray-conman", expectedContainerStatus) {
				passed = false
			}
		}

		// check conman pvc status
		for _, pvcName := range pvcNames {
			if !test.CheckPVCStatus(pvcName) {
				passed = false
			}
		}

		if !passed {
			common.ArtifactsPodsPvcs(podNames, pvcNames)
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
