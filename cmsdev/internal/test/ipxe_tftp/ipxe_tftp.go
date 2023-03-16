// MIT License
//
// (C) Copyright 2019-2023 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
package ipxe_tftp

/*
 * ipxe_tftp.go
 *
 * Tests ipxe and tftp services
 *
 */

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
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
	var ok, onMaster bool
	var err error

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

	// The file transfer subtest cannot run from master NCNs
	onMaster, err = common.RunningOnMaster()
	if err != nil {
		common.Error(err)
		common.Errorf("Error checking node hostname")
		passed = false
		return
	} else if onMaster == true {
		common.Infof("tftp file transfer test cannot run on master NCNs -- skipping")
		return
	}
	for _, srvName := range tftpServiceNames {
		if !TftpServiceFileTransferTest(srvName, ipxePodName) {
			passed = false
		}
	}

	return
}
