// MIT License
//
// (C) Copyright 2020-2023 Hewlett Packard Enterprise Development LP
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
 * tftp_get.go
 *
 * Tests a tftp get from all advertised IPs
 * Tests retrieving both the default iPXE binary and the debug iPXE binary.
 *
 */

import (
	"fmt"
	"github.com/pin/tftp"
	"os"
	"os/exec"
	"regexp"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"strings"
)

const IpxeConfigMapName = "cray-ipxe-settings"
const IpxeConfigMapSettingsFieldName = "settings.yaml"
const IpxeContainerName string = "cray-ipxe"
const IpxePathname = "/shared_tftp"

// Names of fields in settings configmap
const BinaryNameField = "cray_ipxe_binary_name"
const BinaryNameActiveField = "cray_ipxe_binary_name_active"
const DebugBinaryNameField = "cray_ipxe_debug_binary_name"
const DebugBinaryNameActiveField = "cray_ipxe_debug_binary_name_active"
const BuildX86Field = "cray_ipxe_build_x86"

// Default values of unset fields in settings configmap
const DefaultIpxeBasename = "ipxe.efi"
const DefaultIpxeDebugBasename = "debug-ipxe.efi"
const DefaultBuildx86 = false

type IpxeSettings struct {
	BinaryName, BinaryNameActive, DebugBinaryName, DebugBinaryNameActive string
	BuildX86                                                             bool
}

