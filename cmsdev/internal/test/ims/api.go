// MIT License
//
// (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP
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
package ims

/*
 * api.go
 *
 * ims api functions
 *
 */

import (
	"encoding/json"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

// Return specific job record in IMS via API
func getIMSJobRecordAPI(jobId string) (jobRecord IMSJobRecord, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting job record %s in IMS via API", jobId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["jobs"].Url + "/" + jobId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract job record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &jobRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return a list of all job records in IMS via API
func getIMSJobRecordsAPI() (recordList []IMSJobRecord, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting list of all job records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["jobs"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract list of job records from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Check IMS liveness probe. Returns True if live, False otherwise
func checkIMSLivenessProbe() bool {
	var baseurl string = common.BASEURL

	common.Infof("Checking IMS Liveness Probe")
	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}
	url := baseurl + endpoints["ims"]["live"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	// Validate that we can decode the response into a string map
	_, err = common.DecodeJSONIntoStringMap(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Check IMS readiness probe. Returns True if ready, False otherwise
func checkIMSReadinessProbe() bool {
	var baseurl string = common.BASEURL

	common.Infof("Checking IMS Readiness Probe")
	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}
	url := baseurl + endpoints["ims"]["ready"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	// Validate that we can decode the response into a string map
	_, err = common.DecodeJSONIntoStringMap(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Return IMS version
func getIMSVersion() (ver string, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting IMS version")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["version"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract version record from response
	var record IMSVersionRecord
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &record); err != nil {
		common.Error(err)
		return
	} else if len(record.Version) == 0 {
		common.Errorf("IMS version string is empty")
		return
	}
	ok = true
	ver = record.Version

	return
}
