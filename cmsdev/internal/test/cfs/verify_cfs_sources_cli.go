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
 * verify_cfs_sources_cli.go
 *
 * cfs sources test cli functions
 *
 */
package cfs

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func TestCFSSourcesCRUDOperationCLI() (passed bool) {
	passed = true
	// Create a CFS configuration using CLI
	cfsConfigurationRecord, success := TestCLICFSSourcesCreate("v3")
	if !success {
		return false
	}

	// Update the CFS configuration using CLI
	updated := TestCLICFSSourcesUpdate(cfsConfigurationRecord.Name, "v3")

	// Delete the CFS configuration using CLI
	deleted := TestCLICFSSourcesDelete(cfsConfigurationRecord.Name, "v3")

	// Get all CFS configurations using CLI
	getAll := TestCLICFSSourcesGetAll("v3")

	return passed && updated && deleted && getAll
}

func TestCLICFSSourcesCreate(cliVersion string) (cfsSourceRecord CFSSources, passed bool) {
	passed = true
	sourceName := "CFS_Source_" + string(common.GetRandomString(10))
	common.PrintLog("Creating CFS source using CLI: " + sourceName)

	// Create a new CFS source record
	cfsSourceRecord, success := CreateCFSSourceRecordCLI(sourceName, firstCloneURL, cliVersion)
	if !success {
		return CFSSources{}, false
	}

	// Get the CFS source record
	_, success = GetCFSSourceRecordCLI(sourceName, cliVersion)
	if !success {
		common.Errorf("Failed to get CFS source record: %s", sourceName)
		return CFSSources{}, false
	}

	// verify CFS source is in the list of sources
	cfsSourceList, success := GetCFSSourcesListCLI(cliVersion)
	if !success {
		common.Errorf("Failed to get CFS sources list")
		return CFSSources{}, false
	}

	// Check if the source is in the list
	if !CFSSourceExists(cfsSourceList, sourceName) {
		return CFSSources{}, false
	}

	// verify the source record
	if !VerifyCFSSourceRecord(cfsSourceRecord, sourceName, firstCloneURL) {
		return CFSSources{}, false
	}

	common.Infof("CFS source %s created successfully", sourceName)
	return cfsSourceRecord, true
}

func TestCLICFSSourcesUpdate(sourceName, cliVersion string) (passed bool) {
	common.PrintLog("Updating CFS source using CLI: " + sourceName)

	// Update the CFS source record
	cfsSourceRecord, success := UpdateCFSSourceRecordCLI(sourceName, secondCloneURL, cliVersion)
	if !success {
		return false
	}

	// Get the CFS source record
	_, success = GetCFSSourceRecordCLI(sourceName, cliVersion)
	if !success {
		common.Errorf("Failed to get CFS source record: %s", sourceName)
		return false
	}

	// verify the source record
	if !VerifyCFSSourceRecord(cfsSourceRecord, sourceName, secondCloneURL) {
		return false
	}

	// verify CFS source is in the list of sources
	cfsSourceList, success := GetCFSSourcesListCLI(cliVersion)
	if !success {
		common.Errorf("Failed to get CFS sources list")
		return false
	}

	// Check if the source is in the list
	if !CFSSourceExists(cfsSourceList, sourceName) {
		return false
	}

	common.Infof("CFS source %s updated successfully", sourceName)
	return true
}

func TestCLICFSSourcesDelete(sourceName, cliVersion string) (passed bool) {
	common.PrintLog("Deleting CFS source using CLI: " + sourceName)

	// Delete the CFS source record
	success := DeleteCFSSourceRecordCLI(sourceName, cliVersion)
	if !success {
		return false
	}

	// Set CLI execution return code to 2. Since cfs configuration is deleted, the command should return 2.
	test.SetCliExecreturnCode(2)

	// Get the CFS source record
	_, success = GetCFSSourceRecordCLI(sourceName, cliVersion)
	if success {
		common.Errorf("CFS source %s was not deleted successfully", sourceName)
		return false
	}

	// Set CLI execution return code to 0.
	test.SetCliExecreturnCode(0)

	// verify CFS source is not in the list of sources
	cfsSourceList, success := GetCFSSourcesListCLI(cliVersion)
	if !success {
		common.Errorf("Failed to get CFS sources list")
		return false
	}

	// Check if the source is in the list
	if CFSSourceExists(cfsSourceList, sourceName) {
		common.Errorf("CFS source %s was found in the list of cfs sources", sourceName)
		return false
	}

	common.Infof("CFS source %s deleted successfully", sourceName)
	return true
}

func TestCLICFSSourcesGetAll(cliVersion string) (passed bool) {
	common.PrintLog("Getting all CFS sources using CLI")

	// Get all CFS sources using CLI
	cfsSourceList, success := GetCFSSourcesListCLI(cliVersion)
	if !success {
		common.Errorf("Unable to get CFS sources list using CLI")
		return false
	}

	common.Infof("Found %d CFS sources", len(cfsSourceList))
	return true
}