// Retrieve the cray-ipxe-settings Kubernetes ConfigMap,
// look up the data."settings.yaml" data and return it
func GetIpxeSettings() (ipxeSettings IpxeSettings, ok bool) {
	ok = false

	// Retrieve settings data field from the config map
	common.Infof("Retrieving iPXE settings from Kubernetes %s configmap", IpxeConfigMapName)
	ipxeSettingsBytes, err := k8s.GetConfigMapDataField(common.NAMESPACE, IpxeConfigMapName,
		IpxeConfigMapSettingsFieldName)
	if err != nil {
		common.Error(err)
		return
	}
	ipxeSettingsMap, err := common.DecodeYAMLIntoStringMap(ipxeSettingsBytes)
	if err != nil {
		common.Error(err)
		return
	}
	ipxeSettings.BinaryName, err = common.GetStringFieldFromMapObjectWithDefault(BinaryNameField, ipxeSettingsMap,
		DefaultIpxeBasename)
	if err != nil {
		common.Error(err)
		return
	}
	ipxeSettings.BinaryNameActive, err = common.GetStringFieldFromMapObjectWithDefault(BinaryNameActiveField,
		ipxeSettingsMap, "")
	if err != nil {
		common.Error(err)
		return
	}
	ipxeSettings.DebugBinaryName, err = common.GetStringFieldFromMapObjectWithDefault(DebugBinaryNameField,
		ipxeSettingsMap,
		DefaultIpxeDebugBasename)
	if err != nil {
		common.Error(err)
		return
	}
	ipxeSettings.DebugBinaryNameActive, err = common.GetStringFieldFromMapObjectWithDefault(DebugBinaryNameActiveField,
		ipxeSettingsMap, "")
	if err != nil {
		common.Error(err)
		return
	}
	ipxeSettings.BuildX86, err = common.GetBoolFieldFromMapObjectWithDefault(BuildX86Field, ipxeSettingsMap,
		DefaultBuildx86)
	if err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

// Retrieve the iPXE settings from the Kubernetes configmap. Use them to determine the names of the iPXE binaries
// being generated.
func GetIpxeBinaryNames() (ipxeBasenames []string, passed bool) {
	var settings IpxeSettings
	passed = true

	settings, passed = GetIpxeSettings()
	if !passed {
		return
	}

	// Based on the settings, determine the list of iPXE binaries to test
	ipxeBasenames = make([]string, 0, 4)

	if settings.BuildX86 {
		if len(settings.BinaryNameActive) > 0 {
			ipxeBasenames = append(ipxeBasenames, settings.BinaryNameActive)
			if len(settings.BinaryName) > 0 && settings.BinaryName != settings.BinaryNameActive {
				// This means that the next rebuild of this binary will change its name.
				common.Warnf("%s and %s are set to different values in the %s configmap. This should be temporary but"+
					" could cause a test failure. Re-run the test if it fails.", BinaryNameField,
					BinaryNameActiveField, IpxeConfigMapName)
			}
		} else if len(settings.BinaryName) > 0 {
			ipxeBasenames = append(ipxeBasenames, settings.BinaryName)
			// This implies that the binaries may not have been generated yet, since the Active field should be
			// populated when they are generated
			common.Warnf("%s is set but %s is unset in the %s configmap, implying the binary has not been generated "+
				"yet. This may cause a test failure but should be temporary. Re-run the test if it fails.",
				BinaryNameField, BinaryNameActiveField, IpxeConfigMapName)
		}
		if len(settings.DebugBinaryNameActive) > 0 {
			ipxeBasenames = append(ipxeBasenames, settings.DebugBinaryNameActive)
			if len(settings.DebugBinaryName) > 0 && settings.DebugBinaryName != settings.DebugBinaryNameActive {
				// This means that the next rebuild of this binary will change its name.
				common.Warnf("%s and %s are set to different values in the %s configmap. This should be temporary"+
					" but could cause a test failure. Re-run the test if it fails.", DebugBinaryNameField,
					DebugBinaryNameActiveField, IpxeConfigMapName)
			}
		} else if len(settings.DebugBinaryName) > 0 {
			ipxeBasenames = append(ipxeBasenames, settings.DebugBinaryName)
			// This implies that the binaries may not have been generated yet, since the Active field should be
			// populated when they are generated
			common.Warnf("%s is set but %s is unset in the %s configmap, implying the binary has not been generated "+
				"yet. This may cause a test failure but should be temporary. Re-run the test if it fails.",
				DebugBinaryNameField, DebugBinaryNameActiveField, IpxeConfigMapName)
		}
		if len(ipxeBasenames) == 0 {
			common.Errorf("%s is true in %s configmap, but the binary names are blank", IpxeConfigMapName,
				BuildX86Field)
			passed = false
			return
		}
	}

	if len(ipxeBasenames) == 0 {
		common.Infof("Based on the %s configmap, no iPXE binaries are being built.", IpxeConfigMapName)
		return
	}
	common.Infof("Based on the %s configmap, the following iPXE binaries are being built: %s", IpxeConfigMapName,
		strings.Join(ipxeBasenames, ", "))
	return
}

// Just a wrapper function for calling k8s.RunCommandInContainer on the ipxe container
func RunCommandInIpxeContainer(podName string, cmdStrings ...string) (string, error) {
	return k8s.RunCommandInContainer(podName, common.NAMESPACE, IpxeContainerName, cmdStrings...)
}

// Verify:
// 1) can run command on ipxe container
// 2) IpxePathname directory exists and is not empty
// If any are not true, log the error and return false. Otherwise return true.
func IpxeContainerReady(ipxePodName string) bool {
	// Can we even run a basic command in the container?
	common.Infof("Trying to run test command in %s container (%s pod, %s namespace)", IpxeContainerName, ipxePodName,
		common.NAMESPACE)
	if _, err := RunCommandInIpxeContainer(ipxePodName, "date"); err != nil {
		common.Error(err)
		common.Errorf("Unable to run even basic command in ipxe container")
		return false
	}

	// Verify the IpxePathname directory exists in the container (well, in the shared storage space, but we're all
	// friends here -- let's not split hairs)
	common.Infof("Trying to list contents of %s directory in %s container (%s pod, %s namespace)", IpxePathname,
		IpxeContainerName, ipxePodName, common.NAMESPACE)
	if _, err := RunCommandInIpxeContainer(ipxePodName, "sh", "-c", "ls "+IpxePathname+"/*"); err != nil {
		// If sh -c 'ls /shared_tftp/*' fails, it means either the directory does not exist,
		// it is not a directory, or it is an empty directory
		common.Error(err)
		common.Errorf("In the ipxe container, it seems that %s either does not exist, is not a directory, or is empty",
			IpxePathname)
		return false
	}

	return true
}

// 1) Get the port number, external IP, and cluster IP for the specified tftp service
// 2) For each IP, perform a tftp get test for each iPXE binary specified
// Return true if no errors, false otherwise
func TftpServiceFileTransferTest(serviceName, ipxePodName string, ipxeBasenames []string) (passed bool) {
	var IPPort string
	passed = true

	common.Infof("Performing TFTP file transfer test for service %s", serviceName)

	// First, get the IP addresses and port for the service
	clusterIP, externalIP, mainPort, ok := GetTftpIPsPort(serviceName)
	if !ok {
		passed = false
		if mainPort < 1 {
			common.Infof("Cannot run file transfer test without valid port number")
			return
		} else if len(clusterIP) == 0 && len(externalIP) == 0 {
			common.Infof("Cannot run file transfer test without valid IP address")
			return
		}
		common.Infof("Will run as many file transfer tests as we can, despite previous failures")
	}

	if len(ipxeBasenames) == 0 {
		common.Infof("No iPXE binaries are being built. Will not be able to perform the actual file transfer test.")
		return
	}

	for _, ipxeBasename := range ipxeBasenames {
		common.Infof("Performing file transfer tests for %s binary from service %s", ipxeBasename, serviceName)
		if len(clusterIP) > 0 {
			IPPort = fmt.Sprintf("%s:%d", clusterIP, mainPort)
			common.Infof("Testing tftp file transfer of %s from cluster IP:port (%s)", ipxeBasename, IPPort)
			if !TftpIPPortFileTransferTest(IPPort, ipxePodName, ipxeBasename) {
				passed = false
			}
		}

		if len(externalIP) > 0 {
			IPPort = fmt.Sprintf("%s:%d", externalIP, mainPort)
			common.Infof("Testing tftp file transfer of %s from external IP:port (%s)", ipxeBasename, IPPort)
			if !TftpIPPortFileTransferTest(IPPort, ipxePodName, ipxeBasename) {
				passed = false
			}
		}
	}

	return
}

// Return a file object for the specified local (e.g. wherever this test is running) file
func OpenLocalFile(name string) (*os.File, bool) {
	common.Infof("Creating local file for transfer test: %s", name)
	localFile, err := os.Create(name)
	if err != nil {
		common.Error(err)
		common.Errorf("Error creating file: %s", name)
		return nil, false
	}
	return localFile, true
}

// If the file still exists, remove it
func RemoveLocalFile(name string) {
	// If it still exists, remove it
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return
	}
	common.Infof("Removing temporary local file %s", name)
	if err := os.Remove(name); err != nil {
		common.Error(err)
		common.Errorf("Error deleting temporary file %s", name)
	}
}

