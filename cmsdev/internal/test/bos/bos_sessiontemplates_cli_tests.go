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
 * bos_sessiontemplate_cli_tests.go
 *
 * BOS sessiontemplate CLI tests
 *
 */
package bos

import (
	"fmt"
	"os"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func TestSessionTemplatesCRUDOperationsWithTenantUsingCLI() (passed bool) {
	passed = TestSessionTemplatesCRUDOperationsUsingCLI()
	tenantList := []string{}
	dummyTenantName := GetDummyTenantName()
	tenantList = append(tenantList, dummyTenantName)
	// Running the tests with tenants
	tenantName := GetTenantFromList()
	if len(tenantName) != 0 {
		tenantList = append(tenantList, tenantName)
	}

	for _, tenant := range tenantList {
		// Set the tenant name for the tests
		common.SetTenantName(tenant)
		passed = passed && TestSessionTemplatesCRUDOperationsUsingCLI()
		// Unsetting the tenant name after tests
		common.SetTenantName("")
	}
	return passed
}

func TestSessionTemplatesCRUDOperationsUsingCLI() (passed bool) {
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
			common.Warnf("Skipping BOS session template CLI tests for architecture %s", archMap[arch])
			continue
		}
		testRan = true
		for cliVersion := range bosCliVersions {
			common.PrintLog(fmt.Sprintf("Running BOS session template CLI tests with version %s", cliVersion))
			sessionTemplateRecord, ok := TestCLISessionTemplatesCreate(arch, imageId, bosCliVersions[cliVersion])
			if !ok {
				common.Errorf("Session template creation failed for imageId %s and arch %s", imageId, archMap[arch])
				common.Warnf("Skipping rest of the BOS session template CLI tests for imageId %s and architecture %s", imageId, archMap[arch])
				passed = false
				continue
			}

			if len(sessionTemplateRecord.Name) != 0 {
				common.Infof("Session template created successfully with name %s", sessionTemplateRecord.Name)
				passed = TestCLISessionTemplatesUpdate(sessionTemplateRecord.Name, imageId, bosCliVersions[cliVersion]) &&
					TestCLISessionTemplatesDelete(sessionTemplateRecord.Name, bosCliVersions[cliVersion]) &&
					TestCLISessionTemplatesGetAll(sessionTemplateRecord.Name, bosCliVersions[cliVersion])
			}
		}

	}
	if !testRan {
		common.Infof("No image found for supported architecture")
		common.Warnf("Skipping BOS session template CLI tests")
		// No image found for supported architecture, skip the tests
		// return true to indicate that the tests were skipped successfully
		return true
	}
	return passed
}

