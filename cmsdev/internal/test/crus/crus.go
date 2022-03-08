//
//  MIT License
//
//  (C) Copyright 2020-2022 Hewlett Packard Enterprise Development LP
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
package crus

/*
 * crus.go
 *
 * crus commons file
 *
 */

import (
	"net/http"
	"os/exec"
	"regexp"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()

func IsCRUSRunning() (passed bool) {
	passed = true
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

func getFirstUpgradeId(listCmdOut []byte) (string, error) {
	return common.GetStringFieldFromFirstItem("upgrade_id", listCmdOut)
}

func runCLICommand(cmdArgs ...string) []byte {
	return test.RunCLICommandJSON("crus", cmdArgs...)
}

func validateUpgradeId(mapCmdOut []byte, expectedUpgradeId string) bool {
	err := common.ValidateStringFieldValue("CRUS session", "upgrade_id", expectedUpgradeId, mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Make basic CRUS CLI call, checking only status code at this point
func testCRUSCLI() bool {
	common.Infof("CLI: List CRUS sessions")
	cmdOut := runCLICommand("session", "list")
	if cmdOut == nil {
		return false
	}

	// If any CRUS sessions were listed, use the upgrade_id of the first one found
	// to test the describe CLI option
	upgradeId, err := getFirstUpgradeId(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	} else if len(upgradeId) == 0 {
		common.Infof("No CRUS sessions listed -- skipping CLI describe {upgrade_id} test")
		return true
	}

	common.Infof("CLI: Describe CRUS session %s", upgradeId)
	cmdOut = runCLICommand("session", "describe", upgradeId)
	if cmdOut == nil {
		return false
	} else if !validateUpgradeId(cmdOut, upgradeId) {
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

	common.Infof("API: List CRUS sessions")
	url := baseurl + endpoints["crus"]["session"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	// If any CRUS sessions were listed, use the upgrade_id of the first one found
	// to test the describe CLI option
	upgradeId, err := getFirstUpgradeId(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(upgradeId) == 0 {
		common.Infof("No CRUS sessions listed -- skipping API GET {upgrade_id} test")
		return true
	}

	url = baseurl + endpoints["crus"]["session"].Url + "/" + upgradeId
	common.Infof("API: Get CRUS session %s", upgradeId)
	resp, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	// Validate that our CRUS session has the ID that we expect
	if !validateUpgradeId(resp.Body(), upgradeId) {
		return false
	}
	return true
}
