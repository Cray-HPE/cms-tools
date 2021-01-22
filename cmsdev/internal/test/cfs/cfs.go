package cfs

/*
 * cfs.go
 * 
 * cfs commons file  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"regexp"
	"strings"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
   	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

func IsCFSRunning(local, smoke, ct bool, crayctlStage string) bool {
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return true
	case "4", "5":
		podNames, ok := test.GetPodNamesByPrefixKey("cfsServices", 2, -1)
		if !ok {
			return false
		}
		common.Infof("Found %d cfsServices pods", len(podNames))

		// 2 pods minimum since we expect both an api and operator pod
		podNames, ok = test.GetPodNamesByPrefixKey("cfs", 2, -1)
		if !ok {
			return false
		}
		common.Infof("Found %d cfs pods", len(podNames))
		apiPodName, operatorPodName := "", "" 
		passed := true
		for _, podName := range podNames {
			// CFS is running if there is operator and api pod
			// we can ignore the state of cfs-jobs or cfs-db pods
			re := regexp.MustCompile(common.PodServiceNamePrefixes["cfsServices"])
			if re.MatchString(podName) {
				if strings.HasPrefix(podName, common.PodServiceNamePrefixes["cfs-api"]) {
					apiPodName = podName
					common.Infof("Found apiPod=%s",podName)
				} else if strings.HasPrefix(podName, common.PodServiceNamePrefixes["cfs-operator"]) {
					operatorPodName = podName
					common.Infof("Found operatorPod=%s",podName)
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
			common.Infof("Pod status is %s",status)
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
		return passed
	default:
		common.Errorf("Invalid stage for this test")
		return false
	}
	common.Errorf("Programming logic error: this line should never be reached")
	return false
}
