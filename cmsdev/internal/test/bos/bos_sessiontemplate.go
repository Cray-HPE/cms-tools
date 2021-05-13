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
	"fmt"
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

func getFirstSessionTemplateId(listCmdOut []byte) (string, error) {
	return common.GetStringFieldFromFirstItem("name", listCmdOut)
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
	}
	// TODO: deeper validation of returned response

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
	_, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	}
	// TODO: deeper validation of returned response

	return numTestsFailed == 0
}

func sessionTemplateTestsCLI(vnum int) bool {
	var cmdString, verString string
	var numTestsFailed = 0

	if vnum == 0 {
		verString = "bos"
	} else if vnum > 0 {
		verString = fmt.Sprintf("bos v%d", vnum)
	} else {
		common.Errorf("PROGRAMMING LOGIC ERROR: sessionTestCLI: Negative vnum value (%d)", vnum)
	}

	// test #1, get example session template
	common.Infof("Getting example BOS session template via CLI")
	cmdString = fmt.Sprintf("%s sessiontemplatetemplate list --format json", verString)
	cmdOut := test.RunCLICommand(cmdString)
	if cmdOut == nil {
		numTestsFailed++
	}
	// TODO: deeper validation of returned response

	// test #2, list session templates
	common.Infof("Getting list of all BOS session templates via CLI")
	cmdString = fmt.Sprintf("%s sessiontemplate list --format json", verString)
	cmdOut = test.RunCLICommand(cmdString)
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
	cmdString = fmt.Sprintf("%s sessiontemplate describe %s --format json", verString, sessionTemplateId)
	cmdOut = test.RunCLICommand(cmdString)
	if cmdOut == nil {
		return false
	}

	// TODO: deeper validation of returned response
	return numTestsFailed == 0
}
