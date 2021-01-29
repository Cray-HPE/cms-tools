package ipxe_tftp

/*
 * ipxe_tftp.go
 *
 * Tests ipxe and tftp services
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

var pvcNames = []string{
	"cray-tftp-shared-pvc",
}

var tftpServiceNames = []string{
	"cray-tftp",
	"cray-tftp-hmn",
}

func AreTheyRunning() (passed bool) {
	passed = true
	var podNames []string
	var ipxePodName, pvcName string
	var ok bool

	// First validate IPXE k8s status
	podNames, ok = test.GetPodNamesByPrefixKey("ipxe", 1, 1)
	if !ok {
		passed = false
		common.Infof("Found %d ipxe pod(s)", len(podNames))
	} else {
		ipxePodName = podNames[0]
		common.Infof("Found ipxe pod: %s", ipxePodName)
	}

	if !test.CheckPodListStats(podNames) {
		passed = false
	}

	// Next validate TFTP k8s status
	podNames, ok = test.GetPodNamesByPrefixKey("tftp", 1, -1)
	if !ok {
		passed = false
	}
	common.Infof("Found %d tftp pods", len(podNames))
	if !test.CheckPodListStats(podNames) {
		passed = false
	}

	// check pvc status
	for _, pvcName = range pvcNames {
		if !test.CheckPVCStatus(pvcName) {
			passed = false
		}
	}

	if len(ipxePodName) > 0 && !IpxeContainerReady(ipxePodName) {
		passed = false
	}

	if !passed {
		common.ArtifactsPodsPvcs(podNames, pvcNames)
		common.Infof("Because of previous failures, skipping remaining tftp checks")
		return
	}

	for _, srvName := range tftpServiceNames {
		if !TftpServiceFileTransferTest(srvName, ipxePodName) {
			passed = false
		}
	}

	return
}
