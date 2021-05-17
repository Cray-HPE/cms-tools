package ims

/*
 * api.go
 *
 * ims api functions
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
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
	"encoding/json"
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// Return specific image record in IMS via API
func getIMSImageRecordAPI(imageId string) (imageRecord IMSImageRecord, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting image record %s in IMS via API", imageId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["images"].Url + "/" + imageId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract image record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &imageRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return a list of all image records in IMS via API
func getIMSImageRecordsAPI() (recordList []IMSImageRecord, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting list of all image records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["images"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract list of image records from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

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

// Return specific public key record in IMS via API
func getIMSPublicKeyRecordAPI(pkeyId string) (pkeyRecord IMSPublicKeyRecord, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting public key record %s in IMS via API", pkeyId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["public_keys"].Url + "/" + pkeyId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract public key record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &pkeyRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return a list of all public key records in IMS via API
func getIMSPublicKeyRecordsAPI() (recordList []IMSPublicKeyRecord, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting list of all public key records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["public_keys"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract list of public key records from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return a specific recipe record in IMS via API
func getIMSRecipeRecordAPI(recipeId string) (recipeRecord IMSRecipeRecord, ok bool) {
	ok = false
	var baseurl string = common.BASEURL

	common.Infof("Describing recipe record %s in IMS via API", recipeId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}

	url := baseurl + endpoints["ims"]["recipes"].Url + "/" + recipeId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		return
	}

	// Extract recipe record from command output
	common.Infof("Decoding JSON in command output")
	err = json.Unmarshal(resp.Body(), &recipeRecord)
	if err != nil {
		return
	}
	ok = true
	return
}

// Return a list of all recipe records in IMS via API
func getIMSRecipeRecordsAPI() (recordList []IMSRecipeRecord, ok bool) {
	var baseurl string = common.BASEURL
	ok = false

	common.Infof("Getting list of all recipe records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := baseurl + endpoints["ims"]["recipes"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract list of recipe records from response
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
