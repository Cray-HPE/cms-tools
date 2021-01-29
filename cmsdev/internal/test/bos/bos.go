package bos

/*
 * bos.go
 *
 * bos commons file
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"regexp"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()

func IsBOSRunning() (passed bool) {
	passed = true

	// We expect to find 4+ cray-bos pods
	podNames, ok := test.GetPodNamesByPrefixKey("bos", 4, -1)
	if !ok {
		passed = false
	}
	common.Infof("Found %d bos pods", len(podNames))
	// We expect the following bos pods
	// Exactly 1 main cray-bos pod
	// Exactly 3 cray-bos-etcd pods
	// Optionally 1 or more cray-bos-wait-for-etcd pods
	mainPodCount := 0
	etcPodCount := 0
	waitForEtcdRe := regexp.MustCompile("cray-bos-wait-for-etcd-[0-9][0-9]*-")
	pvcNames := make([]string, 0, 3)
	for _, podName := range podNames {
		common.Infof("Getting pod status for %s", podName)
		status, err := k8s.GetPodStatus(common.NAMESPACE, podName)
		if err != nil {
			common.VerboseFailedf(err.Error())
			passed = false
			status = ""
		} else {
			common.Infof("Pod %s status is %s", podName, status)
		}

		if waitForEtcdRe.MatchString(podName) {
			if status != "" && status != "Succeeded" {
				common.VerboseFailedf("Pod %s has status %s, but we expect it to be Succeeded", podName, status)
				passed = false
			}
		} else {
			if strings.HasPrefix(podName, "cray-bos-etcd-") {
				etcPodCount += 1
				// There should be a corresponding pvc with the same name
				common.Infof("There should be a corresponding pvc with the same name as this pod (%s)", podName)
				pvcNames = append(pvcNames, podName)
				if !test.CheckPVCStatus(podName) {
					passed = false
				}
			} else {
				mainPodCount += 1
			}
			if status != "" && status != "Running" {
				common.VerboseFailedf("Pod %s has status %s, but we expect it to be Running", podName, status)
				passed = false
			}
		}
	}
	if mainPodCount == 0 {
		common.VerboseFailedf("Did not find any main cray-bos pod")
		passed = false
	}
	if etcPodCount != 3 {
		common.VerboseFailedf("Found %d cray-bos-etcd- pod(s), but expect to find 3", etcPodCount)
		passed = false
	}

	if !passed {
		common.ArtifactsPodsPvcs(podNames, pvcNames)
	}

	// Run basic API and CLI tests
	if !versionTests() {
		passed = false
	}
	if !sessionTemplateTestsAPI() {
		passed = false
	}
	if !sessionTemplateTestsCLI() {
		passed = false
	}
	if !sessionTestsAPI() {
		passed = false
	}
	if !sessionTestsCLI() {
		passed = false
	}

	return
}
