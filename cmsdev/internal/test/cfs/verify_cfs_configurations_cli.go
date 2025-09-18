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
 * verify_cfs_configurations_cli.go
 *
 * cfs configurations test cli functions
 *
 */
package cfs

import (
	"fmt"
	"os"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func TestCFSConfigurationsCRUDOperationWithTenantsUsingCLI() (passed bool) {
	passed = TestCFSConfigurationsCRUDOperationUsingCLI()
	tenantList := []string{}
	dummyTenant := common.GetDummyTenantName()
	tenantList = append(tenantList, dummyTenant)
	// Get an actual tenant
	tenantName := GetTenantFromList()
	if len(tenantName) > 0 {
		tenantList = append(tenantList, tenantName)
	}
	for _, tenant := range tenantList {
		common.SetTenantName(tenant)
		result := TestCFSConfigurationsCRUDOperationUsingCLI()
		passed = passed && result
		common.SetTenantName("")
	}
	return passed
}

func TestCFSConfigurationsCRUDOperationUsingCLI() (passed bool) {
	passed = true

	if len(common.GetTenantName()) != 0 {
		common.PrintLog(fmt.Sprintf("Running CFS configurations CLI tests with tenant: %s", common.GetTenantName()))
	} else {
		common.PrintLog("Running CFS configurations CLI tests without tenant")
	}

	// Get supported API versions for configurations endpoints
	for _, cliVersion := range GetSupportAPIVersions("configurations") {
		if common.GetTenantName() == "" || cliVersion != "v2" {
			common.PrintLog(fmt.Sprintf("Testing CFS configurations CRUD operations using CLI and version: %s", cliVersion))
			result := TestCFSConfigurationsCRUDOperationCLI(cliVersion)
			passed = passed && result
		}
	}

	if common.GetTenantName() == "" {
		// Test default CLI version which is v2
		common.PrintLog("Testing CFS configurations CRUD operations using default CLI version")
		result := TestCFSConfigurationsCRUDOperationCLI("")
		passed = passed && result
	}

	return passed
}

func TestCFSConfigurationsCRUDOperationCLI(cliVersion string) (passed bool) {
	passed = true
	// Create a CFS configuration using CLI
	cfsConfigurationRecord, createSuccess := TestCLICFSConfigurationCreate(cliVersion)
	if !createSuccess && len(cfsConfigurationRecord.Name) == 0 {
		return false
	}

	currentTenant := common.GetTenantName()
	if len(cfsConfigurationRecord.Name) != 0 && !common.IsDummyTenant(currentTenant) && currentTenant != "" {
		NewTenant := GetAnotherTenantFromList(currentTenant)
		if len(NewTenant) != 0 {
			// Verify that the system admin is able to create a configuration of the same name, but belonging to tenant B
			createdWithAdmin := TestCLICFSConfigurationCreateByAdminWithSameNameDifferentTenant(cfsConfigurationRecord.Name, cliVersion, NewTenant)
			// Verify that the system admin is able to create a configuration of the same name, but that belongs to no tenant
			createdWithNoTenant := TestCLICFSConfigurationCreateByAdminWithSameNameDifferentTenant(cfsConfigurationRecord.Name, cliVersion, "")
			// Verify that the system admin is able to create a configuration of the same name, but belonging to tenant A
			createdWithOldTenant := TestCLICFSConfigurationCreateByAdminWithSameNameDifferentTenant(cfsConfigurationRecord.Name, cliVersion, currentTenant)
			passed = passed && createdWithAdmin && createdWithNoTenant && createdWithOldTenant
			// Seeting the tenant back to the original tenant
			common.SetTenantName(currentTenant)
		}
	}

	if len(cfsConfigurationRecord.Name) != 0 && !common.IsDummyTenant(common.GetTenantName()) && cliVersion == "v3" {
		createdWithTenant := TestCLICFSConfigurationCreateWithSameNameDifferentTenant(cfsConfigurationRecord.Name, cliVersion)
		updatedWithTenant := TestCLICFSConfigurationUpdateWithDifferentTenant(cfsConfigurationRecord.Name, cliVersion)
		deletedWithTenant := TestCLICFSConfigurationDeleteWithDifferentTenant(cfsConfigurationRecord.Name, cliVersion)
		passed = createdWithTenant && updatedWithTenant && deletedWithTenant
	}

	if len(cfsConfigurationRecord.Name) != 0 {
		// Update the CFS configuration using CLI
		updated := TestCLICFSConfigurationUpdate(cfsConfigurationRecord.Name, cliVersion)

		// Delete the CFS configuration using CLI
		deleted := TestCLICFSConfigurationDelete(cfsConfigurationRecord.Name, cliVersion)

		// Get all CFS configurations using CLI
		getAll := TestCLICFSConfigurationGetAll(cliVersion)

		passed = passed && updated && deleted && getAll && createSuccess

		return passed
	}
	return true
}

func TestCLICFSConfigurationCreateByAdminWithSameNameDifferentTenant(cfgName, cliVersion, newTenant string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Admin Creating CFS configuration with same name %s and different tenant.", cfgName))
	currentTenant := common.GetTenantName()
	var addTenant bool

	// Set the tenant name to the new tenant so that it is used for creation of paylaod
	if len(newTenant) != 0 {
		common.SetTenantName(newTenant)
		addTenant = true
	} else {
		addTenant = false
	}

	common.Infof("Creating CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	// Create CFS configuration payload
	fileName, payload, success, hasDummyData := CreateCFSConfigurationFile(cfgName, cliVersion, addTenant)
	if !success {
		return false
	}

	// Unsetting the tenant name to create a configuration that belongs to the new tenant using admin
	common.SetTenantName("")

	// Create a CFS configuration using CLI
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationCLI(cfgName, fileName, cliVersion)

	if !success {
		return false
	}

	// Remove the created configuration file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
	}

	// Set the new tenant for GET call to work
	common.SetTenantName(newTenant)

	// Verify CFS configuration record using CLI
	_, success = GetCFSConfigurationRecordCLI(cfgName, cliVersion)
	if !success {
		return false
	}

	// Verify CFS configurations in the list of configurations using CLI
	cfsConfigurations, success := GetCLIVersionBasedCFSConfigurationRecordList(cliVersion)
	if !success {
		common.Errorf("Unable to get CFS configurations list using CLI")
		return false
	}

	if !CFSConfigurationExists(cfsConfigurations, cfgName) {
		common.Errorf("CFS configuration %s was not found in the list of CFS configurations", cfgName)
		return false
	}

	// Verify the CFS configuration record
	if !VerifyCFSConfigurationRecord(cfsConfigurationRecord, payload, cfgName, cliVersion) {
		return false
	}

	common.Infof("Admin successfully updated CFS configuration %s tenant %s -> tenant %s", cfgName, currentTenant, newTenant)

	// if CreateCFSConfigurationFile has returned paylaod with fake data in it , return false
	if hasDummyData {
		return false
	}

	return true
}

func TestCLICFSConfigurationCreateWithSameNameDifferentTenant(cfgName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Creating CFS configuration with same name and different tenant: %s", cfgName))
	currentTenant := common.GetTenantName()
	newTenant := GetAnotherTenantFromList(currentTenant)

	if len(newTenant) == 0 {
		common.Warnf("No other tenant found to test CFS configuration creation with same name, skipping the test.")
		return true
	}

	common.SetTenantName(newTenant)

	common.Infof("Creating CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	// Create CFS configuration payload
	fileName, _, success, _ := CreateCFSConfigurationFile(cfgName, cliVersion, false)
	if !success {
		return false
	}
	// Set the CLI execution return code to 2, since creating the configuration with the same name with different tenant is expected to fail
	test.SetCliExecreturnCode(2)
	// Create a CFS configuration using CLI
	_, success = CreateUpdateCFSConfigurationCLI(cfgName, fileName, cliVersion)

	// Set the CLi execution code back to 0
	test.SetCliExecreturnCode(0)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if success {
		common.Errorf("Created CFS configuration %s using CLI with different tenant %s.", cfgName, newTenant)
		return false
	}

	// Remove the created configuration file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
	}

	passed = true
	common.Infof("Unable to create CFS configuration with same name %s for a different tenant: %s", cfgName, newTenant)
	return
}

func TestCLICFSConfigurationCreate(cliVersion string) (cfsConfigurationRecord CFSConfiguration, passed bool) {
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))
	passed = true

	common.PrintLog(fmt.Sprintf("Creating CFS configuration: %s", cfgName))

	// Get CFS configuration payload
	fileName, payload, success, hasDummyData := CreateCFSConfigurationFile(cfgName, cliVersion, false)
	if !success {
		return CFSConfiguration{}, false
	}

	common.Infof("Creating CFS configuration with file: %s", fileName)

	if common.IsDummyTenant(common.GetTenantName()) {
		// Set execution return code to 2, since dummy tenant is used
		test.SetCliExecreturnCode(2)
	}

	// Create a CFS configuration using CLI
	cfsConfigurationRecord, success = CreateUpdateCFSConfigurationCLI(cfgName, fileName, cliVersion)
	if !success {
		common.Errorf("Failed to create CFS configuration using CLI")
		// set the CLI execution return code to 0 if code execution reaches here and CLI execution return code was set to 2
		test.SetCliExecreturnCode(0)
		return CFSConfiguration{}, false
	}

	// Remove the created configuration file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
	}

	if test.GetCliExecreturnCode() != 0 {
		common.Infof("CFS configuration %s not successfully created with dummy tenant: %s", cfgName, common.GetTenantName())
		test.SetCliExecreturnCode(0)
		return CFSConfiguration{}, true // if the tenant is dummy, we skip the verification as creation is expected to fail
	}

	// Verify CFS configuration record using CLI
	_, success = GetCFSConfigurationRecordCLI(cfgName, cliVersion)
	if !success {
		return CFSConfiguration{}, false
	}

	// Verify CFS configurations in the list of configurations using CLI
	cfsConfigurations, success := GetCLIVersionBasedCFSConfigurationRecordList(cliVersion)
	if !success {
		common.Errorf("Unable to get CFS configurations list using CLI")
		return CFSConfiguration{}, false
	}

	if !CFSConfigurationExists(cfsConfigurations, cfgName) {
		common.Errorf("CFS configuration %s was not found in the list of cfs configurations", cfgName)
		return CFSConfiguration{}, false
	}

	// Verify the CFS configuration record
	if !VerifyCFSConfigurationRecord(cfsConfigurationRecord, payload, cfgName, cliVersion) {
		return CFSConfiguration{}, false
	}

	common.Infof("CFS configuration created successfully: %s", cfgName)
	// if CreateCFSConfigurationFile has returned paylaod with fake data in it , set the value of success to false
	// to make sure that the test case fails
	if hasDummyData {
		passed = false
	}
	return
}

