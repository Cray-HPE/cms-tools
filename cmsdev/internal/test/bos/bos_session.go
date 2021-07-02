package bos

/*
 * bos_session.go
 *
 * bos session tests
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 *
 * (MIT License)
 */

import (
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
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
