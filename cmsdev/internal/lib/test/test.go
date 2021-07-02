package test

/*
 * test.go
 *
 * cms test helper functions
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
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"strings"
)

// given a service name and minimum and maximum counts, return a list
// a negative value for a minimum or maximum count means it is not checked
func GetPodNamesInNamespace(namespace, serviceName string, minExpectedCount, maxExpectedCount int) (podNames []string, passed bool) {
	var err error
	passed = false
	if minExpectedCount < 0 {
		minExpectedCount = 0
	} else if maxExpectedCount >= 0 && maxExpectedCount < minExpectedCount {
		common.Errorf("Programming logic error: cms.GetPodNames: maxExpectedCount==%d < minExpectedCount==%d",
			maxExpectedCount, minExpectedCount,
		)
		return
	}
	podNames, err = k8s.GetPodNames(namespace, serviceName)
	if err != nil {
		common.Error(err)
		return
	} else if minExpectedCount == maxExpectedCount {
		if len(podNames) != minExpectedCount {
			common.Errorf("Expected exactly %d pod pods but found %d: %s",
				minExpectedCount, len(podNames), strings.Join(podNames, ", "),
			)
		} else {
			passed = true
		}
	} else if len(podNames) < minExpectedCount {
		common.Errorf("Expected at least %d pod pods but found %d: %s",
			minExpectedCount, len(podNames), strings.Join(podNames, ", "),
		)
	} else if maxExpectedCount > 0 && len(podNames) > maxExpectedCount {
		common.Errorf("Expected at most %d pod pods but found %d: %s",
			maxExpectedCount, len(podNames), strings.Join(podNames, ", "),
		)
	} else {
		passed = true
	}
	return
}

func GetPodNamesByPrefixKey(pkey string, minExpectedCount, maxExpectedCount int) ([]string, bool) {
	return GetPodNamesInNamespace(common.NAMESPACE, common.PodServiceNamePrefixes[pkey], minExpectedCount, maxExpectedCount)
}

// give a pod's name, check to see if it and its containers are running
// Logs any errors encountered
// Returns true if no errors, false otherwise
func CheckPodStatsInNamespace(namespace, podName string, params ...string) (passed bool) {
	var err error
	var stats *k8s.PodStats
	passed = true

	// get pod and container status details
	common.Infof("checking pod status for %s expecting %s", podName, "Running")
	stats, err = k8s.GetPodStats(namespace, podName)
	if err != nil {
		common.Error(err)
		passed = false
	} else if stats.Phase != "Running" && stats.Phase != "Succeeded" {
		common.Errorf("pod=%s, phase (%s) is neither Running or Succeeded", podName, stats.Phase)
		passed = false
	} else if len(stats.ContainerStateWaitingReason) != 0 {
		// check here if all pods are in working order
		var containerName, containerState, expectedContainerState string
		if len(params) == 2 {
			var ok bool
			containerName, expectedContainerState = params[0], params[1]

			// Check if targeted container does not exist or is not in the preferred state
			common.Infof("Checking ContainerStateWaitingReason of container %s in pod %s (expecting %s)", containerName, podName,
				expectedContainerState)
			containerState, ok = stats.ContainerStateWaitingReason[params[0]]
			if !ok {
				common.Errorf("Error checking ContainerStateWaitingReason of container %s in pod %s", containerName, podName)
				passed = false
			} else {
				if containerState != expectedContainerState {
					common.Errorf("Container %s in pod %s, has ContainerStateWaitingReason=%s but we expect %s", containerName,
						podName, containerState, expectedContainerState)
					passed = false
				} else {
					common.Infof("Container %s in pod %s has expected ContainerStateWaitingReason", containerName, podName)
				}
				delete(stats.ContainerStateWaitingReason, params[0])
			}
		} else if len(params) > 0 {
			common.Errorf("PROGRAMMING LOGIC ERROR: Invalid # (%d) of params: %v", len(params), params)
			passed = false
		}

		// checked each container status, the expected state for non targeted is the default 'Running'
		expectedContainerState = "Running"
		for containerName, containerState = range stats.ContainerStateWaitingReason {
			common.Infof("Checking ContainerStateWaitingReason of container %s in pod %s (expecting %s)", containerName, podName,
				expectedContainerState)
			if containerState != expectedContainerState {
				common.Errorf("Container %s in pod %s, has ContainerStateWaitingReason=%s but we expect %s", containerName,
					podName, containerState, expectedContainerState)
				passed = false
			} else {
				common.Infof("Container %s in pod %s has expected ContainerStateWaitingReason", containerName, podName)
			}
		}
	}
	if passed {
		common.VerboseOkayf("Pod status is as expected")
	} else {
		common.VerboseFailed()
	}
	return
}

// Wrapper for CheckPodStatsInNamespace that assumes namespace = common.NAMESPACE
func CheckPodStats(podName string, params ...string) bool {
	return CheckPodStatsInNamespace(common.NAMESPACE, podName, params...)
}

// give a pvc name, check to see if its status is Bound
// Logs any errors encountered
// Returns true if no errors, false otherwise
func CheckPVCStatusInNamespace(namespace, pvcName string) (passed bool) {
	passed = false
	common.Infof("checking pvc status for %s expecting %s", pvcName, "Bound")
	status, err := k8s.GetPVCStatus(namespace, pvcName)
	if err == nil {
		if len(status) == 0 {
			common.Errorf("pvc status length returned 0")
		} else if status != "Bound" {
			common.Errorf("Expected Bound status for pvc=%s, found status=%s", pvcName, status)
		} else {
			passed = true
			common.VerboseOkayf("pvc status is as expected")
			return
		}
	} else {
		common.Error(err)
	}
	common.VerboseFailed()
	return
}

func CheckPVCStatus(pvcName string) bool {
	return CheckPVCStatusInNamespace(common.NAMESPACE, pvcName)
}

// Given a list of pod names, calls CheckPodStats on each one of them.
func CheckPodListStatsInNamespace(namespace string, podNames []string) (passed bool) {
	passed = true
	for _, podName := range podNames {
		if !CheckPodStatsInNamespace(namespace, podName) {
			passed = false
		}
	}
	return
}

func CheckPodListStats(podNames []string) bool {
	return CheckPodListStatsInNamespace(common.NAMESPACE, podNames)
}
