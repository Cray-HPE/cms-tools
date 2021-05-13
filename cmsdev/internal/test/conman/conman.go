package conman

/*
 * conman.go
 *
 * conman commons file
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
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

var pvcNames = []string{
	"cray-conman-data-claim",
}

func IsConmanRunning() (passed bool) {
	passed = true

	podNames, ok := test.GetPodNamesByPrefixKey("conman", 1, 1)
	if !ok {
		passed = false
	}
	for _, podName := range podNames {
		if !test.CheckPodStats(podName, "cray-conman", "Running") {
			passed = false
		}
	}

	// check conman pvc status
	for _, pvcName := range pvcNames {
		if !test.CheckPVCStatus(pvcName) {
			passed = false
		}
	}

	if !passed {
		common.ArtifactsPodsPvcs(podNames, pvcNames)
	}
	return
}
