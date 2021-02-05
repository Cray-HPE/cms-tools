/*
 * cmd.go
 *
 * Command execution helper functions
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
 */

package common

import (
	"bytes"
	"fmt"
	"os/exec"
)

var CommandPaths = map[string]string{}

const CmdRcCannotGet = -1
const CmdRcDidNotRun = -2

// Looks up the path of the specified command
// Logs errors, if any
// Returns the path and true if successful, empty string and false otherwise
func GetPath(cmdName string) (string, bool) {
	path, ok := CommandPaths[cmdName]
	if ok {
		Infof("Using cached value of %s path: %s", cmdName, path)
		return path, true
	}
	Infof("Looking up path of %s", cmdName)
	path, err := exec.LookPath(cmdName)
	if err != nil {
		Error(err)
		return "", false
	} else if len(path) == 0 {
		Errorf("Empty path found for %s", cmdName)
		return "", false
	}
	Infof("Found path of %s: %s", cmdName, path)
	CommandPaths[cmdName] = path
	return path, true
}

// Wrapper that runs the command and returns the outputs, return code, and error object
func RunPathBytes(cmdPath string, cmdArgs ...string) (cmdOutBytes, cmdErrBytes []byte, cmdRc int, err error) {
	var cmd *exec.Cmd
	var stdout, stderr bytes.Buffer

	cmd = exec.Command(cmdPath, cmdArgs...)
	Infof("Running command: %s", cmd)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmdRc = CmdRcDidNotRun

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			cmdRc = exitError.ExitCode()
		} else {
			cmdRc = CmdRcCannotGet
		}
	} else {
		cmdRc = 0
	}
	cmdOutBytes, cmdErrBytes = stdout.Bytes(), stderr.Bytes()
	cmdOutString, cmdErrString := string(cmdOutBytes), string(cmdErrBytes)
	if err != nil {
		Infof("Error running command: %s", err.Error())
	}
	if cmdRc == CmdRcCannotGet {
		Infof("WARNING: Unable to determine command return code")
	} else {
		Infof("Command return code: %d", cmdRc)
	}
	if len(cmdOutString) > 0 {
		Infof("Command stdout:\n%s", cmdOutString)
	} else {
		Infof("No stdout from command")
	}
	if len(cmdErrString) > 0 {
		Infof("Command stderr:\n%s", cmdErrString)
	} else {
		Infof("No stderr from command")
	}
	return
}

// Wrapper to RunPathBytes that determines the command path for you first
func RunNameBytes(cmdName string, cmdArgs ...string) (cmdOutBytes, cmdErrBytes []byte, cmdRc int, err error) {
	cmdRc = CmdRcDidNotRun
	cmdPath, ok := GetPath(cmdName)
	if !ok {
		err = fmt.Errorf("Cannot determine path of bash binary")
		return
	}
	return RunPathBytes(cmdPath, cmdArgs...)
}

// Wrapper to RunPathBytes that returns the command outputs as strings rather than bytes
func RunPathStrings(cmdPath string, cmdArgs ...string) (cmdOutString, cmdErrString string, cmdRc int, err error) {
	cmdRc = CmdRcDidNotRun
	cmdOutBytes, cmdErrBytes, cmdRc, err := RunPathBytes(cmdPath, cmdArgs...)
	cmdOutString, cmdErrString = string(cmdOutBytes), string(cmdErrBytes)
	return
}

// Wrapper to RunPathStrings that determines the command path for you first
func RunNameStrings(cmdName string, cmdArgs ...string) (cmdOutString, cmdErrString string, cmdRc int, err error) {
	cmdRc = CmdRcDidNotRun
	cmdPath, ok := GetPath(cmdName)
	if !ok {
		err = fmt.Errorf("Cannot determine path of bash binary")
		return
	}
	return RunPathStrings(cmdPath, cmdArgs...)
}
