package test

/*
 * cli.go
 *
 * cms CLI test helper functions
 *
 * Copyright 2019-2020, Cray Inc.
 */

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

var CliAuthFile = ""

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

func RunCLICommand(cmd string) []byte {
	accessFile := GetAccessFile()
	cmdStr := fmt.Sprintf("CRAY_CREDENTIALS=%s %s", accessFile, cmd)
	common.Infof("Running command: %s", cmdStr)
	cmdOut, err := exec.Command("bash", "-c", cmdStr).Output()
	if err != nil {
		common.Error(err)
		return nil
	} else {
		return cmdOut
	}
}
