package test

/*
 * api.go
 *
 * cms api test helper functions
 *
 * Copyright 2019-2020, Cray Inc.
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
