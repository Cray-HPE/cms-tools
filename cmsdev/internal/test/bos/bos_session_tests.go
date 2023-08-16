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
 * bos_session_tests.go
 *
 * BOS session tests
 *
 */

import (
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"strings"
)

// The sessionsV1TestsURI, sessionsV2TestsURI, sessionsV1TestsCLICommand, and sessionsV2TestsCLICommand functions
// define the API and CLI versions of the BOS v1 and v2 session subtests.
// They all do essentially the same thing, although the v2 API test has some enhancements described later in the file.
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

func sessionsTestsCLI(tenantList []string) (passed bool) {
	passed = true

	// v1 sessions
	if !sessionsV1TestsCLICommand("v1", bosV1SessionsCLI) {
		passed = false
	}

	// v2 sessions
	if !sessionsV2TestsCLICommand(tenantList, "v2", bosV2SessionsCLI) {
		passed = false
	}

	// default (v2) sessions
	if !sessionsV2TestsCLICommand(tenantList, bosDefaultSessionsCLI) {
		passed = false
	}

	return
}

////////////////////////////////////////////////////////////////////////////
// v1 tests
////////////////////////////////////////////////////////////////////////////

// v1 sessions API tests
// See comment at the top of the file for a description of this function
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

// v1 sessions CLI tests
// See comment at the top of the file for a description of this function
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

////////////////////////////////////////////////////////////////////////////
// v2 tests
////////////////////////////////////////////////////////////////////////////

// v2 sessions API tests
// See comment at the top of the file for a general description of this function
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

	// test #1, list sessions with no tenant specified
	sessionDataList, err = listV2SessionDataApi(params, "")
	if err != nil {
		common.Error(err)
		return false
	}
	if len(sessionDataList) == 0 {
		common.Infof("skipping test GET %s/{session_id} because result from previous test is []", bosV2SessionsUri)

		// However, we can still try to list all of the sessions with a tenant name specified.
		// Since no sessions were found from the un-tenanted query, we expect none to be found once a tenant
		// is specified
		tenant := getAnyTenant(tenantList)
		sessionDataList, err = listV2SessionDataApi(params, tenant)
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
		sessionDataList, err = listV2SessionDataApi(params, tenant)
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
	sessionDataList, err = listV2SessionDataApi(params, tenantedSessionData.Tenant)
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

// v2 sessions CLI tests
// See comment at the top of the file for a description of this function
func sessionsV2TestsCLICommand(cmdArgs ...string) bool {
	var tenantedSessionData, untenantedSessionData v2SessionData
	var tenantedSessionCount int
	var err error
	var ok, passed bool
	var sessionDataList []v2SessionData

	// test #1, list sessions with no tenant specified
	sessionDataList, ok = listV2SessionDataCli("", cmdArgs...)
	if !ok {
		return false
	}
	if len(sessionDataList) == 0 {
		common.Infof("skipping test CLI %s describe {session_id} because result from previous test is []", strings.Join(cmdArgs, " "))

		// However, we can still try to list all of the sessions with a tenant name specified.
		// Since no sessions were found from the un-tenanted query, we expect none to be found once a tenant
		// is specified
		tenant := getAnyTenant(tenantList)
		sessionDataList, ok = listV2SessionDataCli("", cmdArgs...)
		if !ok {
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

	passed = true

	if untenantedSessionData.IsNil() {
		common.Infof("skipping test CLI %s describe {session_id} with no tenant specified, because all BOS v2 sessions are owned by tenants", strings.Join(cmdArgs, " "))
	} else {
		// test: describe session using the untenanted session name we found earlier
		_, ok = describeV2SessionDataCli(untenantedSessionData, cmdArgs...)
		passed = passed && ok
	}

	if tenantedSessionData.IsNil() {
		common.Infof("No BOS v2 sessions found belonging to any tenants")

		tenant := getAnyTenant(tenantList)
		sessionDataList, ok = listV2SessionDataCli(tenant, cmdArgs...)
		if !ok {
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
	sessionDataList, ok = listV2SessionDataCli(tenantedSessionData.Tenant, cmdArgs...)
	if !ok {
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
	_, ok = describeV2SessionDataCli(tenantedSessionData, cmdArgs...)
	passed = passed && ok

	return passed
}
