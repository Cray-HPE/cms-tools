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
 * verify_cfs_configurations.go
 *
 * cfs configurations test api functions
 *
 */
package cfs

import (
	"fmt"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func TestCFSConfigurationsCRUDOperationWithTenantsUsingAPIVersions() (passed bool) {
	passed = TestCFSConfigurationsCRUDOperationUsingAPIVersions()
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
		result := TestCFSConfigurationsCRUDOperationUsingAPIVersions()
		passed = passed && result
		common.SetTenantName("")
	}
	return passed

}

func TestCFSConfigurationsCRUDOperationUsingAPIVersions() (passed bool) {
	passed = true
	// Get supported API versions for configurations endpoints
	for _, apiVersion := range GetSupportAPIVersions("configurations") {
		if common.GetTenantName() == "" || apiVersion != "v2" {
			common.PrintLog(fmt.Sprintf("Testing CFS configurations CRUD operations using API version: %s", apiVersion))
			result := TestCFSConfigurationsCRUDOperation(apiVersion)
			passed = passed && result
		}
	}

	return passed
}

func TestCFSConfigurationsCRUDOperation(apiVersion string) (passed bool) {
	passed = true
	if len(common.GetTenantName()) != 0 {
		common.PrintLog(fmt.Sprintf("Running CFS configurations tests with tenant: %s", common.GetTenantName()))
	} else {
		common.PrintLog("Running CFS configurations tests without tenant")
	}
	cfsConfigurationRecord, success := TestCFSConfigurationCreate(apiVersion)
	if !success {
		return false
	}

	// Test Update, Delete and get with dummy tenant after creating with Admin
	if len(cfsConfigurationRecord.Name) == 0 && apiVersion != "v2" {
		result := TestCFSConfigurationsCRUDOperationWithDummyTenant(apiVersion)
		passed = passed && result
	}

	currentTenant := common.GetTenantName()
	if len(cfsConfigurationRecord.Name) != 0 && !common.IsDummyTenant(currentTenant) && currentTenant != "" {
		NewTenant := GetAnotherTenantFromList(currentTenant)
		if len(NewTenant) != 0 {
			// Verify that the system admin is able to create a configuration of the same name, but belonging to tenant B
			createdWithAdmin := TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name, NewTenant)
			// Verify that the system admin is able to create a configuration of the same name, but that belongs to no tenant
			createdWithNoTenant := TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name, "")
			// Verify that the system admin is able to create a configuration of the same name, but belonging to tenant A
			createdWithOldTenant := TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name, currentTenant)
			passed = passed && createdWithAdmin && createdWithNoTenant && createdWithOldTenant
			// Seeting the tenant back to the original tenant
			common.SetTenantName(currentTenant)
		}
	}

	// Testing CFS configuration CRUD operation that does not belong to the same tenant
	if len(cfsConfigurationRecord.Name) != 0 && !common.IsDummyTenant(common.GetTenantName()) && apiVersion != "v2" {

		createdWithTenant := TestCFSConfigurationCreateWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name)
		updatedWithTenant := TestCFSConfigurationUpdatewithDifferentTenant(apiVersion, cfsConfigurationRecord.Name)
		deletedWithTenant := TestCFSConfigurationDeleteUsingDifferentTenant(apiVersion, cfsConfigurationRecord.Name)
		passed = createdWithTenant && updatedWithTenant && deletedWithTenant
	}

	if len(cfsConfigurationRecord.Name) != 0 {
		updated := TestCFSConfigurationUpdate(cfsConfigurationRecord.Name, apiVersion)

		deleted := TestCFSConfigurationDelete(cfsConfigurationRecord.Name, apiVersion)

		getAll := TestCFSConfigurationGetAll(apiVersion)

		return passed && updated && deleted && getAll
	}
	return true
}

func TestCFSConfigurationsCRUDOperationWithDummyTenant(apiVersion string) (passed bool) {
	passed = true
	existingTenant := common.GetTenantName()
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))
	common.PrintLog(fmt.Sprintf("Running CFS configurations update, Delete tests with dummy tenant: %s after creating with Admin.", common.GetTenantName()))

	// get CFS configuration payload with tenane_name set in the payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, true)
	if !success {
		common.Infof("Unable to get CFS configuration payload, skippinhg the test.")
		return true
	}

	// Unsetting the tenant name to create a configuration that belongs to the new tenant using admin
	common.SetTenantName("")

	// Attempt to create CFS configuration with the same name in a different tenant
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, http.StatusOK)
	if !success {
		common.Errorf("CFS configuration %s not successfully created with Admin: %s, skipping other tests.", cfgName, common.GetTenantName())
		return false
	}

	common.SetTenantName(existingTenant)

	// Run CRUD operations with dummy tenant
	common.PrintLog(fmt.Sprintf("Running GET CFS configuration: %s test", cfsConfigurationRecord.Name))
	_, get := GetCFSConfigurationRecordAPI(cfsConfigurationRecord.Name, apiVersion, http.StatusOK)

	getAll := TestCFSConfigurationGetAll(apiVersion)

	updated := TestCFSConfigurationUpdate(cfsConfigurationRecord.Name, apiVersion)

	deleted := TestCFSConfigurationDelete(cfsConfigurationRecord.Name, apiVersion)

	passed = passed && !updated && !deleted && !get && !getAll

	// Delete the configuration using admin
	common.SetTenantName("")
	success = DeleteCFSConfigurationRecordAPI(cfgName, apiVersion, http.StatusNoContent)
	if !success {
		common.Infof("CFS configuration %s not successfully deleted with Admin: %s", cfgName)
	}
	// Setting the tenant back to the original tenant
	common.SetTenantName(existingTenant)

	return
}

func TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfgName, newTenant string) (success bool) {
	common.PrintLog(fmt.Sprintf("Admin Creating CFS configuration with same name %s and different tenant.", cfgName))
	var addTenant bool

	currentTenant := common.GetTenantName()

	// Set the tenant name to the new tenant so that it is used for creation of paylaod
	if len(newTenant) != 0 {
		common.SetTenantName(newTenant)
		addTenant = true
	} else {
		addTenant = false
	}

	// get CFS configuration payload with tenane_name set in the payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, addTenant)
	if !success {
		return false
	}

	common.Infof("Admin Creating CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	// Unsetting the tenant name to create a configuration that belongs to the new tenant using admin
	common.SetTenantName("")
	// Attempt to create CFS configuration with the same name in a different tenant
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, http.StatusOK)
	if !success {
		return false
	}

	// Set the new tenant for GET call to work
	common.SetTenantName(newTenant)
	// verify Cfs configuration record
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, http.StatusOK)
	if !success {
		common.Errorf("Unable to get CFS configuration record: %s", cfgName)
		return false
	}

	// Verify cfs configurations in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion)
	if !success {
		common.Errorf("Unable to get CFS configurations list")
		return false
	}

	if !CFSConfigurationExists(cfsConfigurations, cfgName) {
		common.Errorf("CFS configuration %s was not found in the list of CFS configurations", cfgName)
		return false
	}

	// Verify the CFS configuration record
	if !VerifyCFSConfigurationRecord(cfsConfigurationRecord, cfsConfigurationPayload, cfgName, apiVersion) {
		return false
	}
	common.Infof("Admin successfully updated CFS configuration %s tenant %s -> tenant %s", cfgName, currentTenant, newTenant)
	return true
}

func TestCFSConfigurationCreateWithSameNameDifferentTenant(apiVersion, cfgName string) (success bool) {

	common.PrintLog(fmt.Sprintf("Creating CFS configuration with same name %s and different tenant", cfgName))

	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, false)
	if !success {
		return false
	}

	currentTenant := common.GetTenantName()
	newTenant := GetAnotherTenantFromList(currentTenant)

	if len(newTenant) == 0 {
		common.Warnf("No other tenant found to test CFS configuration creation with same name, skipping the test.")
		return true
	}

	common.Infof("Creating CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	common.SetTenantName(newTenant)
	// Attempt to create CFS configuration with the same name in a different tenant
	_, success = CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, http.StatusForbidden)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if !success {
		common.Errorf("Create CFS configuration successful with same name %s for a different tenant: %s", cfgName, newTenant)
		return false
	}
	common.Infof("Unable to create CFS configuration with same name %s for a different tenant: %s", cfgName, newTenant)
	return true
}

func TestCFSConfigurationCreate(apiVersion string) (cfsConfigurationRecord CFSConfiguration, success bool) {
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	common.PrintLog(fmt.Sprintf("Creating CFS configuration: %s", cfgName))

	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, false)
	if !success {
		return CFSConfiguration{}, false
	}

	// create CFS configuration
	cfsConfigurationRecord, success = CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, http.StatusOK)
	if !success {
		return CFSConfiguration{}, false
	}

	// If the create operation is performed using dummy tenant, we expect the status code not to be 200.
	if GetExpectedHTTPStatusCode() != http.StatusOK {
		common.Infof("CFS configuration %s not successfully created with dummy tenant: %s", cfgName, common.GetTenantName())
		return CFSConfiguration{}, true
	}

	// verify Cfs configuration record
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, http.StatusOK)
	if !success {
		common.Errorf("Unable to get CFS configuration record: %s", cfgName)
		return CFSConfiguration{}, false
	}

	// Verify cfs configurations in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion)
	if !success {
		common.Errorf("Unable to get CFS configurations list")
		return CFSConfiguration{}, false
	}

	if !CFSConfigurationExists(cfsConfigurations, cfgName) {
		common.Errorf("CFS configuration %s was not found in the list of cfs configurations", cfgName)
		return CFSConfiguration{}, false
	}

	// Verify the CFS configuration record
	if !VerifyCFSConfigurationRecord(cfsConfigurationRecord, cfsConfigurationPayload, cfgName, apiVersion) {
		return CFSConfiguration{}, false
	}
	common.Infof("CFS configuration record created successfully: %s", cfsConfigurationRecord.Name)

	return cfsConfigurationRecord, true
}

