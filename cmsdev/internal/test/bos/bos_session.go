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
 * bos_session.go
 *
 * bos session tests
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

const bosV1SessionsUri = bosV1BaseUri + "/session"
const bosV2SessionsUri = bosV2BaseUri + "/sessions"

const bosV1SessionsCLI = "session"
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

// Compares two objects. Returns true if both fields match, false otherwise. If false,
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

// The sessionsV1TestsURI, sessionsV2TestsURI, sessionsV1TestsCLICommand, and sessionsV2TestsCLICommand functions define the API and CLI versions of the
// BOS v1 and v2 session subtests.
// They all do essentially the same thing:
// 1. List all sessions
// 2. Verify that this succeeds and returns something of the right general form (for v1, this is expected to be a list of strings,
//    whereas for v2 this should be a list of dictionary objects)
// 3. If the list returned is empty, then the subtest is over. Otherwise, select the first element of the list. (If bosV2, extract the "name" field of that element).
// 4. Do a GET/describe on that particular session
// 5. Verify that this succeeds and returns something of the right general form. For BOS v2, also verify that it has the expected name
//    (in v1, the session ID is not in the returned object)

func sessionsTestsAPI(params *common.Params, tenantList []string) (passed bool) {
	passed = true

	// v1 sessions
	if !sessionsV1TestsURI(bosV1SessionsUri, params) {
		passed = false
	}

	// v2 sessions
	if !sessionsV2TestsURI(params, tenantList) {
		passed = false
	}

	return
}

func sessionsTestsCLI() (passed bool) {
	passed = true

	// v1 sessions
	if !sessionsV1TestsCLICommand("v1", bosV1SessionsCLI) {
		passed = false
	}

	// v2 sessions
	if !sessionsV2TestsCLICommand("v2", bosV2SessionsCLI) {
		passed = false
	}

	// default (v2) sessions
	if !sessionsV2TestsCLICommand(bosDefaultSessionsCLI) {
		passed = false
	}

	return
}

