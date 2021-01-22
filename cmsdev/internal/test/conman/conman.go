package conman 

/*
 * conman.go
 * 
 * conman commons file  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

var expectedStatus string = "Running"

func IsConmanRunning(local, smoke, ct bool, crayctlStage string) bool {
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return true
	case "4", "5":
		// check pod and container status
		podNames, ok := test.GetPodNamesByPrefixKey("conman", 1, 1)
		if !ok {
			return false
		} else if crayctlStage == "4" {
			expectedStatus = "CrashLoopBackOff"
		}
		if !test.CheckPodStats(podNames[0], true, "cray-conman", expectedStatus) {
			return false
		}
		// check conman pvc pod status
		return test.CheckPVCStatusByPrefixKey("conmanPvc")
	default:
		common.Errorf("Invalid stage for this test")
		return false
	}
	common.Errorf("Programming logic error: this line should never be reached")
	return false
}
