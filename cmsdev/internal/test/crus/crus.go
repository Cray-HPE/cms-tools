package crus

/*
 * crus.go
 *
 * crus commons file
 *
 * Copyright 2020 Hewlett Packard Enterprise Development LP
 */

import (
	"net/http"
	"os/exec"
	"regexp"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()

func IsCRUSRunning(local, smoke, ct bool, crayctlStage string) (passed bool) {
	passed = true
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return
	case "4", "5":
		// We expect to find 4 or more cray-crus pods
		podNames, ok := test.GetPodNamesByPrefixKey("crus", 4, -1)
		if !ok {
			passed = false
		}
		common.Infof("Found %d crus pods", len(podNames))

		slurmOkay := slurmMapPresent()
		mungeOkay := mungeSecretPresent()
		crusReady := slurmOkay && mungeOkay

		// We expect the following crus pods:
		// Exactly 1 main cray-crus pod
		// Exactly 3 cray-crus-etcd pods
		// Optionally 1 or more cray-crus-wait-for-etcd pods
		//
		// If munge or slurm are not yet ready, then we simply do not want the
		// crus pods to be in a failed state
		mainPodCount := 0
		etcPodCount := 0
		waitForEtcdRe := regexp.MustCompile("cray-crus-wait-for-etcd-[0-9][0-9]*-")
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
				if crusReady {
					// In this case, we expect the wait-for-etcd pod to be in the Succeeded state
					if status != "" && status != "Succeeded" {
						common.VerboseFailedf("Pod %s has status %s, but we expect it to be Succeeded", podName, status)
						passed = false
					}
				} else {
					// In this case, we're fine if it's in Pending, Running, or Succeeded
					if status != "" && status != "Pending" && status != "Running" && status != "Succeeded" {
						common.VerboseFailedf("Pod %s has status %s, but we expect it to be Pending, Running, or Succeeded", podName, status)
						passed = false
					}
				}
			} else {
				if strings.HasPrefix(podName, "cray-crus-etcd-") {
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
				if crusReady {
					// In this case, we expect the pod to be in the Running state
					if status != "" && status != "Running" {
						common.VerboseFailedf("Pod %s has status %s, but we expect it to be Running", podName, status)
						passed = false
					}
				} else {
					// In this case, we're fine if it's in Pending, Running, or Succeeded
					if status != "" && status != "Pending" && status != "Running" {
						common.VerboseFailedf("Pod %s has status %s, but we expect it to be Pending or Running", podName, status)
						passed = false
					}
				}
			}
		}
		if mainPodCount == 0 {
			common.VerboseFailedf("Did not find any main cray-crus pod")
			passed = false
		}
		if etcPodCount != 3 {
			common.VerboseFailedf("Found %d cray-crus-etcd- pod(s), but expect to find 3", etcPodCount)
			passed = false
		}

		if !passed {
			common.ArtifactsPodsPvcs(podNames, pvcNames)
		}

		if crusReady {
			if !testCRUSAPI() {
				passed = false
			}
			if !testCRUSCLI() {
				passed = false
			}
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

// Returns True if the slurm-map configmap is present, false otherwise
func slurmMapPresent() bool {
	var cmd *exec.Cmd
	var cmdOut []byte
	var kubectlPath string
	var err error

	kubectlPath, err = k8s.GetKubectlPath()
	if err != nil {
		common.Error(err)
		return false
	}
	cmd = exec.Command(kubectlPath, "get", "configmaps", "-n", common.NAMESPACE, "slurm-map")
	cmdOut, err = cmd.CombinedOutput()
	if err != nil || len(cmdOut) == 0 {
		common.Infof("slurm-map configmap does not appear to be present; cray-crus not expected to be working yet")
		return false
	}
	common.Infof("slurm-map configmap appears to be present")
	return true
}

// Returns True if the munge-secret secret is present, false otherwise
func mungeSecretPresent() bool {
	var cmd *exec.Cmd
	var cmdOut []byte
	var kubectlPath string
	var err error

	kubectlPath, err = k8s.GetKubectlPath()
	if err != nil {
		common.Error(err)
		return false
	}
	cmd = exec.Command(kubectlPath, "get", "secrets", "-n", common.NAMESPACE, "munge-secret")
	cmdOut, err = cmd.CombinedOutput()
	if err != nil || len(cmdOut) == 0 {
		common.Infof("munge-secret secret does not appear to be present; cray-crus not expected to be working yet")
		return false
	}
	common.Infof("munge-secret secret appears to be present")
	return true
}

// Make basic CRUS API call, checking only status code at this point
func testCRUSCLI() bool {
	common.Infof("Checking CRUS CLI list sessions")
	cmdOut := test.RunCLICommand("cray crus session list --format json -vvv")
	if cmdOut == nil {
		return false
	}
	return true
}

// Make basic CRUS API call, checking only status code at this point
func testCRUSAPI() bool {
	var baseurl string = common.BASEURL

	common.Infof("Checking CRUS API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}

	common.Infof("API: Listing CRUS sessions")
	url := baseurl + endpoints["crus"]["session"].Url
	_, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}
