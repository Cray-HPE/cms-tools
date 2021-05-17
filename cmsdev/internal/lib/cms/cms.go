package cms

/*
 * cms.go
 *
 * cms command functions
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
	"errors"
	"fmt"
	c "github.com/fatih/color"
	"sort"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"strings"
)

// struct to hold CMS services data
type serviceData struct {
	serviceAPIName       string
	serviceContainerName string
	podNames             []string
	pvcNames             []string
}

// variable that used to load current CMS services data
var cmsServiceData = make(map[string]*serviceData)

// get cms service names
func loadCMSServiceData() {
	for _, each := range common.CMSServices {
		switch each {
		case "bos":
			podNames, _ := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["bos"])
			if len(podNames) != 0 {
				cmsServiceData["bos"] = &serviceData{
					serviceAPIName:       podNames[0],
					serviceContainerName: "cray-bos",
					podNames:             podNames,
				}
			}
			pvcNames, _ := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["bosPvc"])
			cmsServiceData["bos"].pvcNames = pvcNames
		case "cfs":
			podNames, _ := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["cfsServices"])
			if len(podNames) != 0 {
				cmsServiceData["cfs"] = &serviceData{
					podNames: podNames,
				}
				// find cfs API pod name
				for i := range podNames {
					if strings.HasPrefix(podNames[i], common.PodServiceNamePrefixes["cfs-api"]) {
						cmsServiceData["cfs"].serviceAPIName = podNames[i]
						break
					}
				}
			}
		case "conman":
			podNames, _ := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["conman"])
			if len(podNames) != 0 {
				cmsServiceData["conman"] = &serviceData{
					serviceAPIName: podNames[0],
					podNames:       podNames,
				}
			}
			pvcNames, _ := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["conmanPvc"])
			cmsServiceData["conman"].pvcNames = pvcNames
		case "ims":
			podNames, _ := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["ims"])
			if len(podNames) != 0 {
				cmsServiceData["ims"] = &serviceData{
					serviceAPIName: podNames[0],
					podNames:       podNames,
				}
			}
			pvcNames, _ := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["imsPvc"])
			cmsServiceData["ims"].pvcNames = pvcNames
		case "ipxe":
			podNames, _ := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["ipxe"])
			if len(podNames) != 0 {
				cmsServiceData["ipxe"] = &serviceData{
					serviceAPIName: podNames[0],
					podNames:       podNames,
				}
			}
		case "tftp":
			podNames, _ := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["tftp"])
			if len(podNames) != 0 {
				cmsServiceData["tftp"] = &serviceData{
					serviceAPIName: podNames[0],
					podNames:       podNames,
				}
			}
			pvcNames, _ := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["tftpPvc"])
			cmsServiceData["tftp"].pvcNames = pvcNames
		case "vcs":
			podNames, _ := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["vcs"])
			if len(podNames) != 0 {
				cmsServiceData["vcs"] = &serviceData{
					serviceAPIName: podNames[0],
					podNames:       podNames,
				}
			}
			pvcNames, _ := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["vcs"])
			cmsServiceData["tftp"].pvcNames = pvcNames
		}
	}
}

// GetCMSServiceName returns kubernetes pod name given prefix
func GetCMSServiceName(key string) string {
	loadCMSServiceData()
	return cmsServiceData[key].serviceAPIName
}

// list CMS services names
func ListServicesNames() {
	fmt.Println(strings.Join(common.CMSServices, " "))
	return
}

// retrieve names and count of services that are currently running
func GetCMSServiceNames() (keys []string, numServices int) {
	loadCMSServiceData()
	keys = make([]string, 0, len(cmsServiceData))
	for k := range cmsServiceData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	numServices = len(keys)
	return
}

// routine to list of all cms services
// TODO: make output prettier, update to consider pod phase
func ListServices() {
	loadCMSServiceData()
	keys, _ := GetCMSServiceNames()
	for _, v := range keys {
		fmt.Printf("%-65s%7s\n", "-----------------------------------------------------------", "------")
		fmt.Printf("%-65s%7s\n", string(v), "STATUS")
		fmt.Printf("%-65s%7s\n", "-----------------------------------------------------------", "------")
		for i := range cmsServiceData[v].podNames {
			status, _ := k8s.GetPodStatus(common.NAMESPACE, cmsServiceData[v].podNames[i])
			fmt.Printf("%-65s ", cmsServiceData[v].podNames[i])
			if status != "Running" {
				c.Red("Failed")
			} else {
				c.HiGreen("OK")
			}
		}
		for i := range cmsServiceData[v].pvcNames {
			status, _ := k8s.GetPVCStatus(common.NAMESPACE, cmsServiceData[v].pvcNames[i])
			fmt.Printf("%-65s ", cmsServiceData[v].pvcNames[i])
			if status != "Bound" {
				c.Red("Failed")
			} else {
				c.HiGreen("OK")
			}
		}

		fmt.Println()
	}
}

// get services logs
func PrintServiceLogs(service string) error {
	serviceNames, numServices := GetCMSServiceNames()
	if numServices < 1 {
		return errors.New("no cms services currently running")
	}
	installed := common.StringInArray(service, serviceNames)
	if !installed {
		err := fmt.Errorf("%s, is not running\n", service)
		return err
	}
	logs, err := k8s.GetPodLogs(
		common.NAMESPACE,
		cmsServiceData[service].serviceAPIName,
		cmsServiceData[service].serviceContainerName,
	)
	if err == nil {
		fmt.Println(logs)
	}
	return err
}
