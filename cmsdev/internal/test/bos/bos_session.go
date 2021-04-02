package bos

/*
 * bos_session.go
 *
 * bos session tests
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
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
		// Validate that at least we can decode the JSON into a list of strings
		stringList, err := common.DecodeJSONIntoStringList(resp.Body())
		if err != nil {
			common.Error(err)
			numTestsFailed++
		} else if len(stringList) == 0 {
			common.VerbosePrintDivider()
			common.Infof("skipping test GET /session/{session_id}")
			common.Infof("results from previous test is []")
		} else {
			// a session_id is available
			sessionId := stringList[0]

			url = baseurl + endpoints["bos"]["session"].Url + "/" + sessionId
			numTests++
			test.RestfulTestHeader("GET session_id", numTests, totalNumTests)
			resp, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
			if err != nil {
				common.Error(err)
				numTestsFailed++
			} else {
				// Validate that the session can at least be decoded into a string map
				_, err = common.DecodeJSONIntoStringMap(resp.Body())
				if err != nil {
					common.Error(err)
					numTestsFailed++
				}
			}
		}
		// TODO: deeper validation of returned response
	}

	test.RestfulTestResultSummary(numTestsFailed, numTests)
	return numTestsFailed == 0
}

func sessionTestsCLI(vnum int) bool {
	// test #1, list session
	common.Infof("Getting list of all BOS sessions via CLI")
	cmdOut := runCLICommand(vnum, "session", "list")
	if cmdOut == nil {
		return false
	}

	// Validate that at least we can decode the JSON into a list of strings
	stringList, err := common.DecodeJSONIntoStringList(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	}

	// test #2, list session with session_id
	// use results from previous tests, grab the first session
	if len(stringList) == 0 {
		common.VerbosePrintDivider()
		common.Infof("skipping test CLI describe sessiontemplate {session_template_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// A session id is available
	sessionId := stringList[0]

	common.Infof("Describing BOS session %s via CLI", sessionId)
	cmdOut = runCLICommand(vnum, "session", "describe", sessionId)
	if cmdOut == nil {
		return false
	}

	// Validate that the session can at least be decoded into a string map
	_, err = common.DecodeJSONIntoStringMap(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	}

	// TODO: deeper validation of returned response
	return true
}
