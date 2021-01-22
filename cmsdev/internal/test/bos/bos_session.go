package bos

/*
 * bos_session.go
 *
 * bos session tests
 *
 * Copyright 2019-2020 Hewlett Packard Enterprise Development LP
 */

import (
	"encoding/json"
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// session tests
func sessionTestsAPI() bool {
	var baseurl string = common.BASEURL
	const totalNumTests int = 1

	numTests, numTestsFailed := 0, 0
	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}

	// test #1, list session
	url := baseurl + endpoints["bos"]["session"].Url
	numTests++
	test.RestfulTestHeader("GET session", numTests, totalNumTests)
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	} else {
		// Validate that at least we can decode the JSON into a list
		var m interface{}

		if err := json.Unmarshal(resp.Body(), &m); err != nil {
			common.Error(err)
			numTestsFailed++
		} else {
			_, ok := m.([]interface{})
			if !ok {
				common.Errorf("JSON response object not a list")
				numTestsFailed++
			}
		}
		// TODO: deeper validation of returned response
	}

	test.RestfulTestResultSummary(numTestsFailed, numTests)
	return numTestsFailed == 0
}

func sessionTestsCLI() bool {
	// test #1, list session
	common.Infof("Getting list of all BOS sessions via CLI")
	cmdOut := test.RunCLICommand("cray bos v1 session list --format json")
	if cmdOut == nil {
		return false
	}
	// Validate that at least we can decode the JSON into a list
	var m interface{}

	if err := json.Unmarshal(cmdOut, &m); err != nil {
		common.Error(err)
		return false
	}

	_, ok := m.([]interface{})
	if !ok {
		common.Errorf("JSON response object not a list")
		return false
	}

	// TODO: deeper validation of returned response
	return true
}
