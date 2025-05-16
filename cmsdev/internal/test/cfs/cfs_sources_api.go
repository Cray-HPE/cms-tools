// MIT License
//
// (C) Copyright 2025 Hewlett Packard Enterprise Development LP
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

/*
 * cfs_sources_api.go
 *
 * cfs sources api functions
 *
 */
package cfs

import (
	"encoding/json"
	"fmt"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

var firstCloneURL = "https://api-gw-service-nmn.local/vcs/cray/csm-config-management.git"
var secondCloneURL = "https://api-gw-service-nmn.local/vcs/cray/csm-product-catalog.git"

func CreateCFSSourceRecordAPI(sourceName string) (cfsSourceRecord CFSSources, passed bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return CFSSources{}, false
	}

	// Create CFS source payload
	payload := map[string]interface{}{
		"name":      sourceName,
		"clone_url": firstCloneURL,
		"credentials": map[string]string{
			"username":              "testuser",
			"password":              "testpassword",
			"authentication_method": "password",
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	params.JsonStr = string(jsonPayload)
	common.Infof("CFS configuration payload: %s", string(jsonPayload))

	url := common.BASEURL + endpoints["cfs"]["sources"].Url + "/" + endpoints["cfs"]["sources"].Version + endpoints["cfs"]["sources"].Uri
	resp, err := test.RestfulVerifyStatus("POST", url, *params, http.StatusCreated)

	if err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsSourceRecord); err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	passed = true
	return
}

func UpdateCFSSourceRecordAPI(sourceName string) (cfsSourceRecord CFSSources, passed bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return CFSSources{}, false
	}

	// Create CFS source payload
	payload := map[string]string{
		"clone_url": secondCloneURL,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	params.JsonStrArray = jsonPayload
	common.Infof("CFS configuration payload: %s", string(jsonPayload))

	url := common.BASEURL + endpoints["cfs"]["sources"].Url + "/" + endpoints["cfs"]["sources"].Version + endpoints["cfs"]["sources"].Uri + "/" + sourceName
	resp, err := test.RestfulVerifyStatus("PATCH", url, *params, http.StatusOK)

	if err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsSourceRecord); err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	passed = true
	return
}

func DeleteCFSSourceRecordAPI(sourceName string) (passed bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return false
	}

	url := common.BASEURL + endpoints["cfs"]["sources"].Url + "/" + endpoints["cfs"]["sources"].Version + endpoints["cfs"]["sources"].Uri + "/" + sourceName
	_, err := test.RestfulVerifyStatus("DELETE", url, *params, http.StatusNoContent)

	if err != nil {
		common.Error(err)
		return false
	}

	return true
}

func GetCFSSourceRecordAPI(sourceName string, httpStatus int) (cfsSourceRecord CFSSources, passed bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return CFSSources{}, false
	}

	url := common.BASEURL + endpoints["cfs"]["sources"].Url + "/" + endpoints["cfs"]["sources"].Version + endpoints["cfs"]["sources"].Uri + "/" + sourceName
	resp, err := test.RestfulVerifyStatus("GET", url, *params, httpStatus)

	if err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsSourceRecord); err != nil {
		common.Error(err)
		return CFSSources{}, false
	}

	passed = true
	return
}

func GetCFSSourcesListAPI() (cfsSources []CFSSources, passed bool) {
	var cfsSourcesList CFSSourcesList
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return []CFSSources{}, false
	}

	url := common.BASEURL + endpoints["cfs"]["sources"].Url + "/" + endpoints["cfs"]["sources"].Version + endpoints["cfs"]["sources"].Uri
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)

	if err != nil {
		common.Error(err)
		return []CFSSources{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsSourcesList); err != nil {
		common.Error(err)
		return []CFSSources{}, false
	}

	cfsSources = cfsSourcesList.Sources
	passed = true
	return
}

func CFSSourceExists(cfsSourcesList []CFSSources, sourceName string) (passed bool) {
	passed = false
	for _, source := range cfsSourcesList {
		if source.Name == sourceName {
			common.Infof("CFS source %s found in the list", sourceName)
			return true
		}
	}
	common.Infof("CFS source %s not found in the list", sourceName)
	return false
}

func VerifyCFSSourceRecord(cfsSourceRecord CFSSources, sourceName, cloneUrl string) (passed bool) {
	passed = true
	if cfsSourceRecord.Name != sourceName {
		common.Errorf("CFS source name mismatch: expected %s, found %s", sourceName, cfsSourceRecord.Name)
		passed = false
	}
	if cfsSourceRecord.Clone_url != cloneUrl {
		common.Errorf("CFS source clone_url mismatch: expected %s, found %s", cloneUrl, cfsSourceRecord.Clone_url)
		passed = false
	}
	if cfsSourceRecord.Credentials.Authentication_method != "password" {
		common.Errorf("CFS source authentication method mismatch: expected password, found %s", cfsSourceRecord.Credentials.Authentication_method)
		passed = false
	}
	return passed
}
