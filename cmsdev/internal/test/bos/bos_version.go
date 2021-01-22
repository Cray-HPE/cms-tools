package bos

/*
 * bos_version.go
 *
 * bos version tests
 *
 * Copyright 2019-2020 Hewlett Packard Enterprise Development LP
 */

import (
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// smoke tests
func versionTests() bool {
	var baseurl string = common.BASEURL
	const totalNumTests int = 1

	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}

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
	return numTestsFailed == 0
}
