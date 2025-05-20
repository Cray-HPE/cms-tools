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
 * cfs_sources_cli.go
 *
 * cfs sources cli functions
 *
 */
package cfs

import (
	"encoding/json"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func CreateCFSSourceRecordCLI(sourceName, cloneURL, cliVersion string) (cfsSourceRecord CFSSources, passed bool) {
	common.Infof("Creating source %s in CFS via CLI", sourceName)
	if cmdOut := RunVersionedCFSCommand(cliVersion, "sources", "create", "--name", sourceName,
		"--clone-url", cloneURL, "--credentials-username", "user", "--credentials-password", "pass"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsSourceRecord); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetCFSSourceRecordCLI(sourceName, cliVersion string) (cfsSourceRecord CFSSources, passed bool) {
	common.Infof("Getting source %s in CFS via CLI", sourceName)
	if cmdOut := RunVersionedCFSCommand(cliVersion, "sources", "describe", sourceName); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsSourceRecord); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetCFSSourcesListCLI(cliVersion string) (cfsSourceRecords []CFSSources, passed bool) {
	var cfsSourcesList CFSSourcesList
	common.Infof("Getting all source in CFS via CLI")
	if cmdOut := RunVersionedCFSCommand(cliVersion, "sources", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsSourcesList); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return cfsSourcesList.Sources, passed
}

func UpdateCFSSourceRecordCLI(sourceName, cloneURL, cliVersion string) (cfsSourceRecord CFSSources, passed bool) {
	common.Infof("Updating source %s in CFS via CLI", sourceName)
	if cmdOut := RunVersionedCFSCommand(cliVersion, "sources", "update", sourceName,
		"--clone-url", cloneURL); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsSourceRecord); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func DeleteCFSSourceRecordCLI(sourceName, cliVersion string) (passed bool) {
	common.Infof("Deleting source %s in CFS via CLI", sourceName)
	return RunVersionedCFSCommand(cliVersion, "sources", "delete", sourceName) != nil
}
