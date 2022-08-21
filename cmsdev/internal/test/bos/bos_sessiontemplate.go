//
//  MIT License
//
//  (C) Copyright 2019-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//
package bos

/*
 * bos_sessiontemplate.go
 *
 * bos sessiontemplate tests
 *
 */

import (
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

const bosV1SessionTemplatesUri = bosV1BaseUri + "/sessiontemplate"
const bosV2SessionTemplatesUri = bosV2BaseUri + "/sessiontemplates"

const bosV1SessionTemplateTemplateUri = bosV1BaseUri + "/sessiontemplatetemplate"
const bosV2SessionTemplateTemplateUri = bosV2BaseUri + "/sessiontemplatetemplate"

const bosV1SessionTemplatesCLI = "sessiontemplate"
const bosV2SessionTemplatesCLI = "sessiontemplates"
const bosDefaultSessionTemplatesCLI = bosV2SessionTemplatesCLI

const bosV1SessionTemplateTemplateCLI = "sessiontemplatetemplate"
const bosV2SessionTemplateTemplateCLI = "sessiontemplatetemplate"
const bosDefaultSessionTemplateTemplateCLI = bosV2SessionTemplateTemplateCLI

// The sessionTemplatesTestsURI and sessionTemplatesTestsCLICommand functions define the API and CLI versions of the BOS session template subtests.
// They both do the same thing:
// 1. List all session templates
// 2. Verify that this succeeds and returns something of the right general form
// 3. If the list returned is empty, then the subtest is over. Otherwise, select the first element of the list and extract the "name" field
// 4. Do a GET/describe on that particular session template
// 5. Verify that this succeeds and returns something of the right general form

func sessionTemplatesTestsAPI(params *common.Params) (passed bool) {
	passed = true

	// session template template API tests
	// Just do a GET of the sessiontemplatetemplate endpoint and make sure that the response has
	// 200 status and a dictionary object

	// v1
	if !basicGetUriVerifyStringMapTest(bosV1SessionTemplateTemplateUri, params) {
		passed = false
	}

	// v2
	if !basicGetUriVerifyStringMapTest(bosV2SessionTemplateTemplateUri, params) {
		passed = false
	}

	// session template API tests

	// v1
	if !sessionTemplatesTestsURI(bosV1SessionTemplatesUri, params) {
		passed = false
	}

	// v2
	if !sessionTemplatesTestsURI(bosV2SessionTemplatesUri, params) {
		passed = false
	}

	return
}

func sessionTemplatesTestsCLI() (passed bool) {
	passed = true

	// session template template CLI tests
	// Make sure that "sessiontemplatetemplate list" CLI commmand succeeds and returns a dictionary object.

	// v1 sessiontemplatetemplate list
	if !basicCLIListVerifyStringMapTest("v1", bosV1SessionTemplateTemplateCLI) {
		passed = false
	}

	// v2 sessiontemplatetemplate list
	if !basicCLIListVerifyStringMapTest("v2", bosV2SessionTemplateTemplateCLI) {
		passed = false
	}

	// sessiontemplatetemplate list
	if !basicCLIListVerifyStringMapTest(bosDefaultSessionTemplateTemplateCLI) {
		passed = false
	}

	// session template CLI tests

	// v1 sessiontemplate
	if !sessionTemplatesTestsCLICommand("v1", bosV1SessionTemplatesCLI) {
		passed = false
	}

	// v2 sessiontemplates
	if !sessionTemplatesTestsCLICommand("v2", bosV2SessionTemplatesCLI) {
		passed = false
	}

	// sessiontemplates
	if !sessionTemplatesTestsCLICommand(bosDefaultSessionTemplatesCLI) {
		passed = false
	}

	return
}

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

// session templates API tests
func sessionTemplatesTestsURI(uri string, params *common.Params) bool {
	// test #1, list session templates
	common.Infof("GET %s test scenario", uri)
	resp, err := bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	// use results from previous test, grab the first session template
	sessionTemplateId, err := getFirstSessionTemplateId(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionTemplateId) == 0 {
		common.Infof("skipping test GET %s/{session_template_id}", uri)
		common.Infof("results from previous test is []")
		return true
	}

	// a session_template_id is available
	uri += "/" + sessionTemplateId
	common.Infof("GET %s test scenario", uri)
	resp, err = bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	} else if !ValidateSessionTemplateId(resp.Body(), sessionTemplateId) {
		return false
	}

	return true
}

// session templates CLI tests
func sessionTemplatesTestsCLICommand(cmdArgs ...string) bool {
	// test #1, list session templates
	cmdOut := runBosCLIList(cmdArgs...)
	if cmdOut == nil {
		return false
	}

	// use results from previous test, grab the first session template
	sessionTemplateId, err := getFirstSessionTemplateId(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionTemplateId) == 0 {
		common.Infof("skipping test CLI describe session template {session_template_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// a session template id is available
	// test #2 describe session templates
	cmdOut = runBosCLIDescribe(sessionTemplateId, cmdArgs...)
	if cmdOut == nil {
		return false
	} else if !ValidateSessionTemplateId(cmdOut, sessionTemplateId) {
		return false
	}

	return true
}