// Sync and Close local file
func SyncCloseLocalFile(name string, file *os.File) (ok bool) {
	ok = true

	// First, sync data to file, in case there are errors
	// (my understanding is that the file sync may happen after
	// the file is closed, otherwise, so this is our only chance
	// to definitely detect a write error to the file)
	if err := file.Sync(); err != nil {
		common.Error(err)
		common.Errorf("Error Syncing file: %s", name)
		ok = false
	}
	if err := file.Close(); err != nil {
		common.Error(err)
		common.Errorf("Error closing file: %s", name)
		ok = false
	}
	return
}

// 1) Perform a tftp get of the ipxe.efi file using the provided tftp client
// 2) Generate the md5sum of the received file
// If no problems, return the md5sum and true. Otherwise, return an empty string and false.
func GetFileAndSum(tftpClient *tftp.Client, IPPort, ipxePodName, ipxeBasename, localFileName string,
	localFile *os.File) (string, bool) {
	common.Infof("Opening tftp receive for %s from %s", ipxeBasename, IPPort)
	writerTo, err := tftpClient.Receive(ipxeBasename, "octet")
	if err != nil {
		common.Error(err)
		common.Errorf("Error receiving file %s via tftp from %s", ipxeBasename, IPPort)
		SyncCloseLocalFile(localFileName, localFile)
		return "", false
	}
	common.Infof("Writing remote file data to %s", localFileName)
	numBytes, err := writerTo.WriteTo(localFile)
	if err != nil {
		common.Error(err)
		common.Errorf("Error with tftp transfer of %s from %s to %s", ipxeBasename, IPPort, localFileName)
		SyncCloseLocalFile(localFileName, localFile)
		return "", false
	}
	common.Infof("tftp transfer of %s from %s to %s completed (%d bytes)", ipxeBasename, IPPort, localFileName,
		numBytes)
	common.Infof("Closing local file %s", localFileName)
	if !SyncCloseLocalFile(localFileName, localFile) {
		return "", false
	}
	return GetMd5sumInIpxePod(ipxePodName, ipxeBasename)
}

