// MIT License
//
// (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP
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
package bos

/*
 * bos_cli.go
 *
 * bos CLI helpers
 *
 */

import (
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

// Wrapper function for test.RunCLICommandJSON. Prepends "bos" to a CLI command list and then runs it.
func runBosCLI(cmdArgs ...string) []byte {
	common.Infof("Testing BOS CLI (%s)", strings.Join(cmdArgs, " "))
	return test.RunCLICommandJSON("bos", cmdArgs...)
}

// Same but for TenantRunCLICommandJSON
// If tenant is empty string it means no tenant
func runTenantBosCLI(tenant string, cmdArgs ...string) []byte {
	if len(tenant) == 0 {
		return runBosCLI(cmdArgs...)
	}
	common.Infof("Testing BOS CLI (%s) on behalf of tenant '%s'", strings.Join(cmdArgs, " "), tenant)
	return test.TenantRunCLICommandJSON(tenant, "bos", cmdArgs...)
}

// Wrapper function for runBosCLI. Append "list" to a CLI command list and then runs it.
func runBosCLIList(cmdArgs ...string) []byte {
	newCmdArgs := append(cmdArgs, "list")
	return runBosCLI(newCmdArgs...)
}

// Same but for runTenantBosCLI
// If tenant is empty string it means no tenant
func runTenantBosCLIList(tenant string, cmdArgs ...string) []byte {
	newCmdArgs := append(cmdArgs, "list")
	return runTenantBosCLI(tenant, newCmdArgs...)
}

// Wrapper function for runBosCLI. Append "describe <target>" to a CLI command list and then runs it.
func runBosCLIDescribe(target string, cmdArgs ...string) []byte {
	newCmdArgs := append(cmdArgs, "describe", target)
	return runBosCLI(newCmdArgs...)
}

// Same but for runTenantBosCLI
// If tenant is empty string it means no tenant
func runTenantBosCLIDescribe(tenant, target string, cmdArgs ...string) []byte {
	newCmdArgs := append(cmdArgs, "describe", target)
	return runTenantBosCLI(tenant, newCmdArgs...)
}

// Runs a BOS CLI list with the specified arguments
// If tenant is empty string it means no tenant
// Parses the result to convert it to a list of dictionaries with string keys
// Returns the result and a boolean indicating whether or not this was successful (true == no errors)
func bosTenantListCli(tenant string, cmdArgs ...string) (dictList []map[string]interface{}, passed bool) {
	var err error

	passed = false
	cmdOut := runTenantBosCLIList(tenant, cmdArgs...)
	if cmdOut == nil {
		return
	}

	// Decode JSON into a list of string maps
	dictList, err = common.DecodeJSONIntoStringMapList(cmdOut)
	if err != nil {
		common.Error(err)
		return
	}
	passed = true
	return
}

func bosListCli(cmdArgs ...string) (dictList []map[string]interface{}, passed bool) {
	return bosTenantListCli("", cmdArgs...)
}

// Runs a BOS CLI describe on the specified target with the specified arguments
// If tenant is empty string it means no tenant
// Parses the result to convert it to a dictionary with string keys
// Returns the result and a boolean indicating whether or not this was successful (true == no errors)
func bosTenantDescribeCli(tenant, target string, cmdArgs ...string) (dict map[string]interface{}, passed bool) {
	var err error

	passed = false
	cmdOut := runTenantBosCLIDescribe(tenant, target, cmdArgs...)
	if cmdOut == nil {
		return
	}

	// Decode JSON into a list of string maps
	dict, err = common.DecodeJSONIntoStringMap(cmdOut)
	if err != nil {
		common.Error(err)
		return
	}
	passed = true
	return
}

// Given a BOS CLI command prefix, run that CLI command with "list" appended to the end.
// Verify that the command succeeded and returns a dictionary (aka string map) object.
// Return true if all of that worked fine. Otherwise, log an appropriate error and return false.
// If tenant is empty string it means no tenant
func basicTenantCLIListVerifyStringMapTest(tenant string, cmdArgs ...string) bool {
	cmdOut := runTenantBosCLIList(tenant, cmdArgs...)
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

func basicCLIListVerifyStringMapTest(cmdArgs ...string) bool {
	return basicTenantCLIListVerifyStringMapTest("", cmdArgs...)
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
func cliTests(tenantList []string, includeTenant bool) (passed bool) {
	passed = true

	// Defined in bos_version.go
	if !versionTestsCLI(tenantList) {
		passed = false
	}

	// Defined in bos_healthz.go
	if !healthzTestsCLI(tenantList) {
		passed = false
	}

	// Defined in bos_components.go
	if !componentsTestsCLI(tenantList) {
		passed = false
	}

	// Defined in bos_options.go
	if !optionsTestsCLI() {
		passed = false
	}

	// Defined in bos_sessiontemplate.go
	if !sessionTemplatesTestsCLI(tenantList, includeTenant) {
		passed = false
	}

	// Defined in bos_session.go
	if !sessionsTestsCLI(tenantList, includeTenant) {
		passed = false
	}

	return
}
