package cfs

/*
 * cfs.go
 *
 * cfs commons file
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
	"net/http"
	"regexp"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()

var cfsEndpoints = []string{
	"components",
	"configurations",
	"options",
	"sessions",
}

var cfsEndpointIdFieldName = map[string]string{
	"components":     "id",
	"configurations": "name",
	"sessions":       "name",
}

func IsCFSRunning() (passed bool) {
	passed = true
	// 2 pods minimum since we expect both an api and operator pod
	podNames, ok := test.GetPodNamesByPrefixKey("cfs", 2, -1)
	if !ok {
		passed = false
	}
	common.Infof("Found %d cfs pods", len(podNames))
	apiPodName, operatorPodName := "", ""
	for _, podName := range podNames {
		// CFS is running if there is operator and api pod
		// we can ignore the state of cfs-jobs or cfs-db pods
		re := regexp.MustCompile(common.PodServiceNamePrefixes["cfsServices"])
		if re.MatchString(podName) {
			if strings.HasPrefix(podName, common.PodServiceNamePrefixes["cfs-api"]) {
				apiPodName = podName
				common.Infof("Found apiPod=%s", podName)
			} else if strings.HasPrefix(podName, common.PodServiceNamePrefixes["cfs-operator"]) {
				operatorPodName = podName
				common.Infof("Found operatorPod=%s", podName)
			}
		}
		if re.MatchString(podName) {
			common.Infof("checking pod status for %s expecting %s", podName, "Running")
		} else {
			common.Infof("checking pod status for %s expecting %s", podName, "N/A")
		}
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			continue
		}
		common.Infof("Pod status is %s", status)
		if re.MatchString(podName) {
			if status != "Running" {
				common.VerboseFailedf("expected status=Running, found status=%s for podName=%s", status, podName)
				passed = false
			} else {
				common.VerboseOkay()
			}
		}
	}
	if len(apiPodName) == 0 {
		common.Errorf("No apiPod found")
		passed = false
	}
	if len(operatorPodName) == 0 {
		common.Errorf("No operatorPod found")
		passed = false
	}
	if !testCFSAPI() {
		passed = false
	}
	if !testCFSCLI() {
		passed = false
	}
	if !passed {
		common.ArtifactsPods(podNames)
	}
	return
}

func runCLICommand(cmdArgs ...string) []byte {
	return test.RunCLICommandJSON("cfs", cmdArgs...)
}

func checkIDField(mapCmdOut []byte, cfsEndpoint, idFieldName, expectedIdValue string) bool {
	// The endpoint names are plural ending in s -- this makes it singular
	objectName := "CFS " + cfsEndpoint[:len(cfsEndpoint)-1]
	err := common.ValidateStringFieldValue(objectName, idFieldName, expectedIdValue, mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Make basic CFS API calls, checking only status code at this point
func testCFSCLI() (passed bool) {
	passed = true

	common.Infof("Checking CFS CLI endpoints")
	for _, cfsEndpoint := range cfsEndpoints {
		common.Infof("CLI: Listing CFS %s", cfsEndpoint)
		cmdOut := runCLICommand(cfsEndpoint, "list")
		if cmdOut == nil {
			passed = false
			continue
		}

		idFieldName, ok := cfsEndpointIdFieldName[cfsEndpoint]
		if !ok {
			// This endpoint has no GET/describe command
			continue
		}

		// If our list has any entries, let's get the ID field of the
		// first entry. Then we can do a GET/describe on that object
		idValue, err := common.GetStringFieldFromFirstItem(idFieldName, cmdOut)
		if err != nil {
			common.Error(err)
			passed = false
			continue
		} else if len(idValue) == 0 {
			common.Infof("No CFS %s listed -- skipping CLI describe test", cfsEndpoint)
			continue
		}

		common.Infof("CLI: Describing CFS %s %s", cfsEndpoint, idValue)
		cmdOut = runCLICommand(cfsEndpoint, "describe", idValue)
		if cmdOut == nil {
			passed = false
			continue
		}

		// Validate that we find the expected ID field value
		if !checkIDField(cmdOut, cfsEndpoint, idFieldName, idValue) {
			passed = false
		}
	}
	return
}

// Make basic CFS API calls, checking only status code at this point
func testCFSAPI() (passed bool) {
	var url string
	var baseurl string = common.BASEURL

	passed = false
	common.Infof("Checking CFS API endpoints")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	passed = true

	common.Infof("API: Checking CFS service health")
	url = baseurl + endpoints["cfs"]["healthz"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		passed = false
	} else {
		// At least verify that the response object is a string map as we expect
		_, err = common.DecodeJSONIntoStringMap(resp.Body())
		if err != nil {
			common.Error(err)
			passed = false
		}
	}

	for _, cfsEndpoint := range cfsEndpoints {
		common.Infof("API: Listing CFS %s", cfsEndpoint)
		url = baseurl + endpoints["cfs"][cfsEndpoint].Url
		resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
		if err != nil {
			common.Error(err)
			passed = false
			continue
		}

		idFieldName, ok := cfsEndpointIdFieldName[cfsEndpoint]
		if !ok {
			// This endpoint has no GET/describe command
			continue
		}

		// If our list has any entries, let's get the ID field of the
		// first entry. Then we can do a GET/describe on that object
		idValue, err := common.GetStringFieldFromFirstItem(idFieldName, resp.Body())
		if err != nil {
			common.Error(err)
			passed = false
			continue
		} else if len(idValue) == 0 {
			common.Infof("No CFS %s listed -- skipping API GET specific %s test", cfsEndpoint, cfsEndpoint)
			continue
		}

		common.Infof("API: Getting CFS %s %s", cfsEndpoint, idValue)
		url += "/" + idValue
		resp, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
		if err != nil {
			common.Error(err)
			passed = false
			continue
		}

		// Validate that we find the expected ID field value
		if !checkIDField(resp.Body(), cfsEndpoint, idFieldName, idValue) {
			passed = false
		}
	}
	return
}
