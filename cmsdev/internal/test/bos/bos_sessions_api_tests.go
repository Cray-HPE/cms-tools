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
 * bos_sessions_api_tests.go
 *
 * BOS sessions API tests.
 *
 */
package bos

import (
	"fmt"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func TestBOSSessionsCRUDOperationsUsingTenants() (passed bool) {
	passed = TestBOSSessionsCRUDOperations()
	tenantList := []string{}
	dummyTenantName := common.GetDummyTenantName()
	tenantList = append(tenantList, dummyTenantName)
	// Running the tests with tenants
	tenantName := GetTenantFromList()
	if len(tenantName) != 0 {
		tenantList = append(tenantList, tenantName)
	}

	for _, tenant := range tenantList {
		// Set the tenant name for the tests
		common.SetTenantName(tenant)
		passed = passed && TestBOSSessionsCRUDOperations()
		// Unsetting the tenant name after tests
		common.SetTenantName("")
	}
	return passed
}

func TestBOSSessionsCRUDOperations() (passed bool) {
	passed = true
	var testRan bool
	if len(common.GetTenantName()) != 0 {
		common.PrintLog(fmt.Sprintf("Running BOS session tests with tenant: %s", common.GetTenantName()))
	} else {
		common.PrintLog("Running BOS session tests without tenant")
	}

	// Range over archMap to create session templates with different architectures
	for arch := range archMap {
		imageId, hasDummyData, err := GetLatestImageIdFromCsmProductCatalog(arch)
		if err != nil {
			common.Infof("Unable to get latest image id for architecture %s", archMap[arch])
			common.Warnf("Skipping BOS session tests for architecture %s", archMap[arch])
			continue
		}

		// If Dummy data flag is set, then mark the test as failed
		if hasDummyData {
			passed = false
		}

		testRan = true
		common.PrintLog(fmt.Sprintf("Running BOS session tests for arch %s", archMap[arch]))
		// Ruuning test suite for both staged and non-staged sessions
		for _, staged := range []bool{true, false} {
			sessionRecord, success := TestBOSSessionsCreate(staged, arch, imageId)
			if !success {
				common.Errorf("BOS Session creation failed for imageId %s and arch %s", imageId, archMap[arch])
				common.Warnf("Skipping rest of the BOS session tests for imageId %s and architecture %s", imageId, archMap[arch])
				passed = false
				continue
			}

			if len(sessionRecord.Name) != 0 {
				passed = TestBOSSessionsDelete(sessionRecord.Name) &&
					TestBOSSessionsGetAll() && passed
				// Deleting session template as part of cleanup
				sessionTemplateInventory := GetBOSSessionTemplateInventoryInstance()
				if !DeleteBOSSessionTemplatesAPI(sessionTemplateInventory.TemplateNameList[0]) {
					common.Errorf("Unable to delete BOS session template '%s'", sessionTemplateInventory.TemplateNameList[0])
				}
				// Remove the template from inventory
				if len(sessionTemplateInventory.TemplateNameList) > 0 {
					sessionTemplateInventory.TemplateNameList = sessionTemplateInventory.TemplateNameList[1:]
				}
			}
		}

	}

	if !testRan {
		common.Infof("No image found for supported architecture")
		common.Warnf("Skipping BOS session tests")
		// No image found for supported architecture, skip the tests
		// return true to indicate that the tests were skipped successfully
		return true
	}

	return passed
}

func TestBOSSessionsCreate(staged bool, arch string, imageId string) (sessionRecord BOSSession, passed bool) {
	sessionName := "BOS_Session_" + string(common.GetRandomString(10))
	templateName := "BOS_SessionTemplate_" + string(common.GetRandomString(10))
	common.PrintLog(fmt.Sprintf("Creating BOS session %s with staged=%t", sessionName, staged))

	// Create session payload
	sessionPayload, success := CreateBOSSessionPayload(sessionName, templateName, staged, "reboot", arch, imageId)
	if !success {
		return BOSSession{}, false
	}

	common.Debugf("Create BOS session payload: %s", sessionPayload)
	// Create BOS session
	sessionRecord, success = CreateBOSSessionAPI(sessionPayload)
	if !success {
		return BOSSession{}, false
	}

	if GetExpectedHTTPStatusCode() != http.StatusOK {
		return BOSSession{}, true // If creation was expected to fail (e.g. using a dummy tenant), skip verification of created resource
	}

	// Verify the created session
	if !VerifyBOSSession(sessionRecord, sessionPayload) {
		common.Errorf("Verify failed for BOS session '%s'", sessionRecord.Name)
		return BOSSession{}, false
	}

	// Get the BOS session
	_, success = GetBOSSessionAPI(sessionRecord.Name, http.StatusOK)
	if !success {
		common.Errorf("Failed to get BOS session '%s'", sessionRecord.Name)
		return BOSSession{}, false
	}

	// Check if BOS session is in the list of all sessions
	sessionList, success := GetAllBOSSessionsAPI()
	if !success {
		return BOSSession{}, false
	}

	if !BOSSessionExists(sessionRecord.Name, sessionList) {
		common.Errorf("BOS session '%s' not found in the list of all sessions", sessionRecord.Name)
		return BOSSession{}, false
	}
	// Add the session template to inventory for cleanup
	sessionTemplateInventory := GetBOSSessionTemplateInventoryInstance()
	sessionTemplateInventory.TemplateNameList = append(sessionTemplateInventory.TemplateNameList, templateName)

	common.Infof("BOS session %s created successfully", sessionRecord.Name)
	return sessionRecord, true
}

func TestBOSSessionsDelete(sessionName string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting BOS session '%s'", sessionName))
	// Delete BOS session
	if !DeleteBOSSessionAPI(sessionName) {
		return false
	}

	// Check if BOS session is deleted
	_, success := GetBOSSessionAPI(sessionName, http.StatusNotFound)
	if !success {
		common.Errorf("BOS session '%s' still exists", sessionName)
		return false
	}

	// Check if BOS session is deleted from the list of all sessions
	sessionList, success := GetAllBOSSessionsAPI()
	if !success {
		return false
	}

	if BOSSessionExists(sessionName, sessionList) {
		common.Errorf("BOS session '%s' not deleted, found in the list of all sessions", sessionName)
		return false
	}

	common.Infof("Deleted BOS session '%s' successfully", sessionName)
	return true
}

func TestBOSSessionsGetAll() (passed bool) {
	// Get all BOS sessions
	common.PrintLog("Getting all BOS sessions")
	sessionList, success := GetAllBOSSessionsAPI()
	if !success {
		return false
	}
	common.Infof("Found %d BOS sessions", len(sessionList))
	return true
}
