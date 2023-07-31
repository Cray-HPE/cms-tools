// MIT License
//
// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
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
 * bos_api.go
 *
 * bos API definitions file
 *
 */

import (
	resty "gopkg.in/resty.v1"
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

// BOS base URLs and URIs
const bosBaseUrl = common.BASEURL + "/apis/bos"
const bosV1BaseUri = "/v1"
const bosV2BaseUri = "/v2"

// A wrapper for the common test.RestfulVerifyStatus function that converts a BOS URI into the full URL before
// calling that function
func bosRestfulVerifyStatus(method, uri string, params *common.Params, ExpectedStatus int) (*resty.Response, error) {
	return test.RestfulVerifyStatus(method, bosBaseUrl+uri, *params, ExpectedStatus)
}

func bosTenantRestfulVerifyStatus(method, uri, tenant string, params *common.Params, ExpectedStatus int) (*resty.Response, error) {
	return test.TenantRestfulVerifyStatus(method, bosBaseUrl+uri, tenant, *params, ExpectedStatus)
}

// Given a BOS URI, do a GET request to it. Verify that the response has 200 status code and returns a dictionary (aka string map) object.
// Return true if all of that worked fine. Otherwise, log an appropriate error and return false.
func basicGetUriVerifyStringMapTest(uri string, params *common.Params) bool {
	common.Infof("GET %s test scenario", uri)
	resp, err := bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	// Validate that object can be decoded into a string map at least
	_, err = common.DecodeJSONIntoStringMap(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

func basicTenantGetUriVerifyStringMapTest(uri, tenant string, params *common.Params) bool {
	common.Infof("GET %s (tenant: %s) test scenario", uri, tenant)
	resp, err := bosTenantRestfulVerifyStatus("GET", uri, tenant, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	// Validate that object can be decoded into a string map at least
	_, err = common.DecodeJSONIntoStringMap(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Run all of the BOS API subtests. Return true if they all pass, false otherwise.
func apiTests() (passed bool) {
	passed = true

	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}

	// Defined in bos_version.go
	if !versionTestsAPI(params) {
		passed = false
	}

	// Defined in bos_healthz.go
	if !healthzTestsAPI(params) {
		passed = false
	}

	// Defined in bos_components.go
	if !componentsTestsAPI(params) {
		passed = false
	}

	// Defined in bos_options.go
	if !optionsTestsAPI(params) {
		passed = false
	}

	// Defined in bos_sessiontemplate.go
	if !sessionTemplatesTestsAPI(params) {
		passed = false
	}

	// Defined in bos_session.go
	if !sessionsTestsAPI(params) {
		passed = false
	}

	return
}
