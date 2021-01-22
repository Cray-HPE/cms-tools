package test

/*
 * test.go
 * 
 * cms test helper functions
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"sort"
	"strings"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
   	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

// given a service name and minimum and maximum counts, return a list
// a negative value for a minimum or maximum count means it is not checked
func GetPodNamesInNamespace(namespace, serviceName string, minExpectedCount, maxExpectedCount int) (podNames []string, passed bool) {
	var err error
	passed = false
	if minExpectedCount < 0 { 
		minExpectedCount=0
	} else if maxExpectedCount >= 0 && maxExpectedCount < minExpectedCount {
		common.Errorf("Programming logic error: cms.GetPodNames: maxExpectedCount==%d < minExpectedCount==%d",
			maxExpectedCount,minExpectedCount,
		)
		return
	}
	podNames, err = k8s.GetPodNames(common.NAMESPACE, serviceName)
	if err != nil {
		common.Error(err)
	} else if minExpectedCount == maxExpectedCount {
		if len(podNames) != minExpectedCount {
			common.Errorf("Expected exactly %d pod pods but found %d: %s",
				minExpectedCount,len(podNames),strings.Join(podNames,", "),
			)
		} else {
			passed = true 
		}
	} else if len(podNames) < minExpectedCount {
		common.Errorf("Expected at least %d pod pods but found %d: %s",
			minExpectedCount,len(podNames),strings.Join(podNames,", "),
		)
	} else if maxExpectedCount > 0 && len(podNames) > maxExpectedCount {
		common.Errorf("Expected at most %d pod pods but found %d: %s",
			maxExpectedCount,len(podNames),strings.Join(podNames,", "),
		)
	} else { 
		passed = true 
	}
	return
}

func GetPodNames(serviceName string, minExpectedCount, maxExpectedCount int) ([]string, bool) {
	return GetPodNamesInNamespace(common.NAMESPACE, serviceName, minExpectedCount, maxExpectedCount)
}

func GetPodNamesByPrefixKey(pkey string, minExpectedCount, maxExpectedCount int) ([]string, bool) {
	return GetPodNames(common.PodServiceNamePrefixes[pkey], minExpectedCount, maxExpectedCount)
}

// given a service name and expected minimum and maximum counts, return a list
// a negative value for a minimum or maximum count means it is not checked
func GetPVCNamesInNamespace(namespace, serviceName string, minExpectedCount, maxExpectedCount int) (pvcNames []string, passed bool) {
	passed = false
	// check bos pvc pod status
	if minExpectedCount < 0 { 
		minExpectedCount=0
	} else if maxExpectedCount >= 0 && maxExpectedCount < minExpectedCount {
		common.Errorf("Programming logic error: cms.GetPVCNames: maxExpectedCount==%d < minExpectedCount==%d",
			maxExpectedCount,minExpectedCount,
		)
		return
	}
	pvcNames, err := k8s.GetPVCNames(namespace, serviceName)
	if err != nil {
		common.Error(err)
		return
	} else if minExpectedCount == maxExpectedCount {
		if len(pvcNames) != minExpectedCount {
			common.Errorf("Expected exactly %d pvc pods but found %d: %s",
				minExpectedCount,len(pvcNames),strings.Join(pvcNames,", "),
			)
		} else {
			passed = true
		}
	} else if len(pvcNames) < minExpectedCount {
		common.Errorf("Expected at least %d pvc pods but found %d: %s",
			minExpectedCount,len(pvcNames),strings.Join(pvcNames,", "),
		)
	} else if maxExpectedCount > 0 && len(pvcNames) > maxExpectedCount {
		common.Errorf("Expected at most %d pvc pods but found %d: %s",
			maxExpectedCount,len(pvcNames),strings.Join(pvcNames,", "),
		)
	} else {
		passed = true
	}
	return
}

func GetPVCNames(serviceName string, minExpectedCount, maxExpectedCount int) ([]string, bool) {
	return GetPVCNamesInNamespace(common.NAMESPACE, serviceName, minExpectedCount, maxExpectedCount)
}

func GetPVCNamesByPrefixKey(pkey string, minExpectedCount, maxExpectedCount int) ([]string, bool) {
	return GetPVCNames(common.PodServiceNamePrefixes[pkey], minExpectedCount, maxExpectedCount)
}

// give a pod's name, check to see if it and its containers are running
// Logs any errors encountered
// Returns true if no errors, false otherwise
func CheckPodStatsInNamespace(namespace, podName string, targeted bool, params ...string) bool {
	var err error
	var stats *k8s.PodStats

	// get pod and container status details
	common.Infof("checking pod status for %s expecting %s", podName, "Running")
	stats, err = k8s.GetPodStats(common.NAMESPACE, podName)
	if err != nil {
		common.Error(err)
		common.VerboseFailed()
		return false
	} else if stats.Phase != "Running" && stats.Phase != "Succeeded" {
		common.Errorf("pod=%s, phase (%s) is neither Running or Succeeded", podName, stats.Phase)
		common.VerboseFailed()
		return false
	} else if len(stats.ContainerStateWaitingReason) != 0 {
		// check here if all pods are in working order
		if targeted {
			// if targeted container [0] does not exist or is not in the preferred state [1]
	 		_, ok := stats.ContainerStateWaitingReason[params[0]]
	 		if ! ok || stats.ContainerStateWaitingReason[params[0]] != params[1] {
				common.Errorf("targeted podName=%s, has unexpected ContainerStateWaitingReason", podName)
				common.VerboseFailed()
				return false
			} 
			delete(stats.ContainerStateWaitingReason, params[0])
		}
		// checked each container status, the expected state for non targeted is the default 'Running'
		ok := true
		for k, _ := range stats.ContainerStateWaitingReason {
			if stats.ContainerStateWaitingReason[k] != "Running" {
				common.Errorf("podName=%s, %s!=Running", podName, stats.ContainerStateWaitingReason[k])
				ok = false
			}
		}
		if !ok {
			common.VerboseFailed()
			return false
		}
	}
	common.VerboseOkayf("Pod status is as expected")
	return true
}

// Wrapper for CheckPodStatsInNamespace that assumes namespace = common.NAMESPACE
func CheckPodStats(podName string, targeted bool, params ...string) bool {
	return CheckPodStatsInNamespace(common.NAMESPACE, podName, targeted, params...)
}

// give a pvc name, check to see if its status is Bound
// Logs any errors encountered
// Returns true if no errors, false otherwise
func CheckPVCStatusInNamespace(namespace, pvcName string) (passed bool) {
	passed = false
	common.Infof("checking pvc status for %s expecting %s", pvcName, "Bound")
	status, err := k8s.GetPVCStatus(namespace, pvcName); if err == nil {
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

func CheckPVCStatusByPrefixKey(pkey string) bool {
	return CheckPVCStatus(common.PodServiceNamePrefixes[pkey])
}

// Given a list of pod names, calls CheckPodStats on each one of them.
func CheckPodListStatsInNamespace(namespace string, podNames []string) (passed bool) {
	passed = true
	for _, podName := range podNames {
		if !CheckPodStatsInNamespace(namespace, podName, false) {
			passed = false
		}
	}
	return
}

func CheckPodListStats(podNames []string) bool {
	return CheckPodListStatsInNamespace(common.NAMESPACE, podNames)
}

// Calls GetPodNames to get a list of pods for this service
// Calls CheckPodListStats on that list
func CheckServicePodStatsInNamespace(namespace, serviceName string, minExpectedCount, maxExpectedCount int) bool {
	// check pod and container status
	podNames, ok := GetPodNamesInNamespace(namespace, serviceName, minExpectedCount, maxExpectedCount)
	if !ok {
		return false
	}
	sort.Strings(podNames)
	return CheckPodListStatsInNamespace(namespace, podNames)
}

func CheckServicePodStats(serviceName string, minExpectedCount, maxExpectedCount int) bool {
	return CheckServicePodStatsInNamespace(common.NAMESPACE, serviceName, minExpectedCount, maxExpectedCount)
}

func CheckServicePodStatsByPrefixKey(pkey string, minExpectedCount, maxExpectedCount int) bool {
	return CheckServicePodStats(common.PodServiceNamePrefixes[pkey], minExpectedCount, maxExpectedCount)
}

// Given a list of pvc names, calls CheckPVCStatus on each one of them.
func CheckPVCListStatsInNamespace(namespace string, pvcNames []string) (passed bool) {
	passed = true
	for _, pvcName := range pvcNames {
		if !CheckPVCStatusInNamespace(namespace, pvcName) {
			passed = false
		}
	}
	return
}

func CheckPVCListStats(pvcNames []string) bool {
	return CheckPVCListStatsInNamespace(common.NAMESPACE, pvcNames)
}

// Calls GetPVCNames to get a list of pvcs for this service
// Calls CheckPVCListStats on that list
func CheckServicePVCStatusInNamespace(namespace, serviceName string, minExpectedCount, maxExpectedCount int) bool {
	// check pvc and container status
	pvcNames, ok := GetPVCNamesInNamespace(namespace, serviceName, minExpectedCount, maxExpectedCount)
	if !ok {
		return false
	}
	sort.Strings(pvcNames)
	return CheckPVCListStatsInNamespace(namespace, pvcNames)
}

func CheckServicePVCStatus(serviceName string, minExpectedCount, maxExpectedCount int) bool {
	return CheckServicePVCStatusInNamespace(common.NAMESPACE, serviceName, minExpectedCount, maxExpectedCount)
}

func CheckServicePVCStatusByPrefixKey(pkey string, minExpectedCount, maxExpectedCount int) bool {
	return CheckServicePVCStatus(common.PodServiceNamePrefixes[pkey], minExpectedCount, maxExpectedCount)
}
