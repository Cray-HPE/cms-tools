// MIT License
//
// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
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
 * bos_healthz.go
 *
 * bos healthz tests
 *
 */

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

const bosV1HealthzUri = bosV1BaseUri + "/healthz"
const bosV2HealthzUri = bosV2BaseUri + "/healthz"

const bosV1HealthzCLI = "healthz"
const bosV2HealthzCLI = "healthz"
const bosDefaultHealthzCLI = bosV2HealthzCLI

func healthzTestsAPI(params *common.Params) (passed bool) {
	passed = true

	// Just do a GET of the options endpoint and make sure that the response has
	// 200 status and a dictionary object

	// v1 endpoint
	if !basicGetUriVerifyStringMapTest(bosV1HealthzUri, params) {
		passed = false
	}

	// v2 endpoint
	if !basicGetUriVerifyStringMapTest(bosV2HealthzUri, params) {
		passed = false
	}

	return
}

func healthzTestsCLI() (passed bool) {
	passed = true

	// Make sure that "healthz list" CLI command succeeds and returns a dictionary object.

	// "v1 healthz list"
	if !basicCLIListVerifyStringMapTest("v1", bosV1HealthzCLI) {
		passed = false
	}

	// "v2 healthz list"
	if !basicCLIListVerifyStringMapTest("v2", bosV2HealthzCLI) {
		passed = false
	}

	// "healthz list"
	if !basicCLIListVerifyStringMapTest(bosDefaultHealthzCLI) {
		passed = false
	}

	return
}
