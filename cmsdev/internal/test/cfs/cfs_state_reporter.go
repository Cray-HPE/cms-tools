package cfs

/*
 * cfs_state_reporter.go
 *
 * Verify status of cfs-state-reporter on other NCNs
 *
 * Copyright 2021-2022 Hewlett Packard Enterprise Development LP
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
	"fmt"
	resty "gopkg.in/resty.v1"
	coreV1 "k8s.io/api/core/v1"
	"regexp"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"strings"
)

const cfsStateReporterCheckPath = "/opt/cray/tests/install/ncn/scripts/python/cfs_state_reporter_check.py"
const gossTestPort = 9009
const gossTestEndpoint = "goss-cfs-state-reporter"
const gossUrlFormat = "http://%s.hmn:%d/%s"

// Makes curl call to cfs_state_reporter_check goss endpoint on remote host and
// examines the status code
func checkCfsStateReporterOnRemoteHost(remoteHost string) (ok bool) {
	var err error
	var gossUrl string
	var resp *resty.Response

	ok = false
	common.Infof("Checking cfs-state-reporter status on %s", remoteHost)
	gossUrl = fmt.Sprintf(gossUrlFormat, remoteHost, gossTestPort, gossTestEndpoint)
	common.Debugf("goss URL = %s", gossUrl)

	// Create client to use for goss API call
	// If performance was a concern this could be moved outside this function so we do not
	// create a new client for every request, but the gains in this case would be negligible.
	client := resty.New()
	client.SetHeaders(map[string]string{
		"Accept":     "application/json",
		"User-Agent": "cmsdev",
	})

	common.Debugf("GET %s", gossUrl)
	resp, err = client.R().Get(gossUrl)
	if err != nil {
		common.Error(err)
		common.Errorf("Unable to run cfs-state-reporter check on %s. To do the check manually, run %s on %s",
			remoteHost, cfsStateReporterCheckPath, remoteHost)
		return
	}
	common.Debugf("resp=%v", resp)
	common.PrettyPrintJSON(resp)

	// 200 is the expected status code for test success
	// 503 is the expected status code for test failure
	if resp.StatusCode() == 200 {
		ok = true
		common.Debugf("200 status code from goss API endpoint indicates the cfs-state-reporter test passed on %s",
			remoteHost)
		common.Infof("cfs-state-reporter check PASSED on %s", remoteHost)
	} else if resp.StatusCode() == 503 {
		common.Debugf("503 status code from goss API endpoint indicates the cfs-state-reporter test failed on %s",
			remoteHost)
		common.Errorf("cfs-state-reporter check on %s failed. For details, run %s on %s",
			remoteHost, cfsStateReporterCheckPath, remoteHost)
	} else {
		common.Errorf("Expected status code of 200 or 503 from goss API call, but received %d", resp.StatusCode())
		common.Errorf("Unable to run cfs-state-reporter check on %s. To do the check manually, run %s on %s",
			remoteHost, cfsStateReporterCheckPath, remoteHost)
	}
	return
}

// Extract from /etc/hosts the list of all NCNs, except for the current host
func getNcns() (ncns []string, err error) {
	// Explanation of the piped commands being run:
	// 1. The first grep commands extracts all strings from /etc/hosts of the form ncn-X### where X is m, s, or w
	// 2. But because the grep makes sure it is followed by whitespace or an end-of-line, we add the sed command,
	//   to strip off any trailing whitespace
	// 3. We use the sort command to remove duplicates
	var etcHostsNcns []string
	var k8sNodes []coreV1.Node
	var myHostname string
	var removeNcnM001 bool

	// First determine our hostname, so we can filter it from the list later
	common.Debugf("Finding hostname of local host")
	result, err := common.RunPath("/usr/bin/hostname", "-s")
	if err != nil {
		err = fmt.Errorf("Error determining hostname of local host: %v", err)
		return
	} else if result.Rc != 0 {
		err = fmt.Errorf("Error determining hostname of local host: hostname command exit code=%d", result.Rc)
		return
	}
	// Trim whitespace from the output
	myHostname = strings.TrimSpace(result.OutString())
	common.Debugf("According to hostname command, local hostname is '%s'", myHostname)
	if len(myHostname) == 0 {
		err = fmt.Errorf("Error determining hostname of local host: hostname command gave no output")
		return
	} else if len(myHostname) != 8 {
		common.Debugf("Hostnames of the expected form will be exactly 8 characters long, but this one is %d", len(myHostname))
		err = fmt.Errorf("Local hostname should be of form ncn-[msw]### but it is '%s'", myHostname)
		return
	} else if match, _ := regexp.MatchString("ncn-[msw][0-9][0-9][0-9]", myHostname); !match {
		err = fmt.Errorf("Local hostname should be of form ncn-[msw]### but it is '%s'", myHostname)
		return
	}

	// Now extract list of NCNs from /etc/hosts
	common.Debugf("Getting list of NCNs from /etc/hosts")
	result, err = common.RunPath("/bin/bash", "-c",
		"grep -Eo 'ncn-[msw][0-9][0-9][0-9]([[:space:]]|$)' /etc/hosts | sed 's/[[:space:]]*$//g' | sort -u")
	if err != nil {
		err = fmt.Errorf("Error getting NCN hostnames from /etc/hosts: %v", err)
		return
	} else if result.Rc != 0 {
		err = fmt.Errorf("Error getting NCN hostnames from /etc/hosts: bash command exit code=%d", result.Rc)
		return
	}
	etcHostsNcns = strings.Fields(result.OutString())
	common.Debugf("Found the following NCN hostnames in /etc/hosts: %s", strings.Join(etcHostsNcns, " "))

	// We do not want to include ncn-m001 if it is still the PIT node, or if it is our local host
	if myHostname == "ncn-m001" {
		removeNcnM001 = true
	} else {
		// We want to see if ncn-m001 is still the PIT node.
		// To check for this, we list the nodes in Kubernetes and see if ncn-m001 is included.
		// Being the PIT node is not the only reason ncn-m001 may be excluded from this list,
		// but it is by far the most likely.
		common.Debugf("Getting list of Kubernetes nodes")
		k8sNodes, err = k8s.GetNodes("ncn-m001")
		if err != nil {
			err = fmt.Errorf("Error getting list of nodes from Kubernetes: %v", err)
			return
		}
		ncnM001Found := false
		for _, node := range k8sNodes {
			if node.ObjectMeta.Name == "ncn-m001" {
				ncnM001Found = true
				break
			}
		}
		// If we did not find ncn-m001 in the list of k8s nodes, then we want to remove it from
		// our lists of remote NCNs to test
		removeNcnM001 = !ncnM001Found
	}

	// Now create our list of remote NCNs to test by going through the list from /etc/hosts and removing:
	// 1. Our local hostnaame
	// 2. ncn-m001, if removeNcnM001 is true
	for _, ncn := range etcHostsNcns {
		if ncn == myHostname {
			continue
		} else if removeNcnM001 && ncn == "ncn-m001" {
			continue
		}
		ncns = append(ncns, ncn)
	}
	common.Debugf("List of remote NCNs to test: %s", strings.Join(ncns, " "))

	return
}

// We run our systemctl command on the local host and on all other NCNs, validating
// that our expected success string shows up in the output.
func verifyCfsStateReporterOnNcns() (ok bool) {
	// Let's be optimistic! ok starts out as true
	ok = true

	// First, let's check the cfs-state-reporter status on this host
	common.Infof("Checking status of cfs-state-reporter on this host")
	_, err := common.RunPath(cfsStateReporterCheckPath)
	if err != nil {
		common.Errorf("Command failed or failed to run: \"%s\": %v", cfsStateReporterCheckPath, err)
		ok = false
	} else {
		common.Infof("cfs-state-reporter status looks good on this host")
	}

	common.Infof("Checking status of cfs-state-reporter on other NCNs")

	common.Debugf("Getting list of other NCNs")
	ncnList, err := getNcns()
	if err != nil {
		common.Error(err)
		// Nothing more to do if we cannot get the list
		ok = false
		return
	} else if len(ncnList) == 0 {
		common.Errorf("No other NCNs found in /etc/hosts -- this should never be the case")
		ok = false
		return
	}

	for _, ncnName := range ncnList {
		if !checkCfsStateReporterOnRemoteHost(ncnName) {
			ok = false
		}
	}

	if ok {
		common.Infof("No errors found with cfs-state-reporter status")
	} else {
		// No need to log this as an error here, since we already logged it earlier
		common.Infof("At least one error found with cfs-state-reporter status")
	}
	return
}
