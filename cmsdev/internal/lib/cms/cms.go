package cms

/*
 * cms.go
 * 
 * cms command functions  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	c "github.com/fatih/color"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
   	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

// struct to hold CMS services data
type serviceData struct{
	serviceAPIName string
	serviceContainerName string
	podNames []string
	pvcNames []string
}

// variable that used to load current CMS services data
var cmsServiceData = make(map[string]*serviceData)

// get cms service names
func loadCMSServiceData() {
	for _, each := range common.CMSServices {
		switch each {
		case "bos":	
			podNames, _  := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["bos"])
			if len(podNames) != 0 {
				cmsServiceData["bos"] = &serviceData { 
					serviceAPIName: podNames[0],
					serviceContainerName: "cray-bos",
					podNames: podNames,
				}
			}
			pvcNames, _  := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["bosPvc"])
			cmsServiceData["bos"].pvcNames = pvcNames
		case "cfs":	
			podNames, _  := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["cfsServices"])
			if len(podNames) != 0 {
				cmsServiceData["cfs"] = &serviceData { 
					podNames: podNames,
				}
				// find cfs API pod name
				for i, _ := range podNames {
					if strings.HasPrefix(podNames[i], common.PodServiceNamePrefixes["cfs-api"]) {
						cmsServiceData["cfs"].serviceAPIName = podNames[i]
						break
					}
				}
			}
		case "conman":	
			podNames, _  := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["conman"])
			if len(podNames) != 0 {
				cmsServiceData["conman"] = &serviceData { 
					serviceAPIName: podNames[0],
					podNames: podNames,
				}
			}
			pvcNames, _  := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["conmanPvc"])
			cmsServiceData["conman"].pvcNames = pvcNames
		case "ims":	
			podNames, _  := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["ims"])
			if len(podNames) != 0 {
				cmsServiceData["ims"] = &serviceData { 
					serviceAPIName: podNames[0],
					podNames: podNames,
				}
			}
			pvcNames, _  := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["imsPvc"])
			cmsServiceData["ims"].pvcNames = pvcNames
		case "ipxe":	
			podNames, _  := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["ipxe"])
			if len(podNames) != 0 {
				cmsServiceData["ipxe"] = &serviceData { 
					serviceAPIName: podNames[0],
					podNames: podNames,
				}
			}
		case "tftp":	
			podNames, _  := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["tftp"])
			if len(podNames) != 0 {
				cmsServiceData["tftp"] = &serviceData { 
					serviceAPIName: podNames[0],
					podNames: podNames,
				}
			}
			pvcNames, _  := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["tftpPvc"])
			cmsServiceData["tftp"].pvcNames = pvcNames
		case "vcs":	
			podNames, _  := k8s.GetPodNames(common.NAMESPACE, common.PodServiceNamePrefixes["vcs"])
			if len(podNames) != 0 {
				cmsServiceData["vcs"] = &serviceData { 
					serviceAPIName: podNames[0],
					podNames: podNames,
				}
			}
			pvcNames, _  := k8s.GetPVCNames(common.NAMESPACE, common.PodServiceNamePrefixes["vcs"])
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
	keys, _ :=	GetCMSServiceNames()
	for _, v := range keys {
		fmt.Printf("%-65s%7s\n", "-----------------------------------------------------------", "------")
		fmt.Printf("%-65s%7s\n", string(v), "STATUS")
		fmt.Printf("%-65s%7s\n", "-----------------------------------------------------------", "------")
		for i, _ := range cmsServiceData[v].podNames {
			status, _ := k8s.GetPodStatus(common.NAMESPACE, cmsServiceData[v].podNames[i])
			fmt.Printf("%-65s ", cmsServiceData[v].podNames[i])
			if status != "Running" {
				c.Red("Failed")
			} else {
				c.HiGreen("OK")
			}
		}
		for i, _ := range cmsServiceData[v].pvcNames {
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
	if ! installed {
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

