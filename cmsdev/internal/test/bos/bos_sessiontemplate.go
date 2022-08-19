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

func sessionTemplatesTestsAPI(params *common.Params) (passed bool) {
	passed = true

	// v1 session template template
	if !sessionTemplateTemplateTestURI(bosV1SessionTemplateTemplateUri, params) {
		passed = false
	}

	// v2 session template template
	if !sessionTemplateTemplateTestURI(bosV2SessionTemplateTemplateUri, params) {
		passed = false
	}

	// v1 session templates
	if !sessionTemplatesTestsURI(bosV1SessionTemplatesUri, params) {
		passed = false
	}

	// v2 session templates
	if !sessionTemplatesTestsURI(bosV2SessionTemplatesUri, params) {
		passed = false
	}

	return
}

func sessionTemplatesTestsCLI() (passed bool) {
	passed = true

	// v1 session template template
	if !sessionTemplateTemplateTestCLICommand("v1", bosV1SessionTemplateTemplateCLI) {
		passed = false
	}

	// v2 session template template
	if !sessionTemplateTemplateTestCLICommand("v2", bosV2SessionTemplateTemplateCLI) {
		passed = false
	}

	// default (v2) session template template
	if !sessionTemplateTemplateTestCLICommand(bosDefaultSessionTemplateTemplateCLI) {
		passed = false
	}

	// v1 session templates
	if !sessionTemplatesTestsCLICommand("v1", bosV1SessionTemplatesCLI) {
		passed = false
	}

	// v2 session templates
	if !sessionTemplatesTestsCLICommand("v2", bosV2SessionTemplatesCLI) {
		passed = false
	}

	// default (v2) session templates
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

// sessiontemplatetemplate API test
func sessionTemplateTemplateTestURI(uri string, params *common.Params) bool {
	return basicGetUriVerifyStringMapTest(uri, params)
}

// sessiontemplatetemplate CLI test
func sessionTemplateTemplateTestCLICommand(cmdArgs ...string) bool {
	return basicCLIListVerifyStringMapTest(cmdArgs...)
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
		common.Infof("skipping test CLI describe component {session_template_id}")
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
