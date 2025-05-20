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

func TestCFSConfigurationsCRUDOperationUsingCLI() (passed bool) {
	passed = true
	// Get supported API versions for configurations endpoints
	for _, cliVersion := range GetSupportAPIVersions("configurations") {
		common.PrintLog(fmt.Sprintf("Testing CFS configurations CRUD operations using CLI and version: %s", cliVersion))
		passed = passed && TestCFSConfigurationsCRUDOperationCLI(cliVersion)
	}

	// Test default CLI version
	common.PrintLog("Testing CFS configurations CRUD operations using default CLI version")
	passed = passed && TestCFSConfigurationsCRUDOperationCLI("")

	return passed
}

func TestCFSConfigurationsCRUDOperationCLI(cliVersion string) (passed bool) {
	passed = true
	// Create a CFS configuration using CLI
	cfsConfigurationRecord, success := TestCLICFSConfigurationCreate(cliVersion)
	if !success {
		return false
	}

	// Update the CFS configuration using CLI
	updated := TestCLICFSConfigurationUpdate(cfsConfigurationRecord.Name, cliVersion)

	// Delete the CFS configuration using CLI
	deleted := TestCLICFSConfigurationDelete(cfsConfigurationRecord.Name, cliVersion)

	// Get all CFS configurations using CLI
	getAll := TestCLICFSConfigurationGetAll(cliVersion)

	return updated && deleted && getAll
}

func TestCLICFSConfigurationCreate(cliVersion string) (cfsConfigurationRecord CFSConfiguration, passed bool) {
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	common.PrintLog(fmt.Sprintf("Creating CFS configuration: %s", cfgName))

	// Get CFS configuration payload
	fileName, payload, success := CreateCFSConfigurationFile(cfgName, cliVersion)
	if !success {
		return CFSConfiguration{}, false
	}

	// Create a CFS configuration using CLI
	cfsConfigurationRecord, success = CreateUpdateCFSConfigurationCLI(cfgName, fileName, cliVersion)
	if !success {
		common.Errorf("Failed to create CFS configuration using CLI")
		return CFSConfiguration{}, false
	}

	// Remove the created configuration file
	if err := os.Remove(fileName); err != nil {
		common.Errorf("Unable to remove file %s: %v", fileName, err)
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
	return cfsConfigurationRecord, true
}

func TestCLICFSConfigurationUpdate(cfgName, cliVersion string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS configuration: %s", cfgName))

	// Get CFS configuration payload
	fileName, payload, success := CreateCFSConfigurationFile(cfgName, cliVersion)
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
	return true
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
