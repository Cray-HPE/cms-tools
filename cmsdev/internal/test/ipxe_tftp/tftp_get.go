//
//  MIT License
//
//  (C) Copyright 2020-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//
package ipxe_tftp

/*
 * tftp_get.go
 *
 * Tests a tftp get from all advertised IPs
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

const ipxePathname = "/shared_tftp"
const ipxeBasename = "ipxe.efi"
const ipxeFilename = ipxePathname + "/" + ipxeBasename
const ipxeContainerName string = "cray-ipxe"

// Just a wrapper function for calling k8s.RunCommandInContainer on the ipxe container
func RunCommandInIpxeContainer(podName string, cmdStrings ...string) (string, error) {
	return k8s.RunCommandInContainer(podName, common.NAMESPACE, ipxeContainerName, cmdStrings...)
}

// Verify:
// 1) can run command on ipxe container
// 2) ipxe.efi file exists
// 3) can get md5sum of file
// If any are not true, log the error and return false. Otherwise return true.
func IpxeContainerReady(ipxePodName string) bool {
	// Can we even run a basic command in the container?
	common.Infof("Trying to run test command in ipxe container")
	if _, err := RunCommandInIpxeContainer(ipxePodName, "date"); err != nil {
		common.Error(err)
		common.Errorf("Unable to run even basic command in ipxe container")
		return false
	}

	// Verify the ipxe.efi file exists in the container (well, in the shared storage space, but
	// we're all friends here -- let's not split hairs)
	common.Infof("Trying to list %s file in ipxe container", ipxeFilename)
	if _, err := RunCommandInIpxeContainer(ipxePodName, "ls", "-al", ipxeFilename); err != nil {
		common.Error(err)
		common.Errorf("%s file does not appear to exist in the ipxe container", ipxeFilename)
		return false
	}

	// Finally, let's make sure that the md5sum of the file can be generated
	common.Infof("Trying to generate md5sum of %s file in ipxe container", ipxeFilename)
	if _, ok := GetMd5sumInIpxePod(ipxePodName); !ok {
		return false
	}
	return true
}

// 1) Get the port number, external IP, and cluster IP for the specified tftp service
// 2) For each IP, perform a tftp get test
// Return true if no errors, false otherwise
func TftpServiceFileTransferTest(serviceName, ipxePodName string) (passed bool) {
	var IPPort string
	passed = true

	// First, get the IP addresseses and port for the service
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
		common.Infof("Will run as much of the file transfer test as we can, despite previous failures")
	}

	if len(clusterIP) > 0 {
		IPPort = fmt.Sprintf("%s:%d", clusterIP, mainPort)
		common.Infof("Testing tftp file transfer from cluster IP:port (%s)", IPPort)
		if !TftpIPPortFileTransferTest(IPPort, ipxePodName) {
			passed = false
		}
	}

	if len(externalIP) > 0 {
		IPPort = fmt.Sprintf("%s:%d", externalIP, mainPort)
		common.Infof("Testing tftp file transfer from external IP:port (%s)", IPPort)
		if !TftpIPPortFileTransferTest(IPPort, ipxePodName) {
			passed = false
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
func GetFileAndSum(tftpClient *tftp.Client, IPPort, ipxePodName, localFileName string, localFile *os.File) (string, bool) {
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
	common.Infof("tftp transfer of %s from %s to %s completed (%d bytes)", ipxeBasename, IPPort, localFileName, numBytes)
	common.Infof("Closing local file %s", localFileName)
	if !SyncCloseLocalFile(localFileName, localFile) {
		return "", false
	}
	return GetMd5sumInIpxePod(ipxePodName)
}

// Take the output of the md5sum command, along with the filename of the target file,
// and extract just the md5sum string. If no errors, return that string and true. Otherwise
// return an empty string and false.
func GetMd5sumFromCmdOut(cmdOut, fileName string) (string, bool) {
	reString := fmt.Sprintf("^[0-9a-f]{32}  %s\n$", fileName)
	md5sumRe, err := regexp.Compile(reString)
	if err != nil {
		common.Error(err)
		common.Errorf("Error compiling regular expression to parse md5sum command output")
		return "", false
	} else if !md5sumRe.MatchString(cmdOut) {
		common.Errorf("The command output does not match the expected format for an md5sum command")
		return "", false
	}
	return string(cmdOut[:32]), true
}

// Call the md5sum command on the ipxe.efi file in the ipxe pod
// Return the md5sum string for the file and true, if no errors.
// Otherwise return an empty string and false.
func GetMd5sumInIpxePod(podName string) (cksum string, ok bool) {
	ok = false
	cksum = ""

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
		common.Warnf("%d ports defined for service %s in namespace %s", len(srv.Spec.Ports), serviceName, common.NAMESPACE)
		for _, extraport := range srv.Spec.Ports[1:] {
			if extraport.Port > 0 {
				common.Warnf("Service %s, namespace %s, extra port = %d", serviceName, common.NAMESPACE, extraport.Port)
			} else {
				common.Errorf("Invalid extra port (%d) for service %s in namespace %s", extraport.Port, serviceName, common.NAMESPACE)
				ok = false
			}
		}
	}
	return
}

// Perform a tftp get of the ipxe.efi file from the specified IP address and port.
// Verify via md5sum that the received file matches the remote file
// If error, return false, otherwise return true.
func TftpIPPortFileTransferTest(IPPort, ipxePodName string) bool {
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

	// The file is frequently re-generated, meaning that if we just take the md5sum and transfer the file,
	// we may not have recorded the right sum. Therefore we must take the sum both before and after the transfer.
	// If they do not match, we must repeat the process until they do

	remoteSumBefore, ok = GetMd5sumInIpxePod(ipxePodName)
	if !ok {
		SyncCloseLocalFile(localFileName, localFile)
		return false
	}
	remoteSumAfter, ok = GetFileAndSum(tftpClient, IPPort, ipxePodName, localFileName, localFile)
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
		remoteSumAfter, ok = GetFileAndSum(tftpClient, IPPort, ipxePodName, localFileName, localFile)
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
