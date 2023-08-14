// MIT License
//
// (C) Copyright 2019-2023 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
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

func versionTestsAPI(params *common.Params, tenantList []string) (passed bool) {
	passed = true

	// / endpoint
	// Verify that a GET to this endpoint returns status 200 and a list of dictionary objects
	if !versionListTestAPI(params) {
		passed = false
	}

	// For the remaining endpoints:
	// Do a GET of the version endpoint and make sure that the response has
	// 200 status and a dictionary object

	// /v1 endpoint
	if !basicGetUriVerifyStringMapTest(bosV1BaseUri, params) {
		passed = false
	}

	// /v1/version endpoint
	if !basicGetUriVerifyStringMapTest(bosV1VersionUri, params) {
		passed = false
	}

	// /v2 endpoint
	if !basicGetUriVerifyStringMapTest(bosV2BaseUri, params) {
		passed = false
	}

	// v2 endpoint as random tenant (BOS does not verify that tenant exists on GET requests)
	if !basicTenantGetUriVerifyStringMapTest(bosV2BaseUri, getAnyTenant(tenantList), params) {
		passed = false
	}

	// /v2/version endpoint
	if !basicGetUriVerifyStringMapTest(bosV2VersionUri, params) {
		passed = false
	}

	// /v2/version endpoint as random tenant (BOS does not verify that tenant exists on GET requests)
	if !basicTenantGetUriVerifyStringMapTest(bosV2VersionUri, getAnyTenant(tenantList), params) {
		passed = false
	}

	return
}

func versionTestsCLI() (passed bool) {
	passed = true

	// Make sure that "version list" CLI command succeeds and returns a dictionary object.

	// /v1 endpoint - "cray bos v1 list"
	if !basicCLIListVerifyStringMapTest("v1") {
		passed = false
	}

	// v1 version list - "cray bos v1 version list"
	if !basicCLIListVerifyStringMapTest("v1", bosV1VersionCLI) {
		passed = false
	}

	// /v2 endpoint - "cray bos v2 list"
	if !basicCLIListVerifyStringMapTest("v2") {
		passed = false
	}

	// v2 version list - "cray bos v2 version list"
	if !basicCLIListVerifyStringMapTest("v2", bosV2VersionCLI) {
		passed = false
	}

	// version list - "cray bos version list"
	if !basicCLIListVerifyStringMapTest(bosDefaultVersionCLI) {
		passed = false
	}

	// default endpoint - "cray bos list"
	if !basicCLIListVerifyStringMapTest() {
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
