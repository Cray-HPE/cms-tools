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
package bos

/*
 * bos_session.go
 *
 * BOS session tests utility functions
 *
 */

import (
	"errors"
	"fmt"
	resty "gopkg.in/resty.v1"
	"net/http"
	"reflect"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

const bosV2SessionsUri = bosV2BaseUri + "/sessions"

const bosV2SessionsCLI = "sessions"
const bosDefaultSessionsCLI = bosV2SessionsCLI

// A BOS v2 session is uniquely identified by its name and tenant. The tenant may null, blank,
// or, equivalently, not present in the actual record. In all such cases, the
// following struct uses an empty string value for Tenant.
type v2SessionData struct {
	Name, Tenant string
}

// Returns true if the Name field of the object is "", false otherwise
func (sessionData v2SessionData) IsNil() bool {
	return sessionData.Name == ""
}

// Returns a string representation of the object
func (sessionData v2SessionData) String() string {
	if len(sessionData.Tenant) > 0 {
		return fmt.Sprintf("name: '%s', tenant: '%s'", sessionData.Name, sessionData.Tenant)
	}
	return fmt.Sprintf("name: '%s', no tenant", sessionData.Name)
}

// Compares two sessionData objects. Returns true if both fields match, false otherwise. If false,
// appropriate errors are also logged noting the discrepancies
func (sessionData v2SessionData) HasExpectedValues(expectedSessionData v2SessionData) (passed bool) {
	// Let's be optimistic and assume that they will match
	passed = true

	// Validate that name fields match
	if sessionData.Name != expectedSessionData.Name {
		common.Errorf("BOS session name '%s' does not match expected name '%s'", sessionData.Name, expectedSessionData.Name)
		passed = false
	} else {
		common.Debugf("BOS session name '%s' matches expected value", sessionData.Name)
	}

	// Validate the tenant name
	if sessionData.Tenant == expectedSessionData.Tenant {
		if len(sessionData.Tenant) == 0 {
			common.Debugf("Session does not belong to a tenant, which matches expectations")
		} else {
			common.Debugf("Session belongs to the expected tenant")
		}
		return
	}
	passed = false
	if len(sessionData.Tenant) == 0 {
		common.Errorf("Session does not belong to a tenant, but it should belong to '%s'", expectedSessionData.Tenant)
	} else {
		common.Errorf("Session belongs to tenant '%s', but it should not belong to any tenant", sessionData.Tenant)
	}
	return
}

// Takes as input a mapping from strings to arbitrary objects.
// Parses it to extract the values of the 'name' and 'tenant' fields (although the
// latter is allowed to be absent). Validates that 'name' (and 'tenant', if present and non-null)
// have string values and that 'name' is non-0 length. Returns a v2SessionData object
// populated from those fields. Returns an error with any problems encountered.
func parseV2SessionData(sessionDict map[string]interface{}) (sessionData v2SessionData, totalErr error) {
	var err error
	var fieldFound, fieldIsString bool
	var sessionTenantField interface{}

	common.Debugf("Getting name of session")
	sessionData.Name, err = common.GetStringFieldFromMapObject("name", sessionDict)
	if err != nil {
		err = fmt.Errorf("%w; Error getting 'name' field of BOS session", err)
	} else if len(sessionData.Name) == 0 {
		err = fmt.Errorf("BOS session has a 0-length name")
	} else {
		common.Debugf("Name of session is '%s'", sessionData.Name)
	}
	totalErr = errors.Join(totalErr, err)

	common.Debugf("Checking for 'tenant' field of session '%s'", sessionData.Name)
	sessionTenantField, fieldFound = sessionDict["tenant"]
	if !fieldFound {
		// This session has no tenant field, which is equivalent to a 0-length string tenant field
		sessionData.Tenant = ""
		common.Debugf("Session '%s' has no 'tenant' field", sessionData.Name)
		return
	}
	if sessionTenantField == nil {
		// This session has null-value tenant field, which is equivalent to a 0-length string tenant field
		sessionData.Tenant = ""
		common.Debugf("Session '%s' has null 'tenant' field", sessionData.Name)
		return
	}

	// If it is present and non-null, it should be a string value
	sessionData.Tenant, fieldIsString = sessionTenantField.(string)
	if !fieldIsString {
		// tenant field has non-string value
		err = fmt.Errorf("Session '%s' has a non-null 'tenant' field but its value is type %s, not string",
			sessionData.Name, reflect.TypeOf(sessionTenantField).String())
		totalErr = errors.Join(totalErr, err)
	}
	common.Debugf("Session: %s", sessionData.String())
	return
}

// Performs an API query to get a particular BOS v2 session (possibly belonging to a tenant)
// Parses the result to convert it to a dictionary with string keys
// Returns the result and error (if any)
func getV2SessionApi(params *common.Params, sessionData v2SessionData) (sessionDict map[string]interface{}, err error) {
	var resp *resty.Response
	uri := bosV2SessionsUri + "/" + sessionData.Name
	if len(sessionData.Tenant) == 0 {
		common.Infof("GET %s test scenario", uri)
		resp, err = bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	} else {
		common.Infof("GET %s (tenant: %s) test scenario", uri, sessionData.Tenant)
		resp, err = bosTenantRestfulVerifyStatus("GET", uri, sessionData.Tenant, params, http.StatusOK)
	}
	if err != nil {
		return
	}
	// Decode JSON into a string map
	sessionDict, err = common.DecodeJSONIntoStringMap(resp.Body())
	return
}

// Performs an API query to list BOS v2 sessions (possibly with a tenant specified)
// Parses the result to convert it to a list of dictionaries with string keys
// Returns the result and error (if any)
func listV2SessionsApi(params *common.Params, tenantName string) (dictList []map[string]interface{}, err error) {
	var resp *resty.Response
	if len(tenantName) == 0 {
		common.Infof("GET %s test scenario", bosV2SessionsUri)
		resp, err = bosRestfulVerifyStatus("GET", bosV2SessionsUri, params, http.StatusOK)
	} else {
		common.Infof("GET %s (tenant: %s) test scenario", bosV2SessionsUri, tenantName)
		resp, err = bosTenantRestfulVerifyStatus("GET", bosV2SessionsUri, tenantName, params, http.StatusOK)
	}
	if err != nil {
		return
	}
	// Decode JSON into a list of string maps
	dictList, err = common.DecodeJSONIntoStringMapList(resp.Body())
	return
}

// Get a particular BOS v2 session (possibly belonging to a tenant) using getV2SessionApi,
// parses that dictionary into a v2SessionData struct.
// Validates that it matches the expected session name and (if any) tenant name.
// Returns that struct and error (if any)
func getV2SessionDataApi(params *common.Params, sessionData v2SessionData) (sessionDataFromApi v2SessionData, err error) {
	var sessionDict map[string]interface{}

	sessionDict, err = getV2SessionApi(params, sessionData)
	if err != nil {
		return
	}
	sessionDataFromApi, err = parseV2SessionData(sessionDict)
	if err != nil {
		return
	}
	if !sessionDataFromApi.HasExpectedValues(sessionData) {
		err = fmt.Errorf("Session returned by API query (%s) does not match session requested (%s)",
			sessionDataFromApi.String(), sessionData.String())
	}
	return
}

// Describe a particular BOS v2 session (possibly belonging to a tenant) using bosTenantDescribeCli,
// parse that dictionary into a v2SessionData struct.
// Validates that it matches the expected session name and (if any) tenant name.
// Returns that struct and a boolean indicating pass/fail
func describeV2SessionDataCli(sessionData v2SessionData, cmdArgs ...string) (sessionDataFromCli v2SessionData, passed bool) {
	var sessionDict map[string]interface{}
	var err error
	var ok bool

	passed = false
	sessionDict, ok = bosTenantDescribeCli(sessionData.Tenant, sessionData.Name, cmdArgs...)
	if !ok {
		return
	}
	sessionDataFromCli, err = parseV2SessionData(sessionDict)
	if err != nil {
		return
	}
	if !sessionDataFromCli.HasExpectedValues(sessionData) {
		common.Errorf("Session returned by CLI command (%s) does not match session requested (%s)",
			sessionDataFromCli.String(), sessionData.String())
	} else {
		passed = true
	}
	return
}

// Parses list of string-key dictionaries into a list of v2SessionData structs.
// If a tenant was specified, validate that every session belongs to that tenant.
// Returns that list of structs and error (if any)
func dictListToSessionDataList(dictList []map[string]interface{}, tenantName string) (sessionDataList []v2SessionData, err error) {
	var sessionData v2SessionData
	sessionDataList = make([]v2SessionData, 0, len(dictList))
	for sessionIndex, sessionDict := range dictList {
		sessionData, err = parseV2SessionData(sessionDict)
		if err != nil {
			err = fmt.Errorf("%w; Error parsing session #%d in list", err, sessionIndex)
			return
		}
		// If a tenant was specified, validate that session belongs to the expected tenant
		if len(tenantName) > 0 && sessionData.Tenant != tenantName {
			err = fmt.Errorf("Session #%d in the list (%s) does not belong to expected tenant '%s'",
				sessionIndex, sessionData.String(), tenantName)
			return
		}
		sessionDataList = append(sessionDataList, sessionData)
	}
	return
}

// Gets a list of V2 session dictionary objects using listV2SessionsApi,
// converts that list using dictListToSessionDataList. Returns resulting list and error (if any)
func listV2SessionDataApi(params *common.Params, tenantName string) (sessionDataList []v2SessionData, err error) {
	var dictList []map[string]interface{}

	dictList, err = listV2SessionsApi(params, tenantName)
	if err != nil {
		return
	}
	sessionDataList, err = dictListToSessionDataList(dictList, tenantName)
	return
}

// Gets a list of V2 session dictionary objects using bosListCli,
// converts that list using dictListToSessionDataList. Returns resulting list and boolean
// value indicating whether the function passed or failed (an error will have been logged in the case
// of failure)
func listV2SessionDataCli(tenantName string, cmdArgs ...string) (sessionDataList []v2SessionData, passed bool) {
	var dictList []map[string]interface{}
	var err error

	dictList, passed = bosTenantListCli(tenantName, cmdArgs...)
	if !passed {
		return
	}
	sessionDataList, err = dictListToSessionDataList(dictList, tenantName)
	if err != nil {
		common.Error(err)
		passed = false
	}
	return
}
