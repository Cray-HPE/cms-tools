// MIT License
//
// (C) Copyright 2019-2024 Hewlett Packard Enterprise Development LP
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
package cfs

/*
 * cfs_api.go
 *
 * CFS API definitions
 *
 */

import (
	"fmt"
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

const cfsBaseUrl = common.BASEURL + "/apis/cfs"
const cfsHealthzUrl = cfsBaseUrl + "/healthz"
const cfsMaxVersion = 3
const cfsMinVersion = 2

var cfsVersionUrls = []string{cfsBaseUrl + "/", cfsBaseUrl + "/versions", cfsBaseUrl + "/v2", cfsBaseUrl + "/v3"}

type cfsEndpoint struct {
	Name     string // This must equal what you need to specify in the URI string
	IdField  string
	Versions []int
}

var cfsEndpoints = []cfsEndpoint{
	{
		Name:     "components",
		IdField:  "id",
		Versions: []int{2, 3},
	},
	{
		Name:     "configurations",
		IdField:  "name",
		Versions: []int{2, 3},
	},
	{
		Name:     "sessions",
		IdField:  "name",
		Versions: []int{2, 3},
	},
	{
		Name:     "sources",
		IdField:  "name",
		Versions: []int{3},
	},
}

func (endpoint cfsEndpoint) InVersion(version int) bool {
	for _, ver := range endpoint.Versions {
		if ver == version {
			return true
		}
	}
	return false
}

func (endpoint cfsEndpoint) Url(version int) string {
	return fmt.Sprintf("%s/v%d/%s", cfsBaseUrl, version, endpoint.Name)
}

func (endpoint cfsEndpoint) RunCliCommand(version int, cmdArgs ...string) []byte {
	cmdPrefix := []string{fmt.Sprintf("v%d", version), endpoint.Name}
	return test.RunCLICommandJSON("cfs", append(cmdPrefix, cmdArgs...)...)
}

func (endpoint cfsEndpoint) TestApi(params *common.Params) (passed bool) {
	passed = true
	var itemsList []interface{}
	common.Infof("API: Testing CFS %s endpoint", endpoint.Name)
	version := cfsMaxVersion + 1
	multiplePages := false
	for version > cfsMinVersion {
		version -= 1
		if endpoint.skipTest(version, multiplePages) {
			continue
		}
		common.Infof("API: Listing CFS %s using v%d endpoint", endpoint.Name, version)
		url := endpoint.Url(version)
		resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
		if err != nil {
			common.Error(err)
			passed = false
			continue
		}

		common.Debugf("API: Parsing response from CFS")
		// Paging was implemented in CFS v3
		if version < 3 {
			itemsList, err = endpoint.parseUnpagedListResponse(resp.Body())
			if err != nil {
				common.Error(err)
				passed = false
				continue
			}
		} else {
			itemsList, multiplePages, err = endpoint.parsePagedListResponse(resp.Body())
			if err != nil {
				common.Error(err)
				passed = false
				continue
			}
		}

		if len(itemsList) == 0 {
			common.Infof("API: GET %s returned an empty list -- skipping API test to get individual item", url)
			continue
		}

		// The list has entries, so let's get the ID field of the
		// first entry. Then we can do a GET/describe on that object
		idFieldValue, err := endpoint.getFirstId(itemsList)
		if err != nil {
			common.Error(err)
			passed = false
			continue
		}

		// Now try to get this item directly
		common.Infof("API: Getting CFS %s %s using v%d endpoint", endpoint.Name, idFieldValue, version)
		url += "/" + idFieldValue
		resp, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
		if err != nil {
			common.Error(err)
			passed = false
			continue
		}

		// Validate that we find the expected ID field value
		err = endpoint.checkIDField(resp.Body(), idFieldValue)
		if err != nil {
			common.Error(err)
			passed = false
		}
	}
	return
}

func (endpoint cfsEndpoint) TestCli() (passed bool) {
	passed = true
	var itemsList []interface{}
	var err error
	common.Infof("CLI: Testing CFS %s endpoint", endpoint.Name)
	version := cfsMaxVersion + 1
	multiplePages := false
	for version > cfsMinVersion {
		version -= 1
		if endpoint.skipTest(version, multiplePages) {
			continue
		}
		common.Infof("CLI: Listing CFS %s using v%d endpoint", endpoint.Name, version)
		cmdOut := endpoint.RunCliCommand(version, "list")
		if cmdOut == nil {
			passed = false
			continue
		}

		common.Debugf("API: Parsing response from CFS")
		// Paging was implemented in CFS v3
		if version < 3 {
			itemsList, err = endpoint.parseUnpagedListResponse(cmdOut)
			if err != nil {
				common.Error(err)
				passed = false
				continue
			}
		} else {
			itemsList, multiplePages, err = endpoint.parsePagedListResponse(cmdOut)
			if err != nil {
				common.Error(err)
				passed = false
				continue
			}
		}

		if len(itemsList) == 0 {
			common.Infof("CLI: v%d %s list returned an empty list -- skipping CLI test to get individual item", version, endpoint.Name)
			continue
		}

		// The list has entries, so let's get the ID field of the
		// first entry. Then we can do a GET/describe on that object
		idFieldValue, err := endpoint.getFirstId(itemsList)
		if err != nil {
			common.Error(err)
			passed = false
			continue
		}

		// Now try to get this item directly
		common.Infof("CLI: v%d %s describe %s", version, endpoint.Name, idFieldValue)
		cmdOut = endpoint.RunCliCommand(version, "describe", idFieldValue)
		if cmdOut == nil {
			passed = false
			continue
		}

		// Validate that we find the expected ID field value
		err = endpoint.checkIDField(cmdOut, idFieldValue)
		if err != nil {
			common.Error(err)
			passed = false
		}
	}
	return
}

func (endpoint cfsEndpoint) skipTest(version int, multiplePages bool) bool {
	// Return True if the test of this CFS endpoint/version combo should be skipped
	if !endpoint.InVersion(version) {
		common.Debugf("%s endpoint is not supported in CFS v%d; skipping", endpoint.Name, version)
		return true
	} else if (version == 2) && multiplePages {
		common.Infof("Too many %s in CFS to test v2 endpoint with current default page size; skipping", endpoint.Name)
		return true
	}
	return false
}

func (endpoint cfsEndpoint) parseUnpagedListResponse(responseBytes []byte) ([]interface{}, error) {
	// Paging was implemented in CFS v3, so if this is an earlier version, the response is a simple list
	return common.DecodeJSONIntoList(responseBytes)
}

func (endpoint cfsEndpoint) parsePagedListResponse(responseBytes []byte) (itemsList []interface{}, multiplePages bool, err error) {
	var mapObject map[string]interface{}
	// Our response should be a map object with two fields -- endpoint.Name and "next"
	// endpoint.Name should map to a list
	// next should either be nil or map to a string map
	mapObject, err = common.DecodeJSONIntoStringMap(responseBytes)
	if err != nil {
		return
	}

	nextRawValue, ok := mapObject["next"]
	if !ok {
		err = fmt.Errorf("Response is missing expected 'next' field")
		return
	}

	itemsListRawValue, ok := mapObject[endpoint.Name]
	if !ok {
		err = fmt.Errorf("Response is missing expected '%s' field", endpoint.Name)
		return
	}

	itemsList, ok = itemsListRawValue.([]interface{})
	if !ok {
		err = fmt.Errorf("Response field '%s' should map to a list but does not", endpoint.Name)
		return
	}

	if nextRawValue == nil {
		common.Debugf("Response field 'next' is nil, so there is only one page of items")
		multiplePages = false
	} else if _, ok = nextRawValue.(map[string]interface{}); ok {
		common.Debugf("Response field 'next' is not nil, so there is more than one page of items")
		multiplePages = true
	} else {
		err = fmt.Errorf("Response field 'next' should map to None or a dictionary, but does not")
	}

	return
}

func (endpoint cfsEndpoint) getFirstId(itemsList []interface{}) (idFieldValue string, err error) {
	// This function should only be called after it has been verified that the
	// list is not empty. The list has entries, so let's get the ID field of the
	// first entry. Then we can do a GET/describe on that object
	firstListEntry, ok := itemsList[0].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("First item in list is not a dictionary, but should be")
		return
	}

	idFieldRawValue, ok := firstListEntry[endpoint.IdField]
	if !ok {
		err = fmt.Errorf("First item in list is missing required '%s' field", endpoint.IdField)
		return
	}

	idFieldValue, ok = idFieldRawValue.(string)
	if !ok {
		err = fmt.Errorf("In first item listed, '%s' field does not map to a string, but it should", endpoint.IdField)
	}
	return
}