// Take the output of the md5sum command, along with the filename of the target file, and extract just the md5sum
// string. If no errors, return that string and true. Otherwise return an empty string and false.
func GetMd5sumFromCmdOut(cmdOut, fileName string) (string, bool) {
	reString := fmt.Sprintf("^[0-9a-f]{32}  %s\n$", fileName)
	md5sumRe, err := regexp.Compile(reString)
	if err != nil {
		common.Error(err)
		common.Errorf("Error compiling regular expression to parse md5sum command output")
		return "", false
	} else if !md5sumRe.MatchString(cmdOut) {
		common.Errorf("Command output does not match the expected format for an md5sum command")
		return "", false
	}
	return string(cmdOut[:32]), true
}

// Call the md5sum command on the ipxe.efi file in the ipxe pod.
// Return the md5sum string for the file and true, if no errors.
// Otherwise return an empty string and false.
func GetMd5sumInIpxePod(podName, ipxeBasename string) (cksum string, ok bool) {
	ok = false
	cksum = ""
	ipxeFilename := IpxePathname + "/" + ipxeBasename

	cmdOut, err := RunCommandInIpxeContainer(podName, "md5sum", ipxeFilename)
	if err != nil {
		common.Error(err)
		common.Errorf("Unable to generate md5sum of %s file in the ipxe container", ipxeFilename)
		return
	}
	return GetMd5sumFromCmdOut(cmdOut, ipxeFilename)
}

// For the specified k8s service name, get its cluster IP, external IP, and port number.
// Return those and true if no errors. Otherwise return null values and false.
func GetTftpIPsPort(serviceName string) (clusterIP, externalIP string, port int32, ok bool) {
	ok = true
	clusterIP = ""
	externalIP = ""
	port = 0

	common.Infof("Looking up k8s service %s in namespace %s", serviceName, common.NAMESPACE)
	srv, err := k8s.GetService(common.NAMESPACE, serviceName)
	if err != nil {
		common.Error(err)
		ok = false
		return
	}

	if len(srv.Spec.ClusterIP) == 0 {
		common.Errorf("Unable to determine cluster IP for %s service in namespace %s", serviceName, common.NAMESPACE)
		ok = false
	}
	clusterIP = srv.Spec.ClusterIP
	common.Infof("Service %s, namespace %s, cluster IP = %s",
		serviceName, common.NAMESPACE, clusterIP)

	// Weirdly, when I run "kubectl get service" to show the service, it shows this as the external IP, but
	// using the golang client, the ExternalIP field is blank and the IP is stored in the LoadBalancerIP field.
	if len(srv.Spec.LoadBalancerIP) == 0 {
		common.Errorf("Unable to determine external IP for %s service in namespace %s", serviceName, common.NAMESPACE)
		return
	}
	externalIP = srv.Spec.LoadBalancerIP
	common.Infof("Service %s, namespace %s, external IP = %s", serviceName, common.NAMESPACE, externalIP)

	if len(srv.Spec.Ports) < 1 {
		common.Errorf("No ports defined for service %s in namespace %s", serviceName, common.NAMESPACE)
		return
	}
	port = srv.Spec.Ports[0].Port
	if port > 0 {
		common.Infof("Service %s, namespace %s, port = %d", serviceName, common.NAMESPACE, port)
		ok = true
	} else {
		common.Errorf("Invalid port (%d) for service %s in namespace %s", port, serviceName, common.NAMESPACE)
	}

	if len(srv.Spec.Ports) > 1 {
		common.Warnf("%d ports defined for service %s in namespace %s", len(srv.Spec.Ports), serviceName,
			common.NAMESPACE)
		for _, extraport := range srv.Spec.Ports[1:] {
			if extraport.Port > 0 {
				common.Warnf("Service %s, namespace %s, extra port = %d", serviceName, common.NAMESPACE,
					extraport.Port)
			} else {
				common.Errorf("Invalid extra port (%d) for service %s in namespace %s", extraport.Port, serviceName,
					common.NAMESPACE)
				ok = false
			}
		}
	}
	return
}

