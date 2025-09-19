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
 * bos_sessions_cli_tests.go
 *
 * BOS sessions CLI tests
 *
 */
package bos

import (
	"encoding/json"
	"fmt"
	"os"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func TestSessionsCRUDOperationsWithTenantUsingCLI() (passed bool) {
	passed = TestSessionsCRUDOperationsUsingCLI()
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
		passed = passed && TestSessionsCRUDOperationsUsingCLI()
		// Unsetting the tenant name after tests
		common.SetTenantName("")
	}
	return passed
}

func TestSessionsCRUDOperationsUsingCLI() (passed bool) {
	passed = true
	var testRan bool
	if len(common.GetTenantName()) != 0 {
		common.PrintLog(fmt.Sprintf("Running BOS session CLI tests with tenant: %s", common.GetTenantName()))
	} else {
		common.PrintLog("Running BOS session CLI tests without Tenant")
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
		for cliVersion := range bosCliVersions {
			common.PrintLog(fmt.Sprintf("Running BOS session CLI tests with version %s", cliVersion))
			// Ruuning test suite for both staged and non-staged sessions
			for _, staged := range []bool{true, false} {
				sessionRecord, success := TestCLIBOSSessionsCreate(staged, arch, imageId, bosCliVersions[cliVersion])
				if !success {
					common.Errorf("BOS Session creation failed for imageId %s and arch %s", imageId, archMap[arch])
					common.Warnf("Skipping rest of the BOS session tests for imageId %s and architecture %s", imageId, archMap[arch])
					passed = false
					continue
				}
				if len(sessionRecord.Name) != 0 {
					passed = TestCLIBOSSessionsDelete(sessionRecord.Name, bosCliVersions[cliVersion]) &&
						TestCLiBOSSessionsGetAll(bosCliVersions[cliVersion]) && passed
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

func TestCLIBOSSessionsCreate(staged bool, arch, imageId, cliVersion string) (sessionRecord BOSSession, passed bool) {
	// Create a session using the CLI
	sessionName := "BOS_Session_" + string(common.GetRandomString(10))
	common.PrintLog(fmt.Sprintf("Creating BOS session %s with staged=%t and arch %s", sessionName, staged, archMap[arch]))

	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))
	templateName := "BOS_SessionTemplate_" + string(common.GetRandomString(10))

	// Create Bos session template payload needed for session creation
	fileName, _, success := GetCreateBOSSessionTemplatePayloadCLI(cfgName, true, arch, imageId)
	if !success {
		return BOSSession{}, false
	}

	if common.IsDummyTenant(common.GetTenantName()) {
		// Set execution code to 2 to indiacate that we expect the command to fail
		test.SetCliExecreturnCode(2)
	}

	// Create BOS session template using the payload
	sessionTemplateRecord, success := CreateBOSSessionTemplatesCLI(templateName, fileName, cfgName, cliVersion)
	if !success {
		common.Errorf("Session template creation failed for imageId %s and arch %s", imageId, archMap[arch])
		return BOSSession{}, false
	}

	// remove the session template file after creation
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove session template file %s: %v", fileName, err)
	}

	sessionRecord, success = CreateBOSSessionCLI(staged, sessionName, sessionTemplateRecord.Name, cliVersion)
	if !success {
		common.Errorf("Session creation failed for imageId %s and arch %s", imageId, archMap[arch])
		return BOSSession{}, false
	}

	if test.GetCliExecreturnCode() != 0 {
		test.SetCliExecreturnCode(0)
		return BOSSession{}, true // If creation was expected to fail (e.g. using a dummy tenant), skip verification of created resource
	}

	// Creating expected session record for verification
	expectedBOSSession := BOSSession{
		Name:          sessionName,
		Template_name: sessionTemplateRecord.Name,
		Stage:         staged,
		Operation:     "reboot",
		Limit:         "fakexname",
	}

	// Marshalling the expected session record to JSON format
	expectedBOSSessionJson, err := json.Marshal(expectedBOSSession)
	if err != nil {
		common.Errorf("Error marshalling expected session record: %v", err)
		return BOSSession{}, false
	}

	// Verify the created session
	if !VerifyBOSSession(sessionRecord, string(expectedBOSSessionJson)) {
		common.Errorf("Verify failed for BOS session '%s'", sessionRecord.Name)
		return BOSSession{}, false
	}

	// Get the BOS session
	_, success = GetBOSSessionRecordCLI(sessionRecord.Name, cliVersion)
	if !success {
		return BOSSession{}, false
	}

	// Verify session in list of sessions
	sessionList, success := GetBOSSessionRecordsCLI(cliVersion)

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

func TestCLIBOSSessionsDelete(sessionName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting BOS session '%s'", sessionName))
	// Delete BOS session
	if !DeleteBOSSessionCLI(sessionName, cliVersion) {
		return false
	}

	// Set CLI execution return code to 2. Since the session is deleted, the command should return 2.
	test.SetCliExecreturnCode(2)
	// Check if BOS session is deleted
	_, success := GetBOSSessionRecordCLI(sessionName, cliVersion)
	if success {
		common.Errorf("BOS session '%s' still exists", sessionName)
		return false
	}

	// Set CLI execution return code to 0.
	test.SetCliExecreturnCode(0)

	// Check if BOS session is deleted from the list of all sessions
	sessionList, success := GetBOSSessionRecordsCLI(cliVersion)
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

func TestCLiBOSSessionsGetAll(cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Getting all BOS sessions"))
	// Get all BOS sessions using the CLI
	sessionList, success := GetBOSSessionRecordsCLI(cliVersion)
	if !success {
		return false
	}

	common.Infof("Found %d BOS sessions", len(sessionList))
	return true
}