func TestCLICFSConfigurationUpdateWithDifferentTenant(cfgName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS configuration %s with different tenant.", cfgName))
	currentTenant := common.GetTenantName()
	newTenant := GetAnotherTenantFromList(currentTenant)

	if len(newTenant) == 0 {
		common.Warnf("No other tenant found to test CFS configuration update with same name, skipping the test.")
		return true
	}

	common.SetTenantName(newTenant)

	common.Infof("Updating CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	// Get CFS configuration payload
	fileName, _, success, _ := CreateCFSConfigurationFile(cfgName, cliVersion, false)
	if !success {
		return false
	}
	// Set the CLI execution return code to 2, since updating the configuration with different tenant is expected to fail
	test.SetCliExecreturnCode(2)
	// Update a CFS configuration using CLI
	_, success = CreateUpdateCFSConfigurationCLI(cfgName, fileName, cliVersion)

	// Set the CLi execution code back to 0
	test.SetCliExecreturnCode(0)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if success {
		common.Errorf("Updated CFS configuration %s using CLI with different tenant %s.", cfgName, newTenant)
		return false
	}

	// Remove the created configuration file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
	}

	passed = true
	common.Infof("CFS configuration %s not updated using new tenant: %s", cfgName, newTenant)
	return
}

func TestCLICFSConfigurationUpdate(cfgName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS configuration: %s", cfgName))

	// Get CFS configuration payload
	fileName, payload, success, hasDummyData := CreateCFSConfigurationFile(cfgName, cliVersion, false)
	if !success {
		return false
	}

	// Update the CFS configuration using CLI
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationCLI(cfgName, fileName, cliVersion)
	if !success {
		common.Errorf("Failed to update CFS configuration using CLI")
		return false
	}

	// Remove the created configuration file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
	}

	// Verify CFS configuration record using CLI
	_, success = GetCFSConfigurationRecordCLI(cfgName, cliVersion)
	if !success {
		return false
	}

	// Verify the CFS configuration record
	if !VerifyCFSConfigurationRecord(cfsConfigurationRecord, payload, cfgName, cliVersion) {
		return false
	}

	// Verify CFS configurations in the list of configurations using CLI
	cfsConfigurations, success := GetCLIVersionBasedCFSConfigurationRecordList(cliVersion)
	if !success {
		common.Errorf("Unable to get CFS configurations list using CLI")
		return false
	}

	if !CFSConfigurationExists(cfsConfigurations, cfgName) {
		common.Errorf("CFS configuration %s was not found in the list of cfs configurations", cfgName)
		return false
	}

	common.Infof("CFS configuration updated successfully: %s", cfgName)

	// if CreateCFSConfigurationFile has returned paylaod with fake data in it , return false
	if hasDummyData {
		return false
	}

	return true
}

