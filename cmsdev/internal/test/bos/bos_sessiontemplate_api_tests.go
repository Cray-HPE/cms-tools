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
 * bos_sessiontemplate_api_tests.go
 *
 * BOS sessiontemplate API tests
 *
 */
package bos

import (
	"fmt"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func TestSessionTemplatesCRUDOperationsUsingTenants() (passed bool) {
	passed = TestSessionTemplatesCRUDOperations()
	tenantList := []string{}
	dummyTenantName := "dummy-tenant-" + string(common.GetRandomString(5))
	tenantList = append(tenantList, dummyTenantName)
	// Running the tests with tenants
	tenantName := GetTenantFromList()
	if len(tenantName) != 0 {
		tenantList = append(tenantList, tenantName)
	}

	for _, tenant := range tenantList {
		// Set the tenant name for the tests
		common.SetTenantName(tenant)
		passed = passed && TestSessionTemplatesCRUDOperations()
		// Unsetting the tenant name after tests
		common.SetTenantName("")
	}
	return passed
}

func TestSessionTemplatesCRUDOperations() (passed bool) {
	passed = true
	var testRan bool
	if len(common.GetTenantName()) != 0 {
		common.PrintLog(fmt.Sprintf("Running BOS session template tests with tenant: %s", common.GetTenantName()))
	} else {
		common.PrintLog("Running BOS session template tests without Tenant")
	}
	// Range over archMap to create session templates with different architectures
	for arch := range archMap {
		imageId, err := GetLatestImageIdFromCsmProductCatalog(arch)
		if err != nil {
			common.Infof("Unable to get latest image id for architecture %s", archMap[arch])
			common.Warnf("Skipping BOS session template tests for architecture %s", archMap[arch])
			continue
		}
		testRan = true
		common.PrintLog(fmt.Sprintf("Running BOS session template tests for arch %s", archMap[arch]))
		sessionTemplateRecord, ok := TestSessionTemplatesCreate(arch, imageId)
		if !ok {
			common.Errorf("Session template creation failed for imageId %s and arch %s", imageId, archMap[arch])
			common.Warnf("Skipping rest of the BOS session template tests for imageId %s and architecture %s", imageId, archMap[arch])
			passed = false
			continue
		}

		if len(sessionTemplateRecord.Name) != 0 {
			passed = TestSessionTemplatesUpdate(sessionTemplateRecord.Name, imageId) &&
				TestSessionTemplatesDelete(sessionTemplateRecord.Name) &&
				TestSessionTemplatesGetAll() && passed
		}
	}

	if !testRan {
		common.Infof("No image found for supported architecture")
		common.Warnf("Skipping BOS session template tests")
		// No image found for supported architecture, skip the tests
		// return true to indicate that the tests were skipped successfully
		return true
	}

	return passed
}

func TestSessionTemplatesCreate(imageArch string, imageId string) (sessionTemplateRecord BOSSessionTemplate, passed bool) {
	templateName := "BOS_SessionTemplate_" + string(common.GetRandomString(10))
	common.PrintLog(fmt.Sprintf("Creating BOS session template %s with image ID %s and  arch %s", templateName, imageId, imageArch))

	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	// create sessiontemplates payload
	payload, success := GetCreateBOSSessionTemplatePayload(cfgName, false, imageArch, imageId)
	if !success {
		return BOSSessionTemplate{}, false
	}
	common.Debugf("BOS Session template create payload: %s", payload)
	// Create session template
	sessionTemplateRecord, success = CreateUpdateBOSSessiontemplatesAPI(payload, templateName, "PUT")
	if !success {
		return BOSSessionTemplate{}, false
	}

	if GetExpectedHTTPStatusCode() != http.StatusOK {
		return BOSSessionTemplate{}, true // If the tenant is a dummy tenant, we skip the verification as creation is expected to fail
	}

	// Verify sessiontemplate
	if !VerifyBOSSessionTemplate(sessionTemplateRecord, payload, templateName) {
		common.Errorf("Session template %s verification failed", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// Get the created session template
	_, success = GetBOSSessionTemplatesAPI(sessionTemplateRecord.Name, http.StatusOK)
	if !success {
		common.Errorf("Unable to get BOS session template %s", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// verify session template in list of session templates
	sessionTemplateRecords, success := GetAllBOSSessionTemplatesAPI()
	if !success {
		common.Errorf("Unable to get all session templates")
		return BOSSessionTemplate{}, false
	}

	if !BOSSessionTemplateExists(sessionTemplateRecord.Name, sessionTemplateRecords) {
		common.Errorf("BOS session template %s not found in list of session templates", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// Validate session template
	if !ValidateBOSSessionTemplateAPI(sessionTemplateRecord.Name) {
		return BOSSessionTemplate{}, false
	}

	common.Infof("Session template %s created successfully", sessionTemplateRecord.Name)
	return sessionTemplateRecord, true

}

func TestSessionTemplatesUpdate(templateName string, imageId string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Updating session template %s", templateName))
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	sessionTemplate, success := GetBOSSessionTemplatesAPI(templateName, http.StatusOK)
	if !success {
		common.Errorf("Unable to get session template %s", templateName)
		return false
	}

	payload, success := GetCreateBOSSessionTemplatePayload(cfgName, true, sessionTemplate.Boot_sets.Compute.Arch, imageId)
	if !success {
		return false
	}
	common.Debugf("Session template update payload: %s", payload)
	sessionTemplateRecord, success := CreateUpdateBOSSessiontemplatesAPI(payload, templateName, "PATCH")
	if !success {
		common.Errorf("Session template %s update failed", templateName)
		return false
	}

	// Verify sessiontemplate
	if !VerifyBOSSessionTemplate(sessionTemplateRecord, payload, templateName) {
		common.Errorf("Session template %s verification failed", sessionTemplateRecord.Name)
		return false
	}

	// Get the created session template
	_, success = GetBOSSessionTemplatesAPI(sessionTemplateRecord.Name, http.StatusOK)
	if !success {
		common.Errorf("Unable to get BOS session template %s", sessionTemplateRecord.Name)
		return false
	}

	// Validate session template
	if !ValidateBOSSessionTemplateAPI(sessionTemplateRecord.Name) {
		return false
	}

	common.Infof("Session template %s updated successfully", templateName)
	return true

}

func TestSessionTemplatesDelete(templateName string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting session template %s", templateName))
	// Delete session template
	if !DeleteBOSSessionTemplatesAPI(templateName) {
		common.Errorf("Unable to delete sessiontemplate %s", templateName)
		return false
	}

	// Get the deleted session template
	_, success := GetBOSSessionTemplatesAPI(templateName, http.StatusNotFound)
	if !success {
		common.Errorf("BOS sessiontemplate %s was not deleted", templateName)
		return false
	}

	// verify session template in list of session templates
	sessionTemplateRecords, success := GetAllBOSSessionTemplatesAPI()
	if !success {
		common.Errorf("Unable to get all session templates")
		return false
	}

	if BOSSessionTemplateExists(templateName, sessionTemplateRecords) {
		common.Errorf("BOS session template %s not deleted, found in list of session templates", templateName)
		return false
	}

	common.Infof("Deleted sessiontemplate %s successfully", templateName)
	return true
}

func TestSessionTemplatesGetAll() (passed bool) {
	common.PrintLog("Getting all session templates")
	// Get all session templates
	sessionTemplateRecords, success := GetAllBOSSessionTemplatesAPI()
	if !success {
		common.Errorf("Unable to get all session templates")
		return false
	}

	common.Infof("Found %d BOS sessiontemplates", len(sessionTemplateRecords))
	return true
}
