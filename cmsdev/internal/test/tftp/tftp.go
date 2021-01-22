package tftp

/*
 * tftp.go
 * 
 * tftp commons file  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

func IsTFTPRunning(local, smoke, ct bool, crayctlStage string) bool {
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return true
	case "4", "5":
		// check pod and container status
		if !test.CheckServicePodStatsByPrefixKey("tftp", -1, -1) {
			return false
		}

		// check cray-tftp pvc pod status
		return test.CheckPVCStatusByPrefixKey("tftpPvc")
	default:
		common.Errorf("Invalid stage for this test")
		return false
	}
	common.Errorf("Programming logic error: this line should never be reached")
	return false
}
