package test

/*
 * cli.go
 *
 * cms CLI test helper functions
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	common.Infof("Getting access JSON object")
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
	common.Infof("Writing credentials for CLI authentication to file: %s", CliAuthFile)
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
	common.Infof("Writing credentials for CLI config to file: %s", tmpFile)
	err := ioutil.WriteFile(tmpFile, cli_config_file_text, 0644)
	if err != nil {
		common.Error(err)
		return ""
	}
	return tmpFile
}

func RunCLICommand(cmd string) []byte {
	var cmdStr, tmpCliConfigFile string

	accessFile := GetAccessFile()
	baseCmdStr := fmt.Sprintf("CRAY_CREDENTIALS=%s %s %s", accessFile, cray_cli, cmd)
	if len(CliConfigFile) == 0 {
		var stdout, stderr bytes.Buffer
		var rc int
		cmd := exec.Command("bash", "-c", baseCmdStr)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		common.Infof("Running command: %s", baseCmdStr)
		err := cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				rc = exitError.ExitCode()
			}
		} else {
			rc = 0
		}
		outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		common.Infof("CLI command return code: %d", rc)
		common.Infof("CLI command stdout:\n%s", outStr)
		common.Infof("CLI command stderr:\n%s", errStr)
		if rc == 0 {
			return stdout.Bytes()
		}
		configError := false
		for _, cstr := range ConfigErrorStrings {
			if strings.Contains(errStr, cstr) {
				configError = true
				break
			}
		}
		if !configError {
			common.Error(err)
			common.Infof("CLI error does not look like a CLI config issue")
			return nil
		}
		common.Infof("CLI command failure looks like it may be a CLI config issue")
		common.Infof("Will generate a config file and retry")

		tmpCliConfigFile = MakeConfigFile()
		cmdStr = "CRAY_CONFIG=" + tmpCliConfigFile + " " + baseCmdStr
	} else {
		cmdStr = "CRAY_CONFIG=" + CliConfigFile + " " + baseCmdStr
	}
	common.Infof("Running command: %s", cmdStr)
	cmdOut, err := exec.Command("bash", "-c", cmdStr).Output()
	if err != nil {
		common.Error(err)
		return nil
	} else {
		if len(tmpCliConfigFile) > 0 {
			// Remember the CLI config file for future CLI calls
			CliConfigFile = tmpCliConfigFile
		}
		return cmdOut
	}
}