func TestCFSConfigurationUpdatewithDifferentTenant(apiVersion, cfgName string) (success bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS configuration %s with a non owner tenant.", cfgName))

	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, false)
	if !success {
		return false
	}

	currentTenant := common.GetTenantName()
	newTenant := GetAnotherTenantFromList(currentTenant)

	if len(newTenant) == 0 {
		common.Warnf("No other tenant found to test CFS configuration update with same name, skipping the test.")
		return true
	}

	common.Infof("Updating CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	common.SetTenantName(newTenant)
	// Attempt to update CFS configuration with the same name in a different tenant
	_, success = CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, http.StatusForbidden)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if !success {
		common.Errorf("Successfully updated CFS configuration %s using new tenant: %s", cfgName, newTenant)
		return false
	}
	common.Infof("CFS configuration %s not updated using new tenant: %s", cfgName, newTenant)

	return true
}

func TestCFSConfigurationUpdate(cfgName, apiVersion string) (success bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS configuration: %s", cfgName))
	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, false)
	if !success {
		return false
	}

	// Update CFS configuration record
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, http.StatusOK)
	if !success {
		return false
	}

	// If the update operation is performed using dummy tenant, we expect the status code not to be 200.
	if GetExpectedHTTPStatusCode() != http.StatusOK {
		common.Infof("CFS configuration %s not successfully updated with dummy tenant: %s", cfgName, common.GetTenantName())
		return true
	}

	// Verify the CFS configuration record
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, http.StatusOK)
	if !success {
		return false
	}

	// Verify cfs configurations in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion)
	if !success {
		return false
	}

	if !CFSConfigurationExists(cfsConfigurations, cfgName) {
		return false
	}

	if !VerifyCFSConfigurationRecord(cfsConfigurationRecord, cfsConfigurationPayload, cfgName, apiVersion) {
		return false
	}
	common.Infof("CFS configuration record updated successfully: %s", cfsConfigurationRecord.Name)

	return true
}

func TestCFSConfigurationDeleteUsingDifferentTenant(apiVersion, cfgName string) (success bool) {
	common.PrintLog(fmt.Sprintf("Deleting CFS configuration %s using a different tenant", cfgName))

	// Get another tenant
	currentTenant := common.GetTenantName()
	newTenant := GetAnotherTenantFromList(currentTenant)

	if len(newTenant) == 0 {
		common.Warnf("No other tenant found to test CFS configuration deletion with same name, skipping the test.")
		return true
	}

	common.Infof("Deleting CFS configuration %s belonging to tenant %s using new tenant %s", cfgName, currentTenant, newTenant)

	common.SetTenantName(newTenant)
	// Attempt to delete CFS configuration with the same name in a different tenant
	success = DeleteCFSConfigurationRecordAPI(cfgName, apiVersion, http.StatusForbidden)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if !success {
		common.Errorf("Successfully deleted CFS configuration %s using new tenant: %s", cfgName, newTenant)
		return false
	}
	common.Infof("CFS configuration %s not deleted using new tenant: %s", cfgName, newTenant)

	return true
}

func TestCFSConfigurationDelete(cfgName string, apiVersion string) (success bool) {
	common.PrintLog(fmt.Sprintf("Deleting CFS configuration: %s", cfgName))
	// Delete CFS configuration record
	success = DeleteCFSConfigurationRecordAPI(cfgName, apiVersion, http.StatusNoContent)
	if !success {
		return false
	}

	// / If the delete operation is performed using dummy tenant, we expect the status code not to be 204.
	if GetExpectedHTTPStatusCode() != http.StatusNoContent {
		common.Infof("CFS configuration %s not successfully deleted with dummy tenant: %s", cfgName, common.GetTenantName())
		return true
	}

	// Verify CFS configuration record is deleted
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, http.StatusNotFound)
	if !success {
		return false
	}

	// Verify cfs configurations is not in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion)
	if !success {
		return false
	}

	if CFSConfigurationExists(cfsConfigurations, cfgName) {
		return false
	}
	common.Infof("CFS configuration %s deleted successfully", cfgName)

	return true
}

func TestCFSConfigurationGetAll(apiVersion string) (success bool) {
	common.PrintLog("Getting all CFS configurations")
	// Get CFS configurations list
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion)
	if !success {
		return false
	}

	common.Infof("Found %d CFS configurations", len(cfsConfigurations))

	return true
}
