//
//  MIT License
//
//  (C) Copyright 2021-2022, 2024-2025 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//
/*
 * cmd.go
 *
 * Command execution helper functions
 *
 */

package common

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var CommandPaths = map[string]string{}

const CmdRcCannotGet = -1
const CLI_TIMEOUT_SECONDS = 3 * time.Second // Timeout for CLI calls setting it to 2 minutes

// Ran is set to true if the Run command was called on the
// Cmd object. It does not mean that the command itself actually
// was run necessarily.
// Environment variables in CmdEnv (if any) are APPENDED to the
// current environment variables, not used in place of them
type CommandResult struct {
	CmdPath, CmdString string
	CmdArgs            []string
	CmdEnv             map[string]string
	CmdErr             error
	ExecCmd            *exec.Cmd
	Rc                 int
	OutBytes, ErrBytes []byte
	Ran                bool
}

func (cmdResult *CommandResult) Init(cmdEnv map[string]string, cmdPath string, cmdArgs ...string) error {
	if len(cmdPath) == 0 {
		Debugf("CommandResult Init(): cmdArgs = %v", cmdArgs)
		return fmt.Errorf("CommandResult Init(): cmdPath may not be empty")
	}
	cmdResult.CmdPath = cmdPath
	cmdResult.CmdArgs = cmdArgs
	cmdResult.CmdEnv = cmdEnv
	for envVarName := range cmdEnv {
		if len(envVarName) == 0 {
			return fmt.Errorf("CommandResult Init(): cmdEnv may not contain blank environment variable names")
		}
	}
	return nil
}

func (cmdResult *CommandResult) OutString() string {
	return string(cmdResult.OutBytes)
}

func (cmdResult *CommandResult) ErrString() string {
	return string(cmdResult.ErrBytes)
}

func (cmdResult *CommandResult) SetEnvVars() string {
	if len(cmdResult.CmdEnv) == 0 {
		return ""
	}

	var envVarNames strings.Builder
	cmdResult.ExecCmd.Env = os.Environ()
	for envVarName, envVarValue := range cmdResult.CmdEnv {
		if envVarNames.Len() > 0 {
			envVarNames.WriteString(" ")
		}
		envVarNames.WriteString(envVarName)
		cmdResult.ExecCmd.Env = append(cmdResult.ExecCmd.Env, fmt.Sprintf("%s=%s", envVarName, envVarValue))
	}
	return envVarNames.String()
}

// The command returning non-0 does NOT constitute an error -- that
// is communicated back via the command return code, and the calling
// function is responsible for determining how to handle that
func (cmdResult *CommandResult) Run() (err error) {
	var stdout, stderr bytes.Buffer

	// Create a context for CLI command.
	ctx, cancel := context.WithTimeout(context.Background(), CLI_TIMEOUT_SECONDS)
	defer cancel()

	cmdResult.ExecCmd = exec.CommandContext(ctx, cmdResult.CmdPath, cmdResult.CmdArgs...)
	cmdResult.CmdString = fmt.Sprintf("%s", cmdResult.ExecCmd)
	envVarNames := cmdResult.SetEnvVars()
	if len(envVarNames) > 0 {
		Debugf("Running command: %s", cmdResult.CmdString)
		Debugf("The following additional environment variables are set for the command: %s", envVarNames)
	} else {
		Debugf("Running command with no additional environment variables set: %s", cmdResult.CmdString)
	}
	cmdResult.ExecCmd.Stdout = &stdout
	cmdResult.ExecCmd.Stderr = &stderr

	cmdResult.CmdErr = cmdResult.ExecCmd.Run()
	cmdResult.Ran = true

	// Check for timeout first
	if ctx.Err() == context.DeadlineExceeded {
		cmdResult.Rc = CmdRcCannotGet
		Error(fmt.Errorf("CLI command timed out"))
		err = fmt.Errorf("CLI command timed out")
	} else if cmdResult.CmdErr != nil {
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
func RunPathWithEnv(cmdEnv map[string]string, cmdPath string, cmdArgs ...string) (cmdResult *CommandResult, err error) {
	cmdResult = new(CommandResult)
	err = cmdResult.Init(cmdEnv, cmdPath, cmdArgs...)
	if err != nil {
		return
	}
	err = cmdResult.Run()
	return
}

// Wrapper for RunPathWithEnv with no environment variables
func RunPath(cmdPath string, cmdArgs ...string) (*CommandResult, error) {
	return RunPathWithEnv(nil, cmdPath, cmdArgs...)
}

// The command returning non-0 does NOT constitute an error -- that
// is communicated back via the command return code, and the calling
// function is responsible for determining how to handle that
func RunNameWithEnv(cmdEnv map[string]string, cmdName string, cmdArgs ...string) (cmdResult *CommandResult, err error) {
	cmdResult = new(CommandResult)
	cmdPath, err := GetPath(cmdName)
	if err != nil {
		return
	}
	cmdResult, err = RunPathWithEnv(cmdEnv, cmdPath, cmdArgs...)
	return
}

// Wrapper for RunNameWithEnv with no environment variables
func RunName(cmdName string, cmdArgs ...string) (*CommandResult, error) {
	return RunNameWithEnv(nil, cmdName, cmdArgs...)
}
