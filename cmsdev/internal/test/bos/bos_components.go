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
 * bos_components.go
 *
 * bos component tests
 *
 */

import (
	"fmt"
	resty "gopkg.in/resty.v1"
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

// components is new for BOS v2
const bosV2ComponentsUri = bosV2BaseUri + "/components"

const bosV2ComponentsCLI = "components"
const bosDefaultComponentsCLI = bosV2ComponentsCLI

// The componentsTestsURI and componentsTestsCLICommand functions define the API and CLI versions of the BOS components subtests.
// They both do the same thing:
// 1. List all components
// 2. Verify that this succeeds and returns something of the right general form
// 3. If the list returned is empty, then the subtest is over. Otherwise, select the first element of the list and extract the "id" field
// 4. Do a GET/describe on that particular component
// 5. Verify that this succeeds and returns something of the right general form

func componentsTestsAPI(params *common.Params, tenantList []string) (passed bool) {
	// Test with no tenant specified in the query.
	passed = componentsTestsURI(bosV2ComponentsUri, params, "")

	if len(tenantList) == 0 {
		common.Infof("Skipping tenanted components tests, because no tenants are defined on the system")
		return
	}

	if !componentsTestsURI(bosV2ComponentsUri, params, getAnyTenant(tenantList)) {
		passed = false
	}

	return
}

func componentsTestsCLI(tenantList []string) (passed bool) {
	passed = true

	// v2 (untenanted)
	if !componentsTestsCLICommand("", "v2", bosV2ComponentsCLI) {
		passed = false
	}

	// default (v2) (untenated)
	if !componentsTestsCLICommand("", bosDefaultComponentsCLI) {
		passed = false
	}

	if len(tenantList) == 0 {
		common.Infof("Skipping tenanted components tests, because no tenants are defined on the system")
		return
	}

	// v2
	if !componentsTestsCLICommand(getAnyTenant(tenantList), "v2", bosV2ComponentsCLI) {
		passed = false
	}

	// default (v2)
	if !componentsTestsCLICommand(getAnyTenant(tenantList), bosDefaultComponentsCLI) {
		passed = false
	}

	return
}

func getFirstComponentId(listCmdOut []byte) (string, error) {
	return common.GetStringFieldFromFirstItem("id", listCmdOut)
}

func ValidateComponentId(mapCmdOut []byte, expectedId string) bool {
	err := common.ValidateStringFieldValue("BOS component", "id", expectedId, mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// See comment earler in file for a description of this function
func componentsTestsURI(uri string, params *common.Params, tenantName string) (passed bool) {
	var resp *resty.Response
	var err error

	// test #1, list components
	if len(tenantName) == 0 {
		common.Infof("GET %s test scenario", uri)
		resp, err = bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	} else {
		common.Infof("GET %s (tenant: %s) test scenario", uri, tenantName)
		resp, err = bosTenantRestfulVerifyStatus("GET", uri, tenantName, params, http.StatusOK)
	}
	if err != nil {
		common.Error(err)
		return false
	}

	// use results from previous test, grab the first component ID
	componentId, err := getFirstComponentId(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(componentId) == 0 {
		common.Infof("skipping test GET %s/{component_id}", uri)
		common.Infof("results from previous test is []")
		return true
	}

	// a component id is available
	// test #2 describe component
	uri += "/" + componentId
	if len(tenantName) == 0 {
		common.Infof("GET %s test scenario", uri)
		resp, err = bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	} else {
		common.Infof("GET %s (tenant: %s) test scenario", uri, tenantName)
		resp, err = bosTenantRestfulVerifyStatus("GET", uri, tenantName, params, http.StatusOK)
	}
	if err != nil {
		common.Error(err)
		return false
	} else if !ValidateComponentId(resp.Body(), componentId) {
		return false
	}

	return true
}

// See comment earler in file for a description of this function
func componentsTestsCLICommand(tenantName string, cmdArgs ...string) bool {
	var tenantText string

	if len(tenantName) > 0 {
		tenantText = fmt.Sprintf(" on behalf of tenant '%s'", tenantName)
	}

	// test #1, list components
	cmdOut := runTenantBosCLIList(tenantName, cmdArgs...)
	if cmdOut == nil {
		return false
	}

	// use results from previous test, grab the first component
	componentId, err := getFirstComponentId(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	} else if len(componentId) == 0 {
		common.Infof("skipping test CLI describe component {component_id}%s because result from previous test is []", tenantText)
		return true
	}

	// a component_id is available
	// test #2 describe component
	cmdOut = runTenantBosCLIDescribe(tenantName, componentId, cmdArgs...)
	if cmdOut == nil {
		return false
	} else if !ValidateComponentId(cmdOut, componentId) {
		return false
	}

	return true
}
