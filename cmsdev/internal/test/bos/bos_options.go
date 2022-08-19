//
//  MIT License
//
//  (C) Copyright 2022 Hewlett Packard Enterprise Development LP
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
 * bos_options.go
 *
 * bos options tests
 *
 */

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

// Options are new in BOS v2
const bosV2OptionsUri = bosV2BaseUri + "/options"

const bosV2OptionsCLI = "options"
const bosDefaultOptionsCLI = bosV2OptionsCLI

func optionsTestsAPI(params *common.Params) (passed bool) {
	passed = true

	if !optionsTestURI(bosV2OptionsUri, params) {
		passed = false
	}

	return
}

func optionsTestsCLI() (passed bool) {
	passed = true

	// v2
	if !optionsTestCLICommand("v2", bosV2OptionsCLI) {
		passed = false
	}

	// default (v2)
	if !optionsTestCLICommand(bosDefaultOptionsCLI) {
		passed = false
	}

	return
}

func optionsTestURI(uri string, params *common.Params) bool {
	return basicGetUriVerifyStringMapTest(uri, params)
}

func optionsTestCLICommand(cmdArgs ...string) bool {
	return basicCLIListVerifyStringMapTest(cmdArgs...)
}