// Perform a tftp get of the ipxe.efi file from the specified IP address and port.
// Verify via md5sum that the received file matches the remote file
// If error, return false, otherwise return true.
func TftpIPPortFileTransferTest(IPPort, ipxePodName, ipxeBasename string) bool {
	var remoteSumBefore, remoteSumAfter, localSum, localFileName string
	var ok bool
	var localFile *os.File

	common.Infof("Initializing tftp client for %s", IPPort)
	tftpClient, err := tftp.NewClient(IPPort)
	if err != nil {
		common.Error(err)
		common.Errorf("Unable to initialize tftp client for %s", IPPort)
		return false
	}

	localFileName = fmt.Sprintf("/tmp/cmsdev-ipxetftp-%s-%s.tmp", common.AlnumString(6), ipxeBasename)
	localFile, ok = OpenLocalFile(localFileName)
	if !ok {
		return false
	}
	defer RemoveLocalFile(localFileName)

	// The file is periodically re-generated, meaning that if we just take the md5sum and transfer the file,
	// we may not have recorded the right sum. Therefore we must take the sum both before and after the transfer.
	// If they do not match, we must repeat the process until they do

	remoteSumBefore, ok = GetMd5sumInIpxePod(ipxePodName, ipxeBasename)
	if !ok {
		SyncCloseLocalFile(localFileName, localFile)
		return false
	}
	remoteSumAfter, ok = GetFileAndSum(tftpClient, IPPort, ipxePodName, ipxeBasename, localFileName, localFile)
	if !ok {
		return false
	}
	for remoteSumBefore != remoteSumAfter {
		common.Infof("Remote file changed around the time of our transfer. Must redo transfer.")
		localFile, ok = OpenLocalFile(localFileName)
		if !ok {
			return false
		}
		remoteSumBefore = remoteSumAfter
		remoteSumAfter, ok = GetFileAndSum(tftpClient, IPPort, ipxePodName, ipxeBasename, localFileName, localFile)
		if !ok {
			return false
		}
	}

	common.Infof("Remote file did not change around the time of our transfer, proceeding")
	// Now verify that the md5sum of the local file matches
	path, err := exec.LookPath("md5sum")
	if err != nil {
		common.Error(err)
		common.Errorf("Unable to local path to md5sum locally")
		return false
	}

	cmd := exec.Command(path, localFileName)
	common.Infof("Running command: %s", cmd)
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		common.Error(err)
		common.Errorf("Error determining md5sum of local file %s", localFileName)
		return false
	} else if len(cmdOut) == 0 {
		common.Errorf("Blank output from md5sum command. Error determining md5sum of local file %s", localFileName)
		return false
	}
	localSum, ok = GetMd5sumFromCmdOut(string(cmdOut), localFileName)
	if !ok {
		return false
	} else if localSum != remoteSumAfter {
		common.Errorf("md5sum of remote file (%s) does not match that of received file (%s)", remoteSumAfter, localSum)
		return false
	}
	common.Infof("md5sum of received file matches that of remote file")

	common.Infof("tftp receive test succeeded from %s", IPPort)
	return true
}
