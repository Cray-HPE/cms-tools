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

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	pcu "stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/prod-catalog-utils"
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
	expectedCreateHttpStatus := EXPECTED_CFS_CREATE_HTTP_STATUS
	tenantName := common.GetTenantName()
	if len(tenantName) != 0 {
		common.PrintLog(fmt.Sprintf("Running CFS configurations tests with tenant: %s", tenantName))
		if common.IsDummyTenant(tenantName) {
			expectedCreateHttpStatus = EXPECTED_CFS_BAD_REQUEST_HTTP_STATUS
		}
	} else {
		common.PrintLog("Running CFS configurations tests without tenant")
	}
	cfsConfigurationRecord, success := TestCFSConfigurationCreate(apiVersion, expectedCreateHttpStatus)
	// Create operation completes with dummy product catalog data but it is flagged as failure
	// By design first test with dummy product catalog data should fail but rest of the tests should run
	if !success && len(cfsConfigurationRecord.Name) == 0 {
		return false
	}

	passed = passed && success
	// Test Update, Delete, and Get with dummy tenant after creating with admin
	if len(cfsConfigurationRecord.Name) == 0 && apiVersion != "v2" {
		result := TestCFSConfigurationsCRUDOperationWithDummyTenant(apiVersion)
		passed = passed && result
	}

	currentTenant := common.GetTenantName()
	if len(cfsConfigurationRecord.Name) != 0 && !common.IsDummyTenant(currentTenant) && currentTenant != "" {
		NewTenant := GetAnotherTenantFromList(currentTenant)
		if len(NewTenant) != 0 {
			// Verify that the system admin is able to create a configuration of the same name, but belonging to tenant B
			createdWithAdmin := TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name, NewTenant, EXPECTED_CFS_CREATE_HTTP_STATUS)
			// Verify that the system admin is able to create a configuration of the same name, but that belongs to no tenant
			createdWithNoTenant := TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name, "", EXPECTED_CFS_CREATE_HTTP_STATUS)
			// Verify that the system admin is able to create a configuration of the same name, but belonging to tenant A
			createdWithOldTenant := TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name, currentTenant, EXPECTED_CFS_CREATE_HTTP_STATUS)
			passed = passed && createdWithAdmin && createdWithNoTenant && createdWithOldTenant
			// Seeting the tenant back to the original tenant
			common.SetTenantName(currentTenant)
		}
	}

	// Testing CFS configuration CRUD operation that does not belong to the same tenant
	if len(cfsConfigurationRecord.Name) != 0 && !common.IsDummyTenant(common.GetTenantName()) && apiVersion != "v2" {
		notCreatedWithTenant := TestCFSConfigurationCreateWithSameNameDifferentTenant(apiVersion, cfsConfigurationRecord.Name, EXPECTED_CFS_FORBIDDEN_HTTP_STATUS)
		notUpdatedWithTenant := TestCFSConfigurationUpdatewithDifferentTenant(apiVersion, cfsConfigurationRecord.Name, EXPECTED_CFS_FORBIDDEN_HTTP_STATUS)
		notDeletedWithTenant := TestCFSConfigurationDeleteUsingDifferentTenant(apiVersion, cfsConfigurationRecord.Name, EXPECTED_CFS_FORBIDDEN_HTTP_STATUS)
		passed = passed && notCreatedWithTenant && notUpdatedWithTenant && notDeletedWithTenant
	}

	if len(cfsConfigurationRecord.Name) != 0 {
		updated := TestCFSConfigurationUpdate(cfsConfigurationRecord.Name, apiVersion, EXPECTED_CFS_UPDATE_HTTP_STATUS)

		deleted := TestCFSConfigurationDelete(cfsConfigurationRecord.Name, apiVersion, EXPECTED_CFS_DELETE_HTTP_STATUS)

		getAll := TestCFSConfigurationGetAll(apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)

		return passed && updated && deleted && getAll
	}
	return true
}

