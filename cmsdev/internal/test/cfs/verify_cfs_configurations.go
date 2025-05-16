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

func TestCFSConfigurationsCRUDOperationUsingAPIVersions() (passed bool) {
	passed = true
	// Get supported API versions for configurations endpoints
	for _, apiVersion := range GetSupportAPIVersions("configurations") {
		common.PrintLog(fmt.Sprintf("Testing CFS configurations CRUD operations using API version: %s", apiVersion))
		passed = passed && TestCFSConfigurationsCRUDOperation(apiVersion)
	}

	return passed
}

func TestCFSConfigurationsCRUDOperation(apiVersion string) (passed bool) {
	cfsConfigurationRecord, success := TestCFSConfigurationCreate(apiVersion)
	if !success {
		return false
	}

	updated := TestCFSConfigurationUpdate(cfsConfigurationRecord.Name, apiVersion)

	deleted := TestCFSConfigurationDelete(cfsConfigurationRecord.Name, apiVersion)

	getAll := TestCFSConfigurationGetAll(apiVersion)

	return updated && deleted && getAll
}

func TestCFSConfigurationCreate(apiVersion string) (cfsConfigurationRecord CFSConfiguration, success bool) {
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	common.PrintLog(fmt.Sprintf("Creating CFS configuration: %s", cfgName))

	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion)
	if !success {
		return CFSConfiguration{}, false
	}

	// create CFS configuration
	cfsConfigurationRecord, success = CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload)
	if !success {
		return CFSConfiguration{}, false
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

func TestCFSConfigurationUpdate(cfgName, apiVersion string) (success bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS configuration: %s", cfgName))
	// get CFS configuration payload
	cfsConfigurationPayload, success := GetCreateCFGConfigurationPayload(apiVersion)
	if !success {
		return false
	}

	// Update CFS configuration record
	cfsConfigurationRecord, success := CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, cfsConfigurationPayload)
	if !success {
		return false
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

func TestCFSConfigurationDelete(cfgName string, apiVersion string) (success bool) {
	common.PrintLog(fmt.Sprintf("Deleting CFS configuration: %s", cfgName))
	// Delete CFS configuration record
	success = DeleteCFSConfigurationRecordAPI(cfgName, apiVersion)
	if !success {
		return false
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
