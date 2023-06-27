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
)

const IpxeConfigMapName = "cray-ipxe-settings"
const IpxeConfigMapSettingsFieldName = "settings.yaml"
const IpxeContainerName string = "cray-ipxe"
const IpxePathname = "/shared_tftp"

var IpxeBinaryArchitectures = []string{
	"aarch64",
	"x86-64",
}

var IpxePodPrefixByArch = map[string]string{
	"aarch64": "cray-ipxe-aarch64-",
	"x86-64":  "cray-ipxe-x86-64-",
}

type stringField struct {
	Field, Default string
}

type boolField struct {
	Field   string
	Default bool
}

// Target is the name to use when generating the iPXE binary
// Build is whether or not to build this binary
type cmDebugBinaryName struct {
	Target stringField
	Build  boolField
}

// Build is a boolean specifying whether or not these binaries should be built
type cmArchBinaries struct {
	Build       boolField
	RegularName stringField
	BuildDebug  boolField
	DebugName   stringField
}

var cmBinaries = map[string]cmArchBinaries{
	"aarch64": {
		Build:       boolField{Field: "cray_ipxe_aarch64_enabled", Default: true},
		RegularName: stringField{Field: "cray_ipxe_aarch64_binary_name", Default: ""},
		BuildDebug:  boolField{Field: "cray_ipxe_aarch64_debug_enabled", Default: true},
		DebugName:   stringField{Field: "cray_ipxe_aarch64_debug_binary_name", Default: ""},
	},
	"x86-64": {
		Build:       boolField{Field: "cray_ipxe_build_x86", Default: true},
		RegularName: stringField{Field: "cray_ipxe_binary_name", Default: ""},
		BuildDebug:  boolField{Field: "cray_ipxe_debug_enabled", Default: true},
		DebugName:   stringField{Field: "cray_ipxe_debug_binary_name", Default: ""},
	},
}

var ipxeBinaryNames = map[string][]string{
	"aarch64": {"", ""},
	"x86-64":  {"", ""},
}

// Retrieve the cray-ipxe-settings Kubernetes ConfigMap,
// look up the data."settings.yaml" data, and use it to determine the names of the iPXE binaries
// being generated.
func GetIpxeBinaryNames() (passed bool) {
	var totalNames int
	var buildArch, buildBinary bool
	var binaryName string
	passed = false
	totalNames = 0

	// Retrieve settings data field from the config map
	common.Infof("Retrieving iPXE settings from Kubernetes %s configmap", IpxeConfigMapName)
	ipxeSettingsBytes, err := k8s.GetConfigMapDataField(common.NAMESPACE, IpxeConfigMapName, IpxeConfigMapSettingsFieldName)
	if err != nil {
		common.Error(err)
		return
	}
	ipxeSettingsMap, err := common.DecodeYAMLIntoStringMap(ipxeSettingsBytes)
	if err != nil {
		common.Error(err)
		return
	}

	passed = true

	for arch, cmArchBin := range cmBinaries {
		buildArch, err = common.GetBoolFieldFromMapObjectWithDefault(cmArchBin.Build.Field, ipxeSettingsMap, cmArchBin.Build.Default)
		if err != nil {
			common.Error(err)
			passed = false
			return
		}
		common.Debugf("buildArch = %t", buildArch)
		if !buildArch {
			common.Infof("According to the configmap, no binaries are being built for %s architecture", arch)
			continue
		}
		binaryName, err = common.GetStringFieldFromMapObjectWithDefault(cmArchBin.RegularName.Field, ipxeSettingsMap, cmArchBin.RegularName.Default)
		if err != nil {
			common.Error(err)
			passed = false
			return
		}
		common.Debugf("binaryName = %s", binaryName)
		if len(binaryName) > 0 {
			common.Debugf("Appending %s to binary list for %s architecture", binaryName, arch)
			ipxeBinaryNames[arch][0] = binaryName
			totalNames += 1
		} else {
			common.Errorf("%s is true in %s configmap, but the binary name is blank", cmArchBin.Build.Field, IpxeConfigMapName)
			passed = false
		}

		// Now check for debug binary
		buildBinary, err = common.GetBoolFieldFromMapObjectWithDefault(cmArchBin.BuildDebug.Field, ipxeSettingsMap, cmArchBin.BuildDebug.Default)
		if err != nil {
			common.Error(err)
			passed = false
			return
		}
		common.Debugf("debug buildBinary = %t", buildBinary)
		if !buildBinary {
			common.Infof("According to the configmap, no debug binaries are being built for %s architecture", arch)
			continue
		}
		binaryName, err = common.GetStringFieldFromMapObjectWithDefault(cmArchBin.DebugName.Field, ipxeSettingsMap, cmArchBin.DebugName.Default)
		if err != nil {
			common.Error(err)
			passed = false
			return
		}
		common.Debugf("debug binaryName = %s", binaryName)
		if len(binaryName) > 0 {
			common.Debugf("Appending %s to binary list for %s architecture", binaryName, arch)
			ipxeBinaryNames[arch][1] = binaryName
			totalNames += 1
		} else {
			common.Errorf("%s is true in %s configmap, but the binary name is blank", cmArchBin.BuildDebug.Field, IpxeConfigMapName)
			passed = false
		}
	}

	if totalNames == 0 {
		common.Infof("Based on the %s configmap, no iPXE binaries are being built.", IpxeConfigMapName)
	}
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
	// friends here -- let's not split hairs) and that it is not empty
	common.Infof("Checking to see if directory %s in %s container (%s pod, %s namespace) exists and is not empty", IpxePathname,
		IpxeContainerName, ipxePodName, common.NAMESPACE)
	if cmdOut, err := RunCommandInIpxeContainer(ipxePodName, "sh", "-c", "[ -d "+IpxePathname+" ] && find "+IpxePathname+"  -maxdepth 0 -type d -empty"); err != nil {
		// If this command fails, then it means either the directory does not exist or it is not a directory.
		common.Error(err)
		common.Errorf("In the %s container (%s pod, %s namespace), it seems that %s either does not exist or is not a directory", IpxeContainerName, ipxePodName,
			common.NAMESPACE, IpxePathname)
		return false
	} else if cmdOut != "" {
		// The command should give no output unless the directory is empty
		common.Errorf("In the %s container (%s pod, %s namespace), it seems that %s is a directory but is empty", IpxeContainerName, ipxePodName,
			common.NAMESPACE, IpxePathname)
		return false
	}
	return true
}