func TestCLISessionTemplatesCreate(arch, imageId, cliVersion string) (sessionTemplateRecord BOSSessionTemplate, passed bool) {
	// Create a session template using the CLI
	templateName := "BOS_SessionTemplate_" + string(common.GetRandomString(10))
	common.PrintLog(fmt.Sprintf("Creating BOS session template %s with image ID %s and  arch %s", templateName, imageId, arch))

	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))
	// create sessiontemplates payload
	fileName, payload, success := GetCreateBOSSessionTemplatePayloadCLI(cfgName, false, arch, imageId)
	if !success {
		return BOSSessionTemplate{}, false
	}

	// If the tenant is a dummy tenant, we expect the command to fail
	if IsDummyTenant(common.GetTenantName()) {
		// Set execution code to 2 to indicate that the session template creation is not supported for dummy tenants
		test.SetCliExecreturnCode(2)
	}

	sessionTemplateRecord, success = CreateBOSSessionTemplatesCLI(templateName, fileName, cfgName, cliVersion)
	if !success {
		return BOSSessionTemplate{}, false
	}

	// Remove the created session template file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
	}

	// If creation was expected to fail (e.g. using a dummy tenant), skip verification of created resource
	if test.GetCliExecreturnCode() != 0 {
		test.SetCliExecreturnCode(0)
		return sessionTemplateRecord, true
	}

	// Verify sessiontemplate
	if !VerifyBOSSessionTemplate(sessionTemplateRecord, payload, templateName) {
		common.Errorf("BOS session template %s verification failed", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// Get the created session template
	_, success = GetBOSSessiontemplatesCLI(sessionTemplateRecord.Name, cliVersion)
	if !success {
		common.Errorf("Unable to get BOS session template %s", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	//verify session template in list of session templates
	sessionTemplateRecords, success := GetBOSSessiontemplatesListCLI(cliVersion)
	if !success {
		return BOSSessionTemplate{}, false
	}

	if !BOSSessionTemplateExists(sessionTemplateRecord.Name, sessionTemplateRecords) {
		common.Errorf("BOS session template %s not found in list of session templates", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// Validate session template
	if !ValidateBOSSessionTemplateCLI(sessionTemplateRecord.Name, cliVersion) {
		return BOSSessionTemplate{}, false
	}

	return sessionTemplateRecord, true
}

func TestCLISessionTemplatesUpdate(templateName, imageId, cliVersion string) (passed bool) {
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))
	// Update the session template using the CLI
	common.PrintLog(fmt.Sprintf("Updating BOS session template %s with CFS config %s", templateName, cfgName))

	// Get the existing session template
	sessionTemplate, success := GetBOSSessiontemplatesCLI(templateName, cliVersion)
	if !success {
		common.Errorf("Unable to get BOS session template %s", templateName)
		return false
	}

	// create sessiontemplates payload
	fileName, payload, success := GetCreateBOSSessionTemplatePayloadCLI(cfgName, true, sessionTemplate.Boot_sets.Compute.Arch, imageId)
	if !success {
		return false
	}

	sessionTemplateRecord, success := UpdateBOSSessionTemplatesCLI(templateName, fileName, cfgName, cliVersion)
	if !success {
		return false
	}

	// Remove the created session template file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
	}

	// Verify sessiontemplate
	if !VerifyBOSSessionTemplate(sessionTemplateRecord, payload, templateName) {
		common.Errorf("BOS session template %s verification failed", sessionTemplateRecord.Name)
		return false
	}

	// Get the updated session template
	_, success = GetBOSSessiontemplatesCLI(sessionTemplateRecord.Name, cliVersion)
	if !success {
		common.Errorf("Unable to get BOS session template %s", sessionTemplateRecord.Name)
		return false
	}

	//verify session template in list of session templates
	sessionTemplateRecords, success := GetBOSSessiontemplatesListCLI(cliVersion)
	if !success {
		return false
	}

	if !BOSSessionTemplateExists(sessionTemplateRecord.Name, sessionTemplateRecords) {
		common.Errorf("BOS session template %s not found in list of session templates", sessionTemplateRecord.Name)
		return false
	}

	// Validate session template
	if !ValidateBOSSessionTemplateCLI(sessionTemplateRecord.Name, cliVersion) {
		return false
	}

	common.Infof("BOS Session template %s updated successfully", sessionTemplateRecord.Name)

	return true
}

func TestCLISessionTemplatesDelete(templateName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting BOS session template %s", templateName))

	// Delete the session template using the CLI
	if !DeleteBOSSessionTemplatesCLI(templateName, cliVersion) {
		common.Errorf("Unable to delete BOS session template %s", templateName)
		return false
	}

	// Set CLI execution return code to 2. Since the session template is deleted, the command should return 2.
	test.SetCliExecreturnCode(2)

	// Get the deleted session template
	_, success := GetBOSSessiontemplatesCLI(templateName, cliVersion)
	if success {
		common.Errorf("BOS session template %s was not deleted", templateName)
		return false
	}

	// Set CLI execution return code to 0.
	test.SetCliExecreturnCode(0)

	// Verify that the session template is deleted
	sessionTemplateRecords, success := GetBOSSessiontemplatesListCLI(cliVersion)
	if !success {
		return false
	}

	if BOSSessionTemplateExists(templateName, sessionTemplateRecords) {
		common.Errorf("BOS session template %s not deleted, found in list of session templates", templateName)
		return false
	}

	common.Infof("BOS Session template %s deleted successfully", templateName)

	return true
}

func TestCLISessionTemplatesGetAll(templateName, cliVersion string) (passed bool) {
	common.PrintLog("Getting all BOS session templates")

	// Get all session templates using the CLI
	sessionTemplateRecords, success := GetBOSSessiontemplatesListCLI(cliVersion)
	if !success {
		common.Errorf("Unable to get all session templates")
		return false
	}

	common.Infof("Found %d BOS session templates", len(sessionTemplateRecords))

	return true
}
