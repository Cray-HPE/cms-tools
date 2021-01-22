package bos 

/*
 * bos_version.go
 * 
 * bos version tests  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
   	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// smoke tests
func versionTests(local bool) {
	var baseurl string = common.BASEURL
	const totalNumTests int = 1

	params := test.GetAccessTokenParams()
	if params == nil { return }

	numTests, numTestsFailed := 0, 0

	// scenario # 1, GET /version endpoint
	url := baseurl + endpoints["bos"]["version"].Url 
	numTests++
	test.RestfulTestHeader("GET /version", numTests, totalNumTests)
	_, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	}
	test.RestfulTestResultSummary(numTestsFailed, numTests)
}
