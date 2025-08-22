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
 * verify_cfs_sources.go
 *
 * cfs sources test api functions
 *
 */
package cfs

import (
	"fmt"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func TestCFSSourcesCRUDOperation() (passed bool) {
	passed = true

	cfsSourceRecord, success := TestCFSSourceCreate()
	if !success {
		return false
	}

	if len(cfsSourceRecord.Name) != 0 {
		updated := TestCFSSourceUpdate(cfsSourceRecord.Name)

		deleted := TestCFSSourceDelete(cfsSourceRecord.Name)

		getAll := TestCFSSourceGetAll()

		passed = passed && updated && deleted && getAll
	}

	return passed
}

func TestCFSSourceCreate() (cfsSourceRecord CFSSources, passed bool) {
	sourceName := "CFS_Source_" + string(common.GetRandomString(10))
	common.PrintLog(fmt.Sprintf("Creating CFS source: %s", sourceName))

	// Create a new CFS source record
	cfsSourceRecord, success := CreateCFSSourceRecordAPI(sourceName)
	if !success {
		return CFSSources{}, false
	}

	// Get the CFS source record
	_, success = GetCFSSourceRecordAPI(sourceName, http.StatusOK)
	if !success {
		common.Errorf("Failed to get CFS source record: %s", sourceName)
		return CFSSources{}, false
	}

	// verify CFS source is in the list of sources
	cfsSourceList, success := GetCFSSourcesListAPI()
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

func TestCFSSourceUpdate(sourceName string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Updating CFS source: %s", sourceName))

	// Update the CFS source record
	cfsSourceRecord, success := UpdateCFSSourceRecordAPI(sourceName)
	if !success {
		return false
	}

	// Get the CFS source record
	_, success = GetCFSSourceRecordAPI(sourceName, http.StatusOK)
	if !success {
		common.Errorf("Failed to get CFS source record: %s", sourceName)
		return false
	}

	// verify CFS source is in the list of sources
	cfsSourceList, success := GetCFSSourcesListAPI()
	if !success {
		common.Errorf("Failed to get CFS sources list")
		return false
	}

	// Check if the source is in the list
	if !CFSSourceExists(cfsSourceList, sourceName) {
		return false
	}

	// Verify the CFS source record
	if !VerifyCFSSourceRecord(cfsSourceRecord, sourceName, secondCloneURL) {
		return false
	}

	common.Infof("CFS source %s updated successfully", sourceName)
	return true
}

func TestCFSSourceDelete(sourceName string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting CFS source: %s", sourceName))

	// Delete the CFS source record
	success := DeleteCFSSourceRecordAPI(sourceName)
	if !success {
		return false
	}

	// Verify the CFS source record is deleted
	_, success = GetCFSSourceRecordAPI(sourceName, http.StatusNotFound)
	if !success {
		return false
	}

	// verify CFS source is not in the list of sources
	cfsSourceList, success := GetCFSSourcesListAPI()
	if !success {
		return false
	}

	// Check if the source is in the list
	if CFSSourceExists(cfsSourceList, sourceName) {
		return false
	}

	common.Infof("CFS source %s deleted successfully", sourceName)
	return true
}

func TestCFSSourceGetAll() (passed bool) {
	common.PrintLog("Getting all CFS sources")
	// Get CFS sources list
	cfsSources, success := GetCFSSourcesListAPI()
	if !success {
		return false
	}

	for _, source := range cfsSources {
		common.Infof("CFS source: %s", source.Name)
	}

	common.Infof("Found %d CFS sources", len(cfsSources))
	return true
}