// 1) Get the port number, external IP, and cluster IP for the specified tftp service
// 2) For each IP, perform a tftp get test for each iPXE binary specified
// Return true if no errors, false otherwise
func TftpServiceFileTransferTest(serviceName string, ipxePodNameByArch map[string]string) (passed bool) {

	var IPPort string
	var totalNames int
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

	for _, arch := range IpxeBinaryArchitectures {

		totalNames = 0
		for _, ipxeBinaryName := range ipxeBinaryNames[arch] {
			if ipxeBinaryName != "" {
				totalNames += 1
			}
		}
		if totalNames == 0 {
			common.Infof("No %s iPXE binaries are being built. Will not be able to perform the actual file transfer test for this architecture.", arch)
			continue
		}

		for _, ipxeBinaryName := range ipxeBinaryNames[arch] {
			if ipxeBinaryName == "" {
				continue
			}
			common.Infof("Performing file transfer tests for %s %s binary from service %s", arch, ipxeBinaryName, serviceName)
			if len(clusterIP) > 0 {
				IPPort = fmt.Sprintf("%s:%d", clusterIP, mainPort)
				common.Infof("Testing tftp file transfer of %s from cluster IP:port (%s)", ipxeBinaryName, IPPort)
				if !TftpIPPortFileTransferTest(IPPort, ipxePodNameByArch[arch], ipxeBinaryName) {
					passed = false
				}
			}

			if len(externalIP) > 0 {
				IPPort = fmt.Sprintf("%s:%d", externalIP, mainPort)
				common.Infof("Testing tftp file transfer of %s from external IP:port (%s)", ipxeBinaryName, IPPort)
				if !TftpIPPortFileTransferTest(IPPort, ipxePodNameByArch[arch], ipxeBinaryName) {
					passed = false
				}
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
func GetFileAndSum(tftpClient *tftp.Client, IPPort, ipxePodName, ipxeBinaryName, localFileName string,
	localFile *os.File) (string, bool) {
	common.Infof("Opening tftp receive for %s from %s", ipxeBinaryName, IPPort)
	writerTo, err := tftpClient.Receive(ipxeBinaryName, "octet")
	if err != nil {
		common.Error(err)
		common.Errorf("Error receiving file %s via tftp from %s", ipxeBinaryName, IPPort)
		SyncCloseLocalFile(localFileName, localFile)
		return "", false
	}
	common.Infof("Writing remote file data to %s", localFileName)
	numBytes, err := writerTo.WriteTo(localFile)
	if err != nil {
		common.Error(err)
		common.Errorf("Error with tftp transfer of %s from %s to %s", ipxeBinaryName, IPPort, localFileName)
		SyncCloseLocalFile(localFileName, localFile)
		return "", false
	}
	common.Infof("tftp transfer of %s from %s to %s completed (%d bytes)", ipxeBinaryName, IPPort, localFileName,
		numBytes)
	common.Infof("Closing local file %s", localFileName)
	if !SyncCloseLocalFile(localFileName, localFile) {
		return "", false
	}
	return GetMd5sumInIpxePod(ipxePodName, ipxeBinaryName)
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
func GetMd5sumInIpxePod(podName, ipxeBinaryName string) (cksum string, ok bool) {
	ok = false
	cksum = ""
	ipxeFilename := IpxePathname + "/" + ipxeBinaryName

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
func TftpIPPortFileTransferTest(IPPort, ipxePodName, ipxeBinaryName string) bool {
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

	localFileName = fmt.Sprintf("/tmp/cmsdev-ipxetftp-%s-%s.tmp", common.AlnumString(6), ipxeBinaryName)
	localFile, ok = OpenLocalFile(localFileName)
	if !ok {
		return false
	}
	defer RemoveLocalFile(localFileName)

	// The file is periodically re-generated, meaning that if we just take the md5sum and transfer the file,
	// we may not have recorded the right sum. Therefore we must take the sum both before and after the transfer.
	// If they do not match, we must repeat the process until they do

	remoteSumBefore, ok = GetMd5sumInIpxePod(ipxePodName, ipxeBinaryName)
	if !ok {
		SyncCloseLocalFile(localFileName, localFile)
		return false
	}
	remoteSumAfter, ok = GetFileAndSum(tftpClient, IPPort, ipxePodName, ipxeBinaryName, localFileName, localFile)
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
		remoteSumAfter, ok = GetFileAndSum(tftpClient, IPPort, ipxePodName, ipxeBinaryName, localFileName, localFile)
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
