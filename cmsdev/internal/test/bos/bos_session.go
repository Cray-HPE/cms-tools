package bos 

/*
 * bos_session.go
 * 
 * bos session tests  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// session tests
func sessionTests(local bool) {
	var baseurl string = common.BASEURL
	const totalNumTests int = 1

	numTests, numTestsFailed := 0, 0
	params := test.GetAccessTokenParams()
	if params == nil { return }

	// test #1, list session
	url := baseurl + endpoints["bos"]["session"].Url
	numTests++
	test.RestfulTestHeader("GET session", numTests, totalNumTests)
	_, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	}

	// TODO: deeper validation of returned response
	test.RestfulTestResultSummary(numTestsFailed, numTests)
}
