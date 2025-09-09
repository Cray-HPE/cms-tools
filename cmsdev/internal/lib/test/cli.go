// MIT License
//
// (C) Copyright 2019-2025 Hewlett Packard Enterprise Development LP
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
package test

/*
 * cli.go
 *
 * cms CLI test helper functions
 *
 */

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
)

const cray_cli = "/usr/bin/cray"

var CliAuthFile = ""
var CliConfigFile = ""

var ConfigErrorStrings = []string{
	"Unable to connect to cray",
	"verify your cray hostname",
	"core.hostname",
	"No configuration exists",
	"cray init",
}

var cli_config_file_text = string(`
[core]
hostname = "https://api-gw-service-nmn.local"
tenant = "%s"
`)

// Paths to different CLI config files, based on which tenant
// we are acting on behalf of. The empty string will map to the CLI config
// file with no tenant specified
var cliConfigFilesByTenant = map[string]string{}

// Using a global variable to track the return code of the CLI command execution
// This is used to simulate different return codes for testing purposes
// The default value is 0, which indicates success
var cliExecreturnCode = 0

func SetCliExecreturnCode(code int) {
	cliExecreturnCode = code
}

func GetCliExecreturnCode() int {
	return cliExecreturnCode
}

func GetAccessJSON() []byte {
	common.Debugf("Getting access JSON object")
	jobj, err := k8s.GetAccessJSON()
	if err != nil {
		common.Error(err)
		return nil
	}
	return jobj
}

func GetAccessFile() string {
	if CliAuthFile != "" {
		return CliAuthFile
	}
	jobj := GetAccessJSON()
	if jobj == nil {
		return ""
	}
	CliAuthFile = common.TmpDir + "/cmsdev-cray-credentials-file.json"
	common.Debugf("Writing credentials for CLI authentication to file: %s", CliAuthFile)
	err := ioutil.WriteFile(CliAuthFile, jobj, 0644)
	if err != nil {
		common.Error(err)
		return ""
	}
	return CliAuthFile
}

func MakeConfigFile(tenant string) (filePath string, err error) {
	var file_contents string
	var ok bool

	filePath, ok = cliConfigFilesByTenant[tenant]
	if ok {
		return
	}
	filePath = common.TmpDir + "/" + tenant + ".craycli.config"
	if len(tenant) == 0 {
		common.Debugf("Creating untenanted config file for Cray CLI: '%s'", filePath)
	} else {
		common.Debugf("Creating config file for tenant '%s' for Cray CLI: '%s'", tenant, filePath)
	}
	file_contents = fmt.Sprintf(cli_config_file_text, tenant)
	err = os.WriteFile(filePath, []byte(file_contents), 0600)
	if err != nil {
		err = fmt.Errorf("Error writing CLI configuration file '%s': %v", filePath, err)
	} else {
		cliConfigFilesByTenant[tenant] = filePath
	}
	return
}

func RunCLICommandJSON(baseCmdString string, cmdArgs ...string) []byte {
	return TenantRunCLICommandJSON("", baseCmdString, cmdArgs...)
}

func TenantRunCLICommandJSON(tenant, baseCmdString string, cmdArgs ...string) []byte {
	cmdList := append([]string{baseCmdString}, cmdArgs...)
	cmdList = append(cmdList, "--format", "json")
	return TenantRunCLICommand(tenant, cmdList...)
}

func RunCLICommand(cmdList ...string) []byte {
	return TenantRunCLICommand("", cmdList...)
}

func TenantRunCLICommand(tenant string, cmdList ...string) []byte {
	var cmdResult *common.CommandResult
	var cmdStr, tenantText string
	var err error

	if len(tenant) > 0 {
		tenantText = fmt.Sprintf(" on behalf of tenant '%s'", tenant)
	}

	accessFile := GetAccessFile()
	CliConfigFile, err = MakeConfigFile(tenant)
	if err != nil {
		common.Error(err)
		return nil
	}
	baseCmdStr := fmt.Sprintf("CRAY_CREDENTIALS=%s '%s'", accessFile, cray_cli)
	for _, cliArg := range cmdList {
		baseCmdStr = fmt.Sprintf("%s '%s'", baseCmdStr, cliArg)
	}
	cmdStr = "CRAY_CONFIG=" + CliConfigFile + " " + baseCmdStr
	common.Debugf("Running command%s: %s", tenantText, cmdStr)
	cmdResult, err = common.RunNameWithRetry("bash", "-c", cmdStr)
	if err != nil {
		common.Error(err)
		common.Errorf("Error running CLI command%s (%s)", strings.Join(cmdList, " "), tenantText)
		return nil
	} else if cmdResult.Rc != GetCliExecreturnCode() {
		common.Errorf("CLI command%s (%s) failed with exit code %d", tenantText, strings.Join(cmdList, " "), cmdResult.Rc)
		return nil
	}
	// Check for error code, return nil if there is an error
	if cmdResult.Rc != 0 {
		return nil
	}
	return cmdResult.OutBytes
}