// TestCFSConfigurationsCRUDOperationWithDummyTenant runs update, delete, get, and get all tests with a dummy tenant
// after creating the configuration with admin (no tenant).
// It takes apiVersion as parameter and returns true if all tests pass otherwise false.
// Note: This test is not run for v2 as v2 does not support tenant in the payload.
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
	cfsConfigurationRecord, createSuccess := CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, EXPECTED_CFS_CREATE_HTTP_STATUS)
	if !createSuccess {
		common.Errorf("CFS configuration %s not successfully created with Admin: %s, skipping other tests.", cfgName, common.GetTenantName())
		return false
	}

	// Check if GetCreateCFGConfigurationPayload returned paylaod with fake data in it and set the value of createSuccess to false
	//to make sure that the test case fails
	if pcu.IsUsingDummyData() {
		// Resetting the dummy data flag to false so that the failure is only reported once
		pcu.SetDummyDataFlag(false)
		createSuccess = false
	}

	common.SetTenantName(existingTenant)

	// Run CRUD operations with dummy tenant
	common.PrintLog(fmt.Sprintf("Running GET CFS configuration: %s test", cfsConfigurationRecord.Name))
	_, getFailed := GetCFSConfigurationRecordAPI(cfsConfigurationRecord.Name, apiVersion, EXPECTED_CFS_BAD_REQUEST_HTTP_STATUS)

	getAllFailed := TestCFSConfigurationGetAll(apiVersion, EXPECTED_CFS_BAD_REQUEST_HTTP_STATUS)

	notUpdated := TestCFSConfigurationUpdate(cfsConfigurationRecord.Name, apiVersion, EXPECTED_CFS_BAD_REQUEST_HTTP_STATUS)

	notDeleted := TestCFSConfigurationDelete(cfsConfigurationRecord.Name, apiVersion, EXPECTED_CFS_BAD_REQUEST_HTTP_STATUS)

	// Delete the configuration using admin
	common.SetTenantName("")
	common.Infof("Deleting CFS configuration %s using Admin.", cfsConfigurationRecord.Name)
	success = DeleteCFSConfigurationRecordAPI(cfgName, apiVersion, EXPECTED_CFS_DELETE_HTTP_STATUS)
	if !success {
		common.Infof("CFS configuration %s not successfully deleted with Admin: %s", cfgName)
	}

	passed = passed && notUpdated && notDeleted && getFailed && getAllFailed && success && createSuccess

	// Setting the tenant back to the original tenant
	common.SetTenantName(existingTenant)

	return
}

// TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant verifies that the system admin is able to create a configuration of the same name,
// but belonging to a different tenant or no tenant at all.
// It takes apiVersion, cfs config name, new tenant name, and expected http status code as parameters.
// it returns true for following cases otherwise false:
// - If the create was successful and following verifications are successful
//   - The created configuration can be retrieved using GET operation
//   - The created configuration is present in the list of configurations
//   - The created configuration matches the payload used for create operation
func TestCFSConfigurationCreateByAdminWithSameNameDifferentTenant(apiVersion, cfgName, newTenant string, expectedHttpStatus int) (success bool) {
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
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, expectedHttpStatus)
	if !success {
		return false
	}

	// Set the new tenant for GET call to work
	common.SetTenantName(newTenant)
	// verify Cfs configuration record
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)
	if !success {
		common.Errorf("Unable to get CFS configuration record: %s", cfgName)
		return false
	}

	// Verify cfs configurations in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)
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

	// Check if GetCreateCFGConfigurationPayload returned paylaod with fake data in it, return false
	//to make sure that the test case fails
	if pcu.IsUsingDummyData() {
		// Resetting the dummy data flag to false so that the failure is only reported once
		pcu.SetDummyDataFlag(false)
		return false
	}

	return true
}

// TestCFSConfigurationCreateWithSameNameDifferentTenant attempts to create a CFS configuration with the same name but belonging to a different tenant.
// It takes apiVersion, cfs config name, and expected http status code as parameters.
// it returns true for following cases otherwise false:
// - If the create operation is performed using non-owner tenant and create is not successful and
// expectedHttpStatus matches the actual http status code
func TestCFSConfigurationCreateWithSameNameDifferentTenant(apiVersion, cfgName string, expectedHttpStatus int) (success bool) {

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
	_, success = CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, expectedHttpStatus)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if !success {
		common.Errorf("Create CFS configuration successful with same name %s for a different tenant: %s", cfgName, newTenant)
		return false
	}
	common.Infof("Unable to create CFS configuration with same name %s for a different tenant: %s", cfgName, newTenant)
	return true
}

