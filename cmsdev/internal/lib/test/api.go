package test

/*
 * api.go
 *
 * cms api test helper functions
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
	"github.com/go-resty/resty"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

func GetAccessToken() string {
	common.Infof("Getting access token")
	token, err := k8s.GetAccessToken()
	if err != nil {
		common.Error(fmt.Errorf("Error getting access token: %v", err))
		return ""
	}
	return token
}

func GetAccessTokenParams() *common.Params {
	token := GetAccessToken()
	if len(token) == 0 {
		return nil
	}
	return &common.Params{Token: token}
}

func RestfulTestHeader(label string, testNum, testTotal int) {
	common.VerbosePrintDivider()
	common.Infof("%s test scenario #%d/%d", label, testNum, testTotal)
}

func RestfulVerifyStatus(method, url string, params common.Params, ExpectedStatus int) (resp *resty.Response, err error) {
	common.Infof("%s %s", method, url)
	resp, err = common.Restful(method, url, params)
	if err != nil {
		err = fmt.Errorf("%s %s failed: %v", method, url, err)
		return
	}
	common.PrettyPrintJSON(resp)
	if resp.StatusCode() != ExpectedStatus {
		err = fmt.Errorf("%s %s: expected status code %d, got %d", method, url, ExpectedStatus, resp.StatusCode())
		return
	}
	common.Infof("Received status code %d, as expected", resp.StatusCode())
	return
}

func RestfulTestResultSummary(numFailed, testTotal int) {
	common.Infof("%d passed, %d failed", testTotal-numFailed, numFailed)
}
