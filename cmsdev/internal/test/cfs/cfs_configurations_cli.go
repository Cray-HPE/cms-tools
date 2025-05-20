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
 * cfs_configurations_cli.go
 *
 * cfs configurations cli functions
 *
 */
package cfs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func RunVersionedCFSCommand(cliVersion string, cmdArgs ...string) []byte {
	common.Infof("Testing CFS CLI (%s %s)", cliVersion, strings.Join(cmdArgs, " "))
	if cliVersion == "" {
		return test.RunCLICommandJSON("cfs", cmdArgs...)
	}
	newArgs := append([]string{cliVersion}, cmdArgs...)
	return test.RunCLICommandJSON("cfs", newArgs...)
}

func CreateCFSConfigurationFile(cfgName, cliVersion string) (fileName string, payload string, ok bool) {
	payload, success := GetCreateCFGConfigurationPayload(cliVersion)
	if !success {
		return "", "", false
	}

	dir, err := os.Getwd()
	if err != nil {
		common.Errorf("Error getting current directory: %v\n", err)
		return
	}
	fileName = fmt.Sprintf("%s/%s.json", dir, fileName)

	// Write the formatted JSON payload to the file
	err = os.WriteFile(fileName, []byte(payload), 0644)
	if err != nil {
		common.Errorf("Unable to write payload to file %s: %v", fileName, err)
		return "", "", false
	}

	return fileName, payload, true
}

func CreateUpdateCFSConfigurationCLI(cfgName, fileName, cliVersion string) (cfsConfig CFSConfiguration, passed bool) {
	common.Infof("Creating configuration %s in CFS via CLI", cfgName)
	if cmdOut := RunVersionedCFSCommand(cliVersion, "configurations", "update", cfgName,
		"--file", fileName); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsConfig); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetCFSConfigurationRecordCLI(cfgName, cliVersion string) (cfsConfig CFSConfiguration, passed bool) {
	common.Infof("Getting configuration %s in CFS via CLI", cfgName)
	if cmdOut := RunVersionedCFSCommand(cliVersion, "configurations", "describe", cfgName); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsConfig); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetCFSConfigurationsListCLI(cliVersion string) (cfsConfigList CFSConfigurationsList, passed bool) {
	common.Infof("Getting CFS configurations list via CLI using cli version %s", cliVersion)
	if cmdOut := RunVersionedCFSCommand(cliVersion, "configurations", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsConfigList); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetCFSConfigurationsListCLIV2(cliVersion string) (cfsConfigList []CFSConfiguration, passed bool) {
	common.Infof("Getting CFS configurations list via CLI using cli version %s", cliVersion)
	if cmdOut := RunVersionedCFSCommand(cliVersion, "configurations", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &cfsConfigList); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetCLIVersionBasedCFSConfigurationRecordList(cliVersion string) (cfsConfigurations []CFSConfiguration, passed bool) {
	// Get the CFS configurations list based on the API version. The response will be different for v2 and v3.
	// For v3, the response will be a list of CFSConfigurationsList, while for v2, it will be a CFSConfigurations struct.
	if cliVersion == "v3" {
		cfsConfigV3, ok := GetCFSConfigurationsListCLI(cliVersion)
		if !ok {
			return []CFSConfiguration{}, false
		}
		cfsConfigurations = cfsConfigV3.Configurations
	} else {
		cfsConfigV2, ok := GetCFSConfigurationsListCLIV2(cliVersion)
		if !ok {
			return []CFSConfiguration{}, false
		}
		cfsConfigurations = cfsConfigV2
	}
	return cfsConfigurations, true
}

func DeleteCFSConfigurationRecordCLI(cfgName, cliVersion string) (passed bool) {
	common.Infof("Deleting configuration %s in CFS via CLI", cfgName)
	return RunVersionedCFSCommand(cliVersion, "configurations", "delete", cfgName) != nil
}