// TestCFSConfigurationCreate creates a new CFS configuration.
// It takes apiVersion and expected http status code as parameters.
// it returns true for following cases otherwise false:
// - If the create was successful and following verifications are successful
//   - The created configuration can be retrieved using GET operation
//   - The created configuration is present in the list of configurations
//   - The created configuration matches the payload used for create operation
//
// - If the create operation is performed using dummy or non-owner tenant and create is not successful
// and expectedHttpStatus matches the actual http status code
func TestCFSConfigurationCreate(apiVersion string, expectedHttpStatus int) (cfsConfigurationRecord CFSConfiguration, success bool) {
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	common.PrintLog(fmt.Sprintf("Creating CFS configuration: %s", cfgName))

	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, false)
	if !success {
		return CFSConfiguration{}, false
	}

	// create CFS configuration
	cfsConfigurationRecord, success = CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, expectedHttpStatus)
	if !success {
		return CFSConfiguration{}, false
	}

	// We need to determine if we should continue on to verify the configuration record,
	// or if we should stop here.
	//
	// If the call to CreateUpdateCFSConfigurationRecordAPI returns success = false, then
	// we never reach this point. That means that success = true. That function only returns
	// success = true if the API call was made and its status code matched our expected
	// status code.
	//
	// We want to verify the response in the case that the API call was successful. All
	// successful API calls have status 2xx. So if our expected status was outside of this
	// range, and we reached this point, then we know the API call must have failed.
	if expectedHttpStatus > 299 || expectedHttpStatus < 200 {
		// This means the API call failed, so we should not verify the response
		common.Infof("CFS configuration %s not successfully created with dummy tenant: %s", cfgName, common.GetTenantName())
		return CFSConfiguration{}, true
	}

	// verify Cfs configuration record
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)
	if !success {
		common.Errorf("Unable to get CFS configuration record: %s", cfgName)
		return CFSConfiguration{}, false
	}

	// Verify cfs configurations in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)
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

	// Check if GetCreateCFGConfigurationPayload returned paylaod with fake data in it and set the value of success to false
	//to make sure that the test case fails
	if pcu.IsUsingDummyData() {
		// Resetting the dummy data flag to false so that the failure is only reported once
		pcu.SetDummyDataFlag(false)
		success = false
	}

	return
}

// TestCFSConfigurationUpdatewithDifferentTenant attempts to update an existing CFS configuration using a different non-owner tenant.
// It takes cfs config name, apiVersion, and expected http status code as parameters.
// it returns true for following cases otherwise false:
// - If the update operation is performed using Non owner tenant and update is not successful and
// expectedHttpStatus matches the actual http status code
// - If new tenant is not found to perform the test
func TestCFSConfigurationUpdatewithDifferentTenant(apiVersion, cfgName string, expectedHttpStatus int) (success bool) {
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
	_, success = CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, expectedHttpStatus)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if !success {
		common.Errorf("Successfully updated CFS configuration %s using new tenant: %s", cfgName, newTenant)
		return false
	}
	common.Infof("CFS configuration %s not updated using new tenant: %s", cfgName, newTenant)

	return true
}

// TestCFSConfigurationUpdate updates an existing CFS configuration.
// It takes cfs config name, apiVersion, and expected http status code as parameters.
// it returns true for following cases otherwise false:
// - If the update was successful and following verifications are successful
//   - The updated configuration can be retrieved using GET operation
//   - The updated configuration is present in the list of configurations
//   - The updated configuration matches the payload used for update operation
//
// - If the update operation is performed using dummy or non-owner tenant and update is not successful and
// expectedHttpStatus matches the actual http status code
func TestCFSConfigurationUpdate(cfgName, apiVersion string, expectedHttpStatus int) (success bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS configuration: %s", cfgName))
	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion, false)
	if !success {
		return false
	}

	// Update CFS configuration record
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload, expectedHttpStatus)
	if !success {
		return false
	}

	// We need to determine if we should continue on to verify the configuration record,
	// or if we should stop here.
	//
	// If the call to CreateUpdateCFSConfigurationRecordAPI returns success = false, then
	// we never reach this point. That means that success = true. That function only returns
	// success = true if the API call was made and its status code matched our expected
	// status code.
	//
	// We want to verify the response in the case that the API call was successful. All
	// successful API calls have status 2xx. So if our expected status was outside of this
	// range, and we reached this point, then we know the API call must have failed.
	if expectedHttpStatus > 299 || expectedHttpStatus < 200 {
		// This means the API call failed, so we should not verify the response
		common.Infof("CFS configuration %s not successfully updated with dummy tenant: %s", cfgName, common.GetTenantName())
		return true
	}

	// Verify the CFS configuration record
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)
	if !success {
		return false
	}

	// Verify cfs configurations in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)
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

	// Check if GetCreateCFGConfigurationPayload returned paylaod with fake data in it , return false
	//to make sure that the test case fails
	if pcu.IsUsingDummyData() {
		// Resetting the dummy data flag to false so that the failure is only reported once
		pcu.SetDummyDataFlag(false)
		return false
	}

	return true
}

