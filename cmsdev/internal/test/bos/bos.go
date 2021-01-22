package bos

/*
 * bos.go
 *
 * bos commons file
 *
 * Copyright 2019-2020 Hewlett Packard Enterprise Development LP
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

// BOS supported API tests
var APITestTypes = []string{
	"post",
	"session",
	"sessiontemplate",
	"smoke",
	"version",
}

func IsBOSRunning(local, smoke, ct bool, crayctlStage string) (passed bool) {
	passed = true
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return
	case "4", "5":
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
		if smoke {
			common.Verbosef("")
			smokeTests()
		}
		return
	default:
		common.Errorf("Invalid stage for this test")
		passed = false
		return
	}
	common.Errorf("Programming logic error: this line should never be reached")
	passed = false
	return
}

// access to BOS API tests
func RunAPITests(local bool, params ...string) {
	if IsBOSRunning(local, false, false, "4") == false {
		common.Warnf("aborting API tests, BOS service is not ready run cmsdev test bos --smoke -v")
		return
	}

	// TODO: ensure this works for local testing
	if local {
		common.Warnf("--local currently not supported")
		return
	}

	// run all API tests
	if len(params) == 0 {
		smokeTests()
		postSessionTemplateTests()
		sessionTestsAPI()
		sessionTemplateTestsAPI()
		versionTests()
		return
	}
	switch params[0] {
	case "post":
		postSessionTemplateTests()
	case "session":
		sessionTestsAPI()
	case "sessiontemplate":
		sessionTemplateTestsAPI()
	case "smoke":
		smokeTests()
	case "version":
		versionTests()
	default:
		common.Warnf("--api argument required. current available tests are: %s",
			strings.Join(APITestTypes, " "))
	}
}
