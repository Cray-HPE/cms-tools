package bos

/*
 * bos.go
 * 
 * bos commons file  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"strings"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()

// BOS supported API tests
var APITestTypes = []string{
	"post",
	"session",
	"sessiontemplate",
	"smoke",
	"version", 
}

func IsBOSRunning(local, smoke, ct bool, crayctlStage string) bool {
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return true
	case "4", "5":
		// check pod and container status
		if !test.CheckServicePodStatsByPrefixKey("bos", 5, 5) {
			return false
		}

		// check bos pvc pod status
		if !test.CheckServicePVCStatusByPrefixKey("bosPvc", 3, -1) {
			return false
		}

		if smoke {
			common.Verbosef("")
			smokeTests(local)
		}
		return true
	default:
		common.Errorf("Invalid stage for this test")
		return false
	}
	common.Errorf("Programming logic error: this line should never be reached")
	return false
}

// access to BOS API tests 
func RunAPITests(local bool, params ...string) {
	if IsBOSRunning(local, false, false, "4") == false {
		common.Warnf("aborting API tests, BOS service is not ready run cmsdev test bos --smoke -v")
		return
	}

	// TODO: ensure this works for local testing	
	if local {
		common.Warnf("--local currently not supported")
		return
	}

	// run all API tests
	if len(params) == 0 {
		smokeTests(local)
		postSessionTemplateTests(local)
		sessionTests(local)
		sessionTemplateTests(local)
		versionTests(local)
		return
	}
	switch params[0] {
	case "post":
		postSessionTemplateTests(local)
	case "session":
		sessionTests(local)
	case "sessiontemplate":
		sessionTemplateTests(local)
	case "smoke":
		smokeTests(local)
	case "version":
		versionTests(local)
	default:
		common.Warnf("--api argument required. current available tests are: %s", 
			strings.Join(APITestTypes, " "))
	}
}
