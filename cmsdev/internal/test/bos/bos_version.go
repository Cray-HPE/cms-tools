package bos

/*
 * bos_version.go
 *
 * bos version tests
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
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
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	} else {
		// Validate that object can be decoded into a string map at least
		_, err = common.DecodeJSONIntoStringMap(resp.Body())
		if err != nil {
			common.Error(err)
			numTestsFailed++
		}
	}
	test.RestfulTestResultSummary(numTestsFailed, numTests)
	return numTestsFailed == 0
}
