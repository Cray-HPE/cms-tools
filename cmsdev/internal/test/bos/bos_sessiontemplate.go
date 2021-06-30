package bos

/*
 * bos_sessiontemplate.go
 *
 * bos sessiontemplate tests
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

func getFirstSessionTemplateId(listCmdOut []byte) (string, error) {
	return common.GetStringFieldFromFirstItem("name", listCmdOut)
}

func ValidateSessionTemplateId(mapCmdOut []byte, expectedName string) bool {
	err := common.ValidateStringFieldValue("BOS sessiontemplate", "name", expectedName, mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// sessiontemplate tests
func sessionTemplateTestsAPI() bool {
	var baseurl string = common.BASEURL
	const totalNumTests int = 2

	numTests, numTestsFailed := 0, 0
	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}

	// test #1, get session template template
	url := baseurl + endpoints["bos"]["sessiontemplatetemplate"].Url
	numTests++
	test.RestfulTestHeader("GET sessiontemplatetemplate", numTests, totalNumTests)
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	} else {
		// Validate that we can decode it into a map object
		_, err = common.DecodeJSONIntoStringMap(resp.Body())
		if err != nil {
			common.Error(err)
			numTestsFailed++
		}
		// TODO: deeper validation of returned response
	}

	// test #2, list session templates
	url = baseurl + endpoints["bos"]["sessiontemplate"].Url
	numTests++
	test.RestfulTestHeader("GET sessiontemplate", numTests, totalNumTests)
	resp, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	// TODO: deeper validation of returned response

	// test #3, list sessiontemplate with session_template_id
	// use results from previous tests, grab the first sessiontemplate
	sessionTemplateId, err := getFirstSessionTemplateId(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionTemplateId) == 0 {
		common.VerbosePrintDivider()
		common.Infof("skipping test GET /sessiontemplate/{session_template_id}")
		common.Infof("results from previous test is []")
		return numTestsFailed == 0
	}

	// a session_template_id is available
	url = baseurl + endpoints["bos"]["sessiontemplate"].Url + "/" + sessionTemplateId
	numTests++
	test.RestfulTestHeader("GET session_template_id", numTests, totalNumTests)
	resp, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	} else if !ValidateSessionTemplateId(resp.Body(), sessionTemplateId) {
		numTestsFailed++
	}

	// TODO: deeper validation of returned response

	return numTestsFailed == 0
}

func sessionTemplateTestsCLI(vnum int) bool {
	var err error
	var numTestsFailed = 0

	// test #1, get example session template
	common.Infof("Getting example BOS session template via CLI")
	cmdOut := runCLICommand(vnum, "sessiontemplatetemplate", "list")
	if cmdOut == nil {
		numTestsFailed++
	} else {
		// Validate that we can decode it into a map object
		_, err = common.DecodeJSONIntoStringMap(cmdOut)
		if err != nil {
			common.Error(err)
			numTestsFailed++
		}
		// TODO: deeper validation of returned response
	}

	// test #2, list session templates
	common.Infof("Getting list of all BOS session templates via CLI")
	cmdOut = runCLICommand(vnum, "sessiontemplate", "list")
	if cmdOut == nil {
		return false
	}
	// TODO: deeper validation of returned response

	// test #3, list sessiontemplate with session_template_id
	// use results from previous tests, grab the first sessiontemplate
	sessionTemplateId, err := getFirstSessionTemplateId(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionTemplateId) == 0 {
		common.VerbosePrintDivider()
		common.Infof("skipping test CLI describe sessiontemplate {session_template_id}")
		common.Infof("results from previous test is []")
		return numTestsFailed == 0
	}

	// a session_template_id is available
	common.Infof("Describing BOS session template %s via CLI", sessionTemplateId)
	cmdOut = runCLICommand(vnum, "sessiontemplate", "describe", sessionTemplateId)
	if cmdOut == nil {
		return false
	} else if !ValidateSessionTemplateId(cmdOut, sessionTemplateId) {
		return false
	}

	// TODO: deeper validation of returned response
	return numTestsFailed == 0
}
