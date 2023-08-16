// MIT License
//
// (C) Copyright 2019-2023 Hewlett Packard Enterprise Development LP
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
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"strings"
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

var cli_config_file_text = []byte(`
[core]
hostname = "https://api-gw-service-nmn.local"
`)

// Paths to different CLI config files, based on which tenant
// we are acting on behalf of. The empty string will map to the CLI config
// file with no tenant specified
var cliConfigFilesByTenant = map[string]string{}

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
	pid := os.Getpid()
	CliAuthFile = fmt.Sprintf("/tmp/cmsdev-cray-credentials-file.%d.json", pid)
	common.Debugf("Writing credentials for CLI authentication to file: %s", CliAuthFile)
	err := ioutil.WriteFile(CliAuthFile, jobj, 0644)
	if err != nil {
		common.Error(err)
		return ""
	}
	return CliAuthFile
}

func MakeConfigFile(tenant string) (filePath string, err error) {
	var cmdResult *common.CommandResult
	var ok bool

	filePath, ok = cliConfigFilesByTenant[tenant]
	if ok {
		return
	}
	filePath = common.TmpDir + "/" + tenant + ".craycli.config"
	baseCmdStr := fmt.Sprintf("'%s' init --hostname api-gw-service-nmn.local --no-auth --configuration '%s'", cray_cli, filePath)
	if len(tenant) == 0 {
		common.Debugf("Creating untenanted config file for Cray CLI: '%s'", filePath)
	} else {
		common.Debugf("Creating config file for tenant '%s' for Cray CLI: '%s'", tenant, filePath)
		baseCmdStr = fmt.Sprintf("%s --tenant '%s'", baseCmdStr, tenant)
	}
	cmdResult, err = common.RunName("bash", "-c", baseCmdStr)
	if err != nil {
		err = fmt.Errorf("Error running CLI command (%s): %v", baseCmdStr, err)
		return
	}
	cliConfigFilesByTenant[tenant] = filePath
	return
}

func RunCLICommandJSON(baseCmdString string, cmdArgs ...string) []byte {
	cmdList := append([]string{baseCmdString}, cmdArgs...)
	cmdList = append(cmdList, "--format", "json")
	return RunCLICommand(cmdList...)
}

func RunCLICommand(cmdList ...string) []byte {
	var cmdResult *common.CommandResult
	var cmdStr string
	var err error

	accessFile := GetAccessFile()
	CliConfigFile, err = MakeConfigFile("")
	if err != nil {
		common.Error(err)
		return nil
	}
	baseCmdStr := fmt.Sprintf("CRAY_CREDENTIALS=%s '%s'", accessFile, cray_cli)
	for _, cliArg := range cmdList {
		baseCmdStr = fmt.Sprintf("%s '%s'", baseCmdStr, cliArg)
	}
	cmdStr = "CRAY_CONFIG=" + CliConfigFile + " " + baseCmdStr
	common.Debugf("Running command: %s", cmdStr)
	cmdResult, err = common.RunName("bash", "-c", cmdStr)
	if err != nil {
		common.Error(err)
		common.Errorf("Error running CLI command (%s)", strings.Join(cmdList, " "))
		return nil
	} else if cmdResult.Rc != 0 {
		common.Errorf("CLI command failed (%s)", strings.Join(cmdList, " "))
		return nil
	}
	return cmdResult.OutBytes
}
