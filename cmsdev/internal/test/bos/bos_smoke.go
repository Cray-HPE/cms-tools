package bos

/*
 * bos_smoke.go
 *
 * bos smoke tests
 *
 * Copyright 2019-2020 Hewlett Packard Enterprise Development LP
 */

import (
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// smoke tests
func smokeTests() {
	var baseurl string = common.BASEURL
	const totalNumTests int = 1

	numTests, numTestsFailed := 0, 0
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}

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