// v1 sessions API tests
// See comment earlier in the file for a description of this function
func sessionsV1TestsURI(uri string, params *common.Params) bool {
	// test #1, list session
	common.Infof("GET %s test scenario", uri)
	resp, err := bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	// BOS v1: Validate that at least we can decode the JSON into a list of strings
	stringList, err := common.DecodeJSONIntoStringList(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(stringList) == 0 {
		common.Infof("skipping test GET %s/{session_id}", uri)
		common.Infof("results from previous test is []")
		return true
	}

	// a session_id is available
	sessionData := stringList[0]
	common.Infof("Found BOS v1 session with ID '%s'", sessionData)

	// test #2, describe session with session_id
	uri += "/" + sessionData
	if !basicGetUriVerifyStringMapTest(uri, params) {
		return false
	}

	return true
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
// parses that dictionaries into a v2SessionData struct.
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

// Gets a list of V2 session dictionary objects using listV2SessionsApi,
// parses those dictionaries into v2SessionData structs.
// If a tenant was specified, validate that every session belongs to that tenant.
// Returns that list of structs and error (if any)
func listV2SessionDatasApi(params *common.Params, tenantName string) (sessionDataList []v2SessionData, err error) {
	var dictList []map[string]interface{}
	var sessionData v2SessionData

	dictList, err = listV2SessionsApi(params, tenantName)
	if err != nil {
		return
	}
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

// v2 sessions API tests
// See comment earlier in the file for a general description of this function
//
// The v2 API version of this function has had further improvements made in order to test good-path multitenancy
// queries. Specifically, this function now does the following (all using API calls)
// 1. API query to list all BOS v2 sessions (no tenant name specified)
// 2. Parses the result, verifying that it is a JSON list, and verifying that each item in that list:
//    a. Is a JSON dictionary
//    b. Has a "name" field which maps to a non-0-length string
//    c. Either has no "tenant" field or has a "tenant" field that maps to a (possibly 0-length) string
// 3. If the list is empty, then pick a random tenant name (possibly not one which exists on the system)
//    and issue a query for all BOS sessions belonging to that tenant. Verify that the resulting list is empty.
// 4. If the list from #1 is not empty, but no sessions belong to a tenant, then do #3.
// 5. If the list from #1 is not empty and has sessions owned by a tenant, then pick one such tenant, pick above
//    session owned by that tenant, and also count how many total sessions are owned by that tenant.
//    a. Query for all BOS sessions belonging to that tenant and verify that we get the same total number. Also
//       perform the same validation steps as in #2, and validate that every returned session belongs to the specified tenant.
//    b. Query for the specific BOS session belonging to that tenant and verify that it can be retrieved.
// 6. If the list from #1 is not empty and has at least one session that is not owned by a tenant, then pick such
//    a session, query BOS for that specific session, and verify that it can be retrieved.

// Several of these checks could fail if BOS sessions are being created or deleted while this test is running.
// While an administrator may be unlikely to choose to create or delete BOS sessions while running this test,
// the BOS operator responsible for cleaning up old sessions could delete a session during test execution and cause
// it to fail. For now, this will be a documented limitation of the test, with the recommendation to re-run just
// the BOS health check in the case that certain failures are seen. Eventually the test could be improved to automatically
// retry the relevant checks a limited number of times, to reduce the likelihood of false failures.

func sessionsV2TestsURI(params *common.Params, tenantList []string) bool {
	var tenantedSessionData, untenantedSessionData v2SessionData
	var tenantedSessionCount int
	var err error
	var sessionDataList []v2SessionData

	// test #1, list sessions
	sessionDataList, err = listV2SessionDatasApi(params, "")
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionDataList) == 0 {
		common.Infof("skipping test GET %s/{session_id}", bosV2SessionsUri)
		common.Infof("results from previous test is []")

		// However, we can still try to list all of the sessions with a tenant name specified.
		// Since no sessions were found from the un-tenanted query, we expect none to be found once a tenant
		// is specified
		tenant := getAnyTenant(tenantList)
		sessionDataList, err = listV2SessionDatasApi(params, tenant)
		if err != nil {
			common.Error(err)
			return false
		}

		// Validate that this list is empty
		if len(sessionDataList) > 0 {
			common.Errorf("Listing of all sessions empty, but found %d when listing sessions for tenant '%s'",
				len(sessionDataList), tenant)
			return false
		}
		common.Infof("Session list is still empty, as expected")
		return true
	}

	// From the session ID list, we want to identify:
	// * One session that has no tenant (untenantedSessionData)
	// * One session that belongs to a tenant (tenantedSessionData), and a count (tenantedSessionCount)
	//   of how many sessions in total have that same tenant
	for sessionIndex, sessionData := range sessionDataList {
		common.Debugf("Parsing session #%d in the session list (%s)", sessionIndex, sessionData.String())
		if len(sessionData.Tenant) == 0 {
			// This session has no tenant
			if untenantedSessionData.IsNil() {
				// This is the first session we have encountered that has no tenant field,
				// so take note of it
				untenantedSessionData = sessionData
				common.Infof("Found BOS v2 session #%d (%s)", sessionIndex, untenantedSessionData.String())
			}
			continue
		}
		// This session is owned by a tenant.
		if tenantedSessionData.IsNil() {
			// This is the first session we've found that is owned by a tenant, so remember it,
			// and note that we have found 1 session belonging to this tenant so far
			tenantedSessionData = sessionData
			tenantedSessionCount = 1
			common.Infof("Found BOS v2 session #%d (%s)", sessionIndex, tenantedSessionData.String())
			continue
		}
		// We have already found a session belonging to a tenant. If it is the tenant we found first,
		// increment our session count
		if sessionData.Tenant == tenantedSessionData.Tenant {
			tenantedSessionCount += 1
		}
	}

	passed := true

	if untenantedSessionData.IsNil() {
		common.Infof("skipping test GET %s/{session_id} with no tenant specified, because all BOS v2 sessions are owned by tenants",
			bosV2SessionsUri)
	} else {
		// test: describe session using the untenanted session name we found earlier
		_, err = getV2SessionDataApi(params, untenantedSessionData)
		if err != nil {
			common.Error(err)
			passed = false
		}
	}

	if tenantedSessionData.IsNil() {
		common.Infof("No BOS v2 sessions found belonging to any tenants")

		tenant := getAnyTenant(tenantList)
		sessionDataList, err = listV2SessionDatasApi(params, tenant)
		if err != nil {
			common.Error(err)
			passed = false
		} else {
			// Validate that this list is empty
			if len(sessionDataList) > 0 {
				common.Errorf("Listing all sessions found none owned by tenants, but found %d when listing sessions for tenant '%s'",
					len(sessionDataList), tenant)
				passed = false
			}
			common.Infof("List of sessions belonging to tenant '%s' is empty, as expected", tenant)
		}
		return passed
	}
	common.Infof("Counted %d BOS v2 sessions belonging to tenant '%s'", tenantedSessionCount, tenantedSessionData.Tenant)
	sessionDataList, err = listV2SessionDatasApi(params, tenantedSessionData.Tenant)
	if err != nil {
		common.Error(err)
		passed = false
	} else {
		// Validate that this list is expected length
		if len(sessionDataList) != tenantedSessionCount {
			common.Errorf("Listing all sessions found %d owned by tenant '%s', but found %d when listing sessions for that tenant",
				tenantedSessionCount, tenantedSessionData.Tenant, len(sessionDataList))
			passed = false
		}
		common.Infof("List of sessions belonging to tenant '%s' is the expected length", tenantedSessionData.Tenant)
	}

	// test: describe session using the session name we found owned by a tenant in the earlier loop
	_, err = getV2SessionDataApi(params, tenantedSessionData)
	if err != nil {
		common.Error(err)
		return false
	}

	return passed
}

// v1 sessions CLI tests
// See comment earlier in the file for a description of this function
func sessionsV1TestsCLICommand(cmdArgs ...string) bool {
	// test #1, list sessions
	cmdOut := runBosCLIList(cmdArgs...)
	if cmdOut == nil {
		return false
	}

	// BOS v1: Validate that at least we can decode the JSON into a list of strings
	stringList, err := common.DecodeJSONIntoStringList(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	}

	// Grab the first session ID
	if len(stringList) == 0 {
		common.Infof("skipping test CLI describe {session_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// A session id is available
	sessionData := stringList[0]
	common.Infof("Found BOS v1 session with ID '%s'", sessionData)

	// test #2, describe session with session_id
	if !basicCLIDescribeVerifyStringMapTest(sessionData, cmdArgs...) {
		return false
	}

	return true
}

// Given a response object (as an array of bytes), validate that:
// 1. It resolves to a JSON dictonary
// 2. That dictionary has a "name" field
// 3. The "name" field of that dictionary has a value which matches our expectedName string
//
// Return true if all of the above is true. Otherwise, log an appropriate error and return false.
func ValidateV2Session(mapCmdOut []byte, expectedId v2SessionData) bool {
	common.Infof("Validating that BOS v2 session has expected values for name and tenant")

	// This function should always receive a non-0-length expected name for the session
	if len(expectedId.Name) == 0 {
		common.Errorf("Programming logic error: ValidateV2Session function received 0-length string for expected name")
		return false
	}

	// Should be a dictionary object mapping strings to values
	common.Debugf("Parsing session as a dictionary")
	sessionDict, err := common.DecodeJSONIntoStringMap(mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}

	sessionData, err := parseV2SessionData(sessionDict)
	if err != nil {
		common.Error(err)
		return false
	}

	return sessionData.HasExpectedValues(expectedId)
}

// v2 sessions CLI tests
// See comment earlier in the file for a description of this function
func sessionsV2TestsCLICommand(cmdArgs ...string) bool {
	// test #1, list sessions
	cmdOut := runBosCLIList(cmdArgs...)
	if cmdOut == nil {
		return false
	}

	// BOS v2: Decode JSON into a list
	sessionList, err := common.DecodeJSONIntoList(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionList) == 0 {
		common.Infof("skipping test CLI describe {session_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// Take the first session listed. It should be a dictionary object.
	sessionObject := sessionList[0]
	session, ok := sessionObject.(map[string]interface{})
	if !ok {
		common.Errorf("First BOS session listed is not a dictionary object: %v", sessionObject)
		return false
	}

	common.Debugf("Getting 'name' field BOS session: %v", session)
	sessionData, err := common.GetStringFieldFromMapObject("name", session)
	if err != nil {
		common.Error(err)
		return false
	}
	common.Infof("Found BOS v2 session with name '%s'", sessionData)

	// test #2, describe session with session_id
	cmdOut = runBosCLIDescribe(sessionData, cmdArgs...)
	if cmdOut == nil {
		return false
	} else if !ValidateV2Session(cmdOut, v2SessionData{Name: sessionData, Tenant: ""}) {
		return false
	}

	return true
}
