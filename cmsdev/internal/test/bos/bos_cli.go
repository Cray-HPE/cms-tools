//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
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
package bos

/*
 * bos_cli.go
 *
 * bos CLI helpers
 *
 */

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

// Wrapper function for test.RunCLICommandJSON. Prepends "bos" to a CLI command list and then runs it.
func runBosCLI(cmdArgs ...string) []byte {
	common.Infof("Testing BOS CLI (%s)", strings.Join(cmdArgs, " "))
	return test.RunCLICommandJSON("bos", cmdArgs...)
}

// Wrapper function for runBosCLI. Append "list" to a CLI command list and then runs it.
func runBosCLIList(cmdArgs ...string) []byte {
	newCmdArgs := append(cmdArgs, "list")
	return runBosCLI(newCmdArgs...)
}

// Wrapper function for runBosCLI. Append "describe <target>" to a CLI command list and then runs it.
func runBosCLIDescribe(target string, cmdArgs ...string) []byte {
	newCmdArgs := append(cmdArgs, "describe", target)
	return runBosCLI(newCmdArgs...)
}

// Given a BOS CLI command prefix, run that CLI command with "list" appended to the end.
// Verify that the command succeeded and returns a dictionary (aka string map) object.
// Return true if all of that worked fine. Otherwise, log an appropriate error and return false.
func basicCLIListVerifyStringMapTest(cmdArgs ...string) bool {
	cmdOut := runBosCLIList(cmdArgs...)
	if cmdOut == nil {
		return false
	}

	// Validate that object can be decoded into a string map at least
	_, err := common.DecodeJSONIntoStringMap(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Given a BOS CLI command prefix and a target name, run that CLI command with "describe <target>" appended to the end.
// Verify that the command succeeded and returns a dictionary (aka string map) object.
// Return true if all of that worked fine. Otherwise, log an appropriate error and return false.
func basicCLIDescribeVerifyStringMapTest(target string, cmdArgs ...string) bool {
	cmdOut := runBosCLIDescribe(target, cmdArgs...)
	if cmdOut == nil {
		return false
	}

	// Validate that object can be decoded into a string map at least
	_, err := common.DecodeJSONIntoStringMap(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Run all of the BOS CLI subtests. Return true if they all pass, false otherwise.
func cliTests() (passed bool) {
	passed = true

	// Defined in bos_version.go
	if !versionTestsCLI() {
		passed = false
	}

	// Defined in bos_healthz.go
	if !healthzTestsCLI() {
		passed = false
	}

	// Defined in bos_components.go
	if !componentsTestsCLI() {
		passed = false
	}

	// Defined in bos_options.go
	if !optionsTestsCLI() {
		passed = false
	}

	// Defined in bos_sessiontemplate.go
	if !sessionTemplatesTestsCLI() {
		passed = false
	}

	// Defined in bos_session.go
	if !sessionsTestsCLI() {
		passed = false
	}

	return
}