// TestCFSConfigurationDeleteUsingDifferentTenant attempts to delete an existing CFS configuration using a different tenant.
// It takes cfs config name, apiVersion, and expected http status code as parameters.
// it returns true for following cases otherwise false:
// - If the delete operation is performed using non-owner tenant and delete is not successful and
// expectedHttpStatus matches the actual http status code
// - If new tenant is not found to perform the test
func TestCFSConfigurationDeleteUsingDifferentTenant(apiVersion, cfgName string, expectedHttpStatus int) (success bool) {
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
	success = DeleteCFSConfigurationRecordAPI(cfgName, apiVersion, expectedHttpStatus)
	// Reset tenant name to the original tenant
	common.SetTenantName(currentTenant)

	if !success {
		common.Errorf("Successfully deleted CFS configuration %s using new tenant: %s", cfgName, newTenant)
		return false
	}
	common.Infof("CFS configuration %s not deleted using new tenant: %s", cfgName, newTenant)

	return true
}

// TestCFSConfigurationDelete deletes an existing CFS configuration.
// It takes cfs config name, apiVersion, and expected http status code as parameters.
// Delete operation is expected to succeed if performed using admin or the owner tenant
// it returns true for following cases otherwise false:
// - If the delete was successful and following verifications are successful
//   - The deleted configuration cannot be retrieved using GET operation
//   - The deleted configuration is not present in the list of configurations
//
// - If the delete operation is performed using dummy or non-owner tenant and delete is not successful and
// expectedHttpStatus matches the actual http status code
func TestCFSConfigurationDelete(cfgName string, apiVersion string, expectedHttpStatus int) (success bool) {
	common.PrintLog(fmt.Sprintf("Deleting CFS configuration: %s", cfgName))
	// Delete CFS configuration record
	success = DeleteCFSConfigurationRecordAPI(cfgName, apiVersion, expectedHttpStatus)
	if !success {
		return false
	}

	// We need to determine if we should continue on to verify that the configuration,
	//  record was actually deleted, or if we should stop here.
	//
	// If the call to DeleteCFSConfigurationRecordAPI returns success = false, then
	// we never reach this point. That means that success = true. That function only returns
	// success = true if the API call was made and its status code matched our expected
	// status code.
	//
	// We want to verify that the delete actually happened in the case that the API call was
	// successful. All successful API calls have status 2xx. So if our expected status was
	// outside of this range, and we reached this point, then we know the API call must
	// have failed.
	if expectedHttpStatus > 299 || expectedHttpStatus < 200 {
		// This means the API call failed, so we should not verify the delete
		common.Infof("CFS configuration %s not successfully created with dummy tenant: %s", cfgName, common.GetTenantName())
		return true
	}

	// Verify CFS configuration record is deleted
	_, success = GetCFSConfigurationRecordAPI(cfgName, apiVersion, EXPECTED_CFS_NOT_FOUND_HTTP_STATUS)
	if !success {
		return false
	}

	// Verify cfs configurations is not in the list of configurations
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion, EXPECTED_CFS_GET_HTTP_STATUS)
	if !success {
		return false
	}

	if CFSConfigurationExists(cfsConfigurations, cfgName) {
		return false
	}
	common.Infof("CFS configuration %s deleted successfully", cfgName)

	return true
}

// TestCFSConfigurationGetAll gets all existing CFS configurations using the given apiVersion and expected http status code.
// It returns true if the GET ALL API call's status code matches the expected http status code, otherwise false.
func TestCFSConfigurationGetAll(apiVersion string, expectedHttpStatus int) (success bool) {
	common.PrintLog("Getting all CFS configurations")
	// Get CFS configurations list
	cfsConfigurations, success := GetAPIBasedCFSConfigurationRecordList(apiVersion, expectedHttpStatus)
	if !success {
		return false
	}

	common.Infof("Found %d CFS configurations", len(cfsConfigurations))

	return true
}
