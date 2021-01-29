package conman

/*
 * conman.go
 *
 * conman commons file
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

var pvcNames = []string{
	"cray-conman-data-claim",
}

func IsConmanRunning() (passed bool) {
	passed = true

	podNames, ok := test.GetPodNamesByPrefixKey("conman", 1, 1)
	if !ok {
		passed = false
	}
	for _, podName := range podNames {
		if !test.CheckPodStats(podName, "cray-conman", "Running") {
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
}