func (endpoint cfsEndpoint) checkIDField(mapCmdOut []byte, expectedIdValue string) error {
	// The endpoint names are plural ending in s -- this makes it singular
	objectName := "CFS " + endpoint.Name[:len(endpoint.Name)-1]
	return common.ValidateStringFieldValue(objectName, endpoint.IdField, expectedIdValue, mapCmdOut)
}

func validJSONStringMap(jsonBytes []byte) bool {
	_, err := common.DecodeJSONIntoStringMap(jsonBytes)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

func testCFSAPI() (passed bool) {
	passed = false
	common.Infof("Checking CFS API endpoints")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	passed = true

	common.Infof("API: Checking CFS service health")
	resp, err := test.RestfulVerifyStatus("GET", cfsHealthzUrl, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		passed = false
	} else if !validJSONStringMap(resp.Body()) {
		passed = false
	}

	common.Infof("API: Checking CFS version endpoints")
	for _, url := range cfsVersionUrls {
		resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
		if err != nil {
			common.Error(err)
			passed = false
		} else if !validJSONStringMap(resp.Body()) {
			passed = false
		}
	}

	common.Infof("API: Checking CFS option endpoints")
	version := cfsMinVersion
	for version <= cfsMaxVersion {
		url := fmt.Sprintf("%s/v%d/options", cfsBaseUrl, version)
		resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
		if err != nil {
			common.Error(err)
			passed = false
		} else if !validJSONStringMap(resp.Body()) {
			passed = false
		}
		version += 1
	}

	for _, endpoint := range cfsEndpoints {
		if !endpoint.TestApi(params) {
			passed = false
		}
	}
	return
}

func testCFSCLI() (passed bool) {
	passed = true

	// cray cfs healthz list
	common.Infof("CLI: Checking CFS service health")
	cmdOut := test.RunCLICommandJSON("cfs", "healthz", "list")
	if cmdOut == nil || !validJSONStringMap(cmdOut) {
		passed = false
	}

	common.Infof("CLI: Checking CFS version endpoints")
	// cray cfs list
	cmdOut = test.RunCLICommandJSON("cfs", "list")
	if cmdOut == nil || !validJSONStringMap(cmdOut) {
		passed = false
	}

	// cray cfs versions list
	cmdOut = test.RunCLICommandJSON("cfs", "versions", "list")
	if cmdOut == nil || !validJSONStringMap(cmdOut) {
		passed = false
	}

	// cray cfs v# list
	version := cfsMinVersion
	for version <= cfsMaxVersion {
		cmdOut = test.RunCLICommandJSON("cfs", fmt.Sprintf("v%d", version), "list")
		if cmdOut == nil || !validJSONStringMap(cmdOut) {
			passed = false
		}
		version += 1
	}

	// cray cfs v# options list
	common.Infof("API: Checking CFS option endpoints")
	version = cfsMinVersion
	for version <= cfsMaxVersion {
		cmdOut = test.RunCLICommandJSON("cfs", fmt.Sprintf("v%d", version), "options", "list")
		if cmdOut == nil || !validJSONStringMap(cmdOut) {
			passed = false
		}
		version += 1
	}

	for _, endpoint := range cfsEndpoints {
		if !endpoint.TestCli() {
			passed = false
		}
	}
	return
}
