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
	"net/http"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

const bosV1SessionsUri = bosV1BaseUri + "/session"
const bosV2SessionsUri = bosV2BaseUri + "/sessions"

const bosV1SessionsCLI = "session"
const bosV2SessionsCLI = "sessions"
const bosDefaultSessionsCLI = bosV1SessionsCLI

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

func sessionsTestsAPI(params *common.Params) (passed bool) {
	passed = true

	// v1 sessions
	if !sessionsV1TestsURI(bosV1SessionsUri, params) {
		passed = false
	}

	// v2 sessions
	if !sessionsV2TestsURI(bosV2SessionsUri, params) {
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

	// default (v1) sessions
	if !sessionsV1TestsCLICommand(bosDefaultSessionsCLI) {
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
	sessionId := stringList[0]
	common.Infof("Found BOS v1 session with ID '%s'", sessionId)

	// test #2, describe session with session_id
	uri += "/" + sessionId
	if !basicGetUriVerifyStringMapTest(uri, params) {
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
func ValidateV2Session(mapCmdOut []byte, expectedName string) bool {
	// For BOSv2, the session object we get back should have a 'name' field matching the name that we requested.
	// So let's validate that
	err := common.ValidateStringFieldValue("BOS session", "name", expectedName, mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// v2 sessions API tests
// See comment earlier in the file for a description of this function
func sessionsV2TestsURI(uri string, params *common.Params) bool {
	// test #1, list sessions
	common.Infof("GET %s test scenario", uri)
	resp, err := bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	// BOS v2: Decode JSON into a list
	sessionList, err := common.DecodeJSONIntoList(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionList) == 0 {
		common.Infof("skipping test GET %s/{session_id}", uri)
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
	sessionId, err := common.GetStringFieldFromMapObject("name", session)
	if err != nil {
		common.Error(err)
		return false
	}
	common.Infof("Found BOS v2 session with name '%s'", sessionId)

	// test #2: describe session
	uri += "/" + sessionId
	resp, err = bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	} else if !ValidateV2Session(resp.Body(), sessionId) {
		return false
	}

	return true
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
	sessionId := stringList[0]
	common.Infof("Found BOS v1 session with ID '%s'", sessionId)

	// test #2, describe session with session_id
	if !basicCLIDescribeVerifyStringMapTest(sessionId, cmdArgs...) {
		return false
	}

	return true
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
	sessionId, err := common.GetStringFieldFromMapObject("name", session)
	if err != nil {
		common.Error(err)
		return false
	}
	common.Infof("Found BOS v2 session with name '%s'", sessionId)

	// test #2, describe session with session_id
	cmdOut = runBosCLIDescribe(sessionId, cmdArgs...)
	if cmdOut == nil {
		return false
	} else if !ValidateV2Session(cmdOut, sessionId) {
		return false
	}

	return true
}