func TestCLICFSConfigurationDeleteWithDifferentTenant(cfgName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting CFS configuration %s with different tenant.", cfgName))
	currentTenant := common.GetTenantName()
	newTenant := GetAnotherTenantFromList(currentTenant)

	if len(newTenant) == 0 {
		common.Warnf("No other tenant found to test CFS configuration deletion with same name, skipping the test.")
		return true
	}

	common.SetTenantName(newTenant)

	common.Infof("Deleting CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	// Set the CLI execution return code to 2, since deleting the configuration with different tenant is expected to fail
	test.SetCliExecreturnCode(2)
	// Delete a CFS configuration using CLI
	success := DeleteCFSConfigurationRecordCLI(cfgName, cliVersion)

	// Set the CLi execution code back to 0
	test.SetCliExecreturnCode(0)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if success {
		common.Errorf("Deleted CFS configuration %s using CLI with different tenant %s.", cfgName, newTenant)
		return false
	}

	passed = true
	common.Infof("CFS configuration %s not deleted using new tenant: %s", cfgName, newTenant)
	return
}

func TestCLICFSConfigurationDelete(cfgName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting CFS configuration: %s", cfgName))

	// Delete the CFS configuration using CLI

	success := DeleteCFSConfigurationRecordCLI(cfgName, cliVersion)
	if !success {
		common.Errorf("Failed to delete CFS configuration using CLI")
		return false
	}

	// Set CLI execution return code to 2. Since cfs configuration is deleted, the command should return 2.
	test.SetCliExecreturnCode(2)
	// Verify CFS configuration record using CLI
	_, success = GetCFSConfigurationRecordCLI(cfgName, cliVersion)
	if success {
		common.Errorf("CFS configuration %s was not deleted successfully", cfgName)
		return false
	}

	// Set CLI execution return code to 0.
	test.SetCliExecreturnCode(0)

	// Verify CFS configurations in the list of configurations using CLI
	cfsConfigurations, success := GetCLIVersionBasedCFSConfigurationRecordList(cliVersion)
	if !success {
		common.Errorf("Unable to get CFS configurations list using CLI")
		return false
	}

	if CFSConfigurationExists(cfsConfigurations, cfgName) {
		common.Errorf("CFS configuration %s was found in the list of cfs configurations", cfgName)
		return false
	}

	common.Infof("CFS configuration deleted successfully: %s", cfgName)
	return true
}

func TestCLICFSConfigurationGetAll(cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Getting all CFS configurations"))

	// Get all CFS configurations using CLI
	cfsConfigurations, success := GetCLIVersionBasedCFSConfigurationRecordList(cliVersion)
	if !success {
		common.Errorf("Unable to get CFS configurations list using CLI")
		return false
	}

	common.Infof("Found %d CFS configurations", len(cfsConfigurations))
	return true
}
