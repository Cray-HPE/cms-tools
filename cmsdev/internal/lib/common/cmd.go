/*
 * cmd.go
 *
 * Command execution helper functions
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
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

package common

import (
	"bytes"
	"fmt"
	"os/exec"
)

var CommandPaths = map[string]string{}

const CmdRcCannotGet = -1

// Ran is set to true if the Run command was called on the
// Cmd object. It does not mean that the command itself actually
// was run necessarily
type CommandResult struct {
	CmdPath, CmdString string
	CmdArgs            []string
	CmdErr             error
	ExecCmd            *exec.Cmd
	Rc                 int
	OutBytes, ErrBytes []byte
	Ran                bool
}

func (cmdResult *CommandResult) Init(cmdPath string, cmdArgs ...string) error {
	if len(cmdPath) == 0 {
		Debugf("CommandResult init(): cmdArgs = %v", cmdArgs)
		return fmt.Errorf("CommandResult init(): cmdPath may not be empty")
	}
	cmdResult.CmdPath = cmdPath
	cmdResult.CmdArgs = cmdArgs
	return nil
}

func (cmdResult *CommandResult) OutString() string {
	return string(cmdResult.OutBytes)
}

func (cmdResult *CommandResult) ErrString() string {
	return string(cmdResult.ErrBytes)
}

// The command returning non-0 does NOT constitute an error -- that
// is communicated back via the command return code, and the calling
// function is responsible for determining how to handle that
func (cmdResult *CommandResult) Run() (err error) {
	var stdout, stderr bytes.Buffer

	cmdResult.ExecCmd = exec.Command(cmdResult.CmdPath, cmdResult.CmdArgs...)
	cmdResult.CmdString = fmt.Sprintf("%s", cmdResult.ExecCmd)
	Debugf("Running command: %s", cmdResult.CmdString)

	cmdResult.ExecCmd.Stdout = &stdout
	cmdResult.ExecCmd.Stderr = &stderr

	cmdResult.CmdErr = cmdResult.ExecCmd.Run()
	cmdResult.Ran = true
	if cmdResult.CmdErr != nil {
		if exitError, ok := cmdResult.CmdErr.(*exec.ExitError); ok {
			cmdResult.Rc = exitError.ExitCode()
		} else {
			cmdResult.Rc = CmdRcCannotGet
			Error(cmdResult.CmdErr)
			err = fmt.Errorf("Unable to determine command return code")
		}
	} else {
		cmdResult.Rc = 0
	}
	cmdResult.OutBytes, cmdResult.ErrBytes = stdout.Bytes(), stderr.Bytes()
	if cmdResult.Rc != CmdRcCannotGet {
		Debugf("Command return code: %d", cmdResult.Rc)
	}
	if len(cmdResult.OutString()) > 0 {
		Debugf("Command stdout:\n%s", cmdResult.OutString())
	} else {
		Debugf("No stdout from command")
	}
	if len(cmdResult.ErrString()) > 0 {
		Debugf("Command stderr:\n%s", cmdResult.ErrString())
	} else {
		Debugf("No stderr from command")
	}
	return
}

// Looks up the path of the specified command
// Logs errors, if any
func GetPath(cmdName string) (path string, err error) {
	path, ok := CommandPaths[cmdName]
	if ok {
		Debugf("Using cached value of %s path: %s", cmdName, path)
		return
	}
	Debugf("Looking up path of %s", cmdName)
	path, err = exec.LookPath(cmdName)
	if err != nil {
		return
	} else if len(path) == 0 {
		err = fmt.Errorf("Empty path found for %s", cmdName)
		return
	}
	Debugf("Found path of %s: %s", cmdName, path)
	CommandPaths[cmdName] = path
	return
}

// The command returning non-0 does NOT constitute an error -- that
// is communicated back via the command return code, and the calling
// function is responsible for determining how to handle that
func RunPath(cmdPath string, cmdArgs ...string) (cmdResult *CommandResult, err error) {
	cmdResult = new(CommandResult)
	err = cmdResult.Init(cmdPath, cmdArgs...)
	if err != nil {
		return
	}
	err = cmdResult.Run()
	return
}

// The command returning non-0 does NOT constitute an error -- that
// is communicated back via the command return code, and the calling
// function is responsible for determining how to handle that
func RunName(cmdName string, cmdArgs ...string) (cmdResult *CommandResult, err error) {
	cmdResult = new(CommandResult)
	cmdPath, err := GetPath(cmdName)
	if err != nil {
		return
	}
	cmdResult, err = RunPath(cmdPath, cmdArgs...)
	return
}
