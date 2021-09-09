package test

/*
 * cli.go
 *
 * cms CLI test helper functions
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 *
 * (MIT License)
 */

import (
	"fmt"
	"io/ioutil"
	"os"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
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

func MakeConfigFile() string {
	pid := os.Getpid()
	tmpFile := fmt.Sprintf("/tmp/cmsdev-cray-cli-config-file.%d.tmp", pid)
	common.Debugf("Writing credentials for CLI config to file: %s", tmpFile)
	err := ioutil.WriteFile(tmpFile, cli_config_file_text, 0644)
	if err != nil {
		common.Error(err)
		return ""
	}
	return tmpFile
}

func RunCLICommandJSON(baseCmdString string, cmdArgs ...string) []byte {
	cmdList := append([]string{baseCmdString}, cmdArgs...)
	cmdList = append(cmdList, "--format", "json")
	return RunCLICommand(cmdList...)
}

func RunCLICommand(cmdList ...string) []byte {
	var cmdResult *common.CommandResult
	var cmdStr, tmpCliConfigFile string
	var err error

	accessFile := GetAccessFile()
	baseCmdStr := fmt.Sprintf("CRAY_CREDENTIALS=%s '%s'", accessFile, cray_cli)
	for _, cliArg := range cmdList {
		baseCmdStr = fmt.Sprintf("%s '%s'", baseCmdStr, cliArg)
	}
	if len(CliConfigFile) == 0 {
		cmdResult, err = common.RunName("bash", "-c", baseCmdStr)
		if err != nil {
			common.Error(err)
			common.Errorf("Error running CLI command")
			return nil
		} else if cmdResult.Rc == 0 {
			return cmdResult.OutBytes
		}
		configError := false
		errStr := cmdResult.ErrString()
		for _, cstr := range ConfigErrorStrings {
			if strings.Contains(errStr, cstr) {
				configError = true
				break
			}
		}
		if !configError {
			common.Errorf("CLI command failed (and does not look like a CLI config issue)")
			return nil
		}
		common.Debugf("CLI command failure looks like it may be a CLI config issue")
		common.Debugf("Will generate a config file and retry")

		tmpCliConfigFile = MakeConfigFile()
		cmdStr = "CRAY_CONFIG=" + tmpCliConfigFile + " " + baseCmdStr
	} else {
		cmdStr = "CRAY_CONFIG=" + CliConfigFile + " " + baseCmdStr
	}
	common.Debugf("Running command: %s", cmdStr)
	cmdResult, err = common.RunName("bash", "-c", cmdStr)
	if err != nil {
		common.Error(err)
		common.Errorf("Error running CLI command")
		return nil
	} else if cmdResult.Rc != 0 {
		common.Errorf("CLI command failed")
		return nil
	}
	if len(tmpCliConfigFile) > 0 {
		// Remember the CLI config file for future CLI calls
		CliConfigFile = tmpCliConfigFile
	}
	return cmdResult.OutBytes
}
