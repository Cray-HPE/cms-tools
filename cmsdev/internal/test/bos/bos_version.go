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
 * bos_version.go
 *
 * bos version tests
 *
 */

import (
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

// This endpoint returns a list of BOS versions
const bosBaseUri = "/"

const bosV1VersionUri = bosV1BaseUri + "/version"
const bosV2VersionUri = bosV2BaseUri + "/version"

const bosV1VersionCLI = "version"
const bosV2VersionCLI = "version"
const bosDefaultVersionCLI = bosV2VersionCLI

func versionTestsAPI(params *common.Params) (passed bool) {
	passed = true
	if !versionListTestAPI(params) {
		passed = false
	}

	if !versionTestURI(bosV1BaseUri, params) {
		passed = false
	}

	if !versionTestURI(bosV1VersionUri, params) {
		passed = false
	}

	if !versionTestURI(bosV2BaseUri, params) {
		passed = false
	}

	if !versionTestURI(bosV2VersionUri, params) {
		passed = false
	}

	return
}

func versionTestsCLI() (passed bool) {
	passed = true

	// v1
	if !versionTestCLICommand("v1", bosV1VersionCLI) {
		passed = false
	}

	// v2
	if !versionTestCLICommand("v2", bosV2VersionCLI) {
		passed = false
	}

	// default (v2)
	if !versionTestCLICommand(bosDefaultVersionCLI) {
		passed = false
	}

	return
}

// There is no corresponding CLI test for the / endpoint
func versionListTestAPI(params *common.Params) bool {
	common.VerbosePrintDivider()
	common.Infof("GET %s test scenario", bosBaseUri)
	resp, err := bosRestfulVerifyStatus("GET", bosBaseUri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	// Validate that object can be decoded into a list of string maps at least
	_, err = common.DecodeJSONIntoStringMapList(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

func versionTestURI(uri string, params *common.Params) bool {
	return basicGetUriVerifyStringMapTest(uri, params)
}

func versionTestCLICommand(cmdArgs ...string) bool {
	return basicCLIListVerifyStringMapTest(cmdArgs...)
}
