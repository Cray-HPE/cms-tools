package bos 

/*
 * bos_post_sessiontemplate.go
 * 
 * bos post tests  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"net/http"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// post tests
func postSessionTemplateTests(local bool) {
	var baseurl string = common.BASEURL
	const totalNumTests int = 1

	numTests, numTestsFailed := 0, 0
	token := test.GetAccessToken()
	if len(token) == 0 { return }

	// scenario # 1, POST /sessiontemplate endpoint
	data := []byte(`{"name": "cmsdev-test-id-abc3", "boot_sets": { "boot_set1": {"boot_ordinal": 1, "ims_image_id": "efdfe6fc-af3f-40f0-9053-dd1ad6c359d3", "kernel_parameters": "console=tty0 console=ttyS0,115200n8 root=crayfs nfsserver=10.2.0.1 nfspath=/var/opt/cray/boot_images imagename=/SLES15 selinux=0 rd.shell rd.net.timeout.carrier=40 rd.retry=40 ip=dhcp rd.neednet=1 crashkernel=256M htburl=https://10.2.100.50/apis/hbtd/hmi/v1/heartbeat", "network": "nmn", "node_list": ["x0c0s28b0n0"], "rootfs_provider": "", "rootfs_provider_passthrough": ""}, "boot_set2": {"boot_ordinal": 1, "ims_image_id": "efdfe6fc-af3f-40f0-9053-dd1ad6c359d3", "kernel_parameters": "console=tty0 console=ttyS0,115200n8 root=crayfs nfsserver=10.2.0.1 nfspath=/var/opt/cray/boot_images imagename=/SLES15 selinux=0 rd.shell rd.net.timeout.carrier=40 rd.retry=40 ip=dhcp rd.neednet=1 crashkernel=256M htburl=https://10.2.100.50/apis/hbtd/hmi/v1/heartbeat", "network": "nmn", "node_list": ["x0c0s28b0n0"], "rootfs_provider": "", "rootfs_provider_passthrough": ""}}, "cfs_branch": "my-test-branch", "cfs_url": "https://api-gw-service-nmn.local/vcs/cray/config-management.git", "enable_cfs": true, "partition": "p1"}`)
	copy(data[18:], common.GetRandomString(4))
	args := common.Params{
		Token: token,
		JsonStrArray: data,
	}
	url := baseurl + endpoints["bos"]["sessiontemplate"].Url 
	numTests++
	test.RestfulTestHeader("POST sessiontemplate", numTests, totalNumTests)
	common.Infof("args: %s", string(args.JsonStrArray))
	_, err := test.RestfulVerifyStatus("POST", url, args, http.StatusCreated)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	}

	test.RestfulTestResultSummary(numTestsFailed, numTests)
}
