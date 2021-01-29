package cfs

/*
 * cfs.go
 *
 * cfs commons file
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"fmt"
	"net/http"
	"regexp"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
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

// Make basic CFS API calls, checking only status code at this point
func testCFSCLI() (passed bool) {
	passed = true

	common.Infof("Checking CFS CLI endpoints")
	for _, cfsEndpoint := range cfsEndpoints {
		cmdString := fmt.Sprintf("cray cfs %s list --format json -vvv", cfsEndpoint)
		cmdOut := test.RunCLICommand(cmdString)
		if cmdOut == nil {
			passed = false
		}
	}
	return
}

// Make basic CFS API calls, checking only status code at this point
func testCFSAPI() (passed bool) {
	var baseurl string = common.BASEURL

	passed = false
	common.Infof("Checking CFS API endpoints")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	passed = true

	common.Infof("API: Checking CFS service health")
	url := baseurl + endpoints["cfs"]["healthz"].Url
	_, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		passed = false
	}

	for _, cfsEndpoint := range cfsEndpoints {
		common.Infof("API: Listing CFS %s", cfsEndpoint)
		url := baseurl + endpoints["cfs"][cfsEndpoint].Url
		_, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
		if err != nil {
			common.Error(err)
			passed = false
		}
	}
	return
}
