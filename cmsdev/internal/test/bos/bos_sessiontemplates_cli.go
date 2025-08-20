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
 * bos_sessiontemplate_cli.go
 *
 * BOS sessiontemplate CLI helper functions
 *
 */
package bos

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

// BOS cli versions
var bosCliVersions = map[string]string{
	"default": "",
	"v2":      "v2",
}

func RunVersionedBOSCommand(cliVersion string, cmdArgs ...string) []byte {
	if cliVersion == "" {
		return runTenantBosCLI(common.GetTenantName(), cmdArgs...)
	}
	newArgs := append([]string{cliVersion}, cmdArgs...)
	return runTenantBosCLI(common.GetTenantName(), newArgs...)
}

func GetCreateBOSSessionTemplatePayloadCLI(cfsConfigName string, enableCFS bool, arch string, imageId string) (fileName string, payload string, ok bool) {
	fileName = "bos_sessiontemplate_create_payload"
	payload, success := GetCreateBOSSessionTemplatePayload(cfsConfigName, enableCFS, arch, imageId)
	if !success {
		return "", "", false
	}

	dir, err := os.Getwd()
	if err != nil {
		common.Errorf("Error getting current directory: %v\n", err)
		return
	}
	fileName = fmt.Sprintf("%s/%s_%s.json", dir, fileName, arch)

	// Write the formatted JSON payload to the file
	err = os.WriteFile(fileName, []byte(payload), 0644)
	if err != nil {
		common.Errorf("Unable to write payload to file %s: %v", fileName, err)
		return "", "", false
	}

	return fileName, payload, true
}

func UpdateBOSSessionTemplatesCLI(templateName, filename, cfgName, cliVersion string) (sessionTemplateRecord BOSSessionTemplate, passed bool) {
	common.Infof("Updating session template %s in BOS via CLI", templateName)
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessiontemplates", "update", templateName,
		"--file", filename); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &sessionTemplateRecord); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func CreateBOSSessionTemplatesCLI(templateName, filename, cfgName, cliVersion string) (sessionTemplateRecord BOSSessionTemplate, passed bool) {
	common.Infof("Creating session template %s in BOS via CLI", templateName)
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessiontemplates", "create", templateName,
		"--file", filename); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &sessionTemplateRecord); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	// if the tenant is a dummy tenant, we expect the command to fail
	if IsDummyTenant(common.GetTenantName()) {
		return BOSSessionTemplate{}, true
	}
	return
}

func GetBOSSessiontemplatesCLI(templateName, cliVersion string) (sessionTemplateRecord BOSSessionTemplate, passed bool) {
	common.Infof("Getting BOS session template %s via CLI", templateName)
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessiontemplates", "describe", templateName); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &sessionTemplateRecord); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetBOSSessiontemplatesListCLI(cliVersion string) (sessionTemplateRecords []BOSSessionTemplate, passed bool) {
	common.Infof("Getting all BOS session templates via CLI")
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessiontemplates", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &sessionTemplateRecords); err == nil {
			passed = true
		} else {
			common.Error(err)
		}
	}
	return
}

func DeleteBOSSessionTemplatesCLI(templateName, cliVersion string) (passed bool) {
	common.Infof("Deleting session template %s in BOS via CLI", templateName)
	return RunVersionedBOSCommand(cliVersion, "sessiontemplates", "delete", templateName) != nil
}

func ValidateBOSSessionTemplateCLI(templateName, cliVersion string) (passed bool) {
	common.Infof("Validating BOS session template %s via CLI", templateName)
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessiontemplatesvalid", "describe", templateName); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		output := string(cmdOut)
		if strings.Contains(output, "Valid") {
			common.Infof("Session template %s is valid", templateName)
			passed = true
		} else {
			common.Errorf("Session template %s is not valid", templateName)
			passed = false
		}
	}
	return
}
