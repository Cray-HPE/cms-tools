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
	var arch, pvcName string
	var ok, onMaster bool
	var err error

	iPxePodNameByArch := make(map[string]string)

	// Binaries for each architecture are built in an iPXE pod for that architecture
	for _, arch = range IpxeBinaryArchitectures {
		// Find iPXE pod and verify status
		podNames, ok = test.GetPodNamesInNamespace(common.NAMESPACE, IpxePodPrefixByArch[arch], 1, 1)
		if !ok {
			passed = false
			common.Infof("Found %d %s iPXE pod(s)", len(podNames), arch)
		} else {
			iPxePodNameByArch[arch] = podNames[0]
			common.Infof("Found %s iPXE pod: %s", arch, iPxePodNameByArch[arch])
			if !IpxeContainerReady(iPxePodNameByArch[arch]) {
				passed = false
			}
		}

		if !test.CheckPodListStats(podNames) {
			passed = false
		}
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

	if !passed {
		common.ArtifactsKubernetes()
		if len(podNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, podNames)
		}
		if len(pvcNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, pvcNames)
		}
		common.Infof("Because of previous failures, skipping remaining tftp checks")
		return
	}

	// Even though the file transfer subtest will not run if this is a master NCN,
	// we always have the test check the configmap, just to make sure it doesn't have
	// errors.
	if !GetIpxeBinaryNames() {
		passed = false
	}

	// The file transfer subtest cannot run from master NCNs
	onMaster, err = common.RunningOnMaster()
	if err != nil {
		common.Error(err)
		common.Errorf("Error checking node hostname")
		passed = false
	} else if onMaster == true {
		common.Infof("tftp file transfer test cannot run on master NCNs -- skipping")
	} else {
		for _, srvName := range tftpServiceNames {
			if !TftpServiceFileTransferTest(srvName, iPxePodNameByArch) {
				passed = false
			}
		}
	}

	if !passed {
		common.ArtifactsKubernetes()
		if len(podNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, podNames)
		}
		if len(pvcNames) > 0 {
			common.ArtifactDescribeNamespacePods(common.NAMESPACE, pvcNames)
		}
	}

	return
}
