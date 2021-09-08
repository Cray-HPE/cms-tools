package cfs

/*
 * cfs_state_reporter.go
 *
 * Verify status of cfs-state-reporter on other NCNs
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
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
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"io/ioutil"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"strings"
)

const systemctlCommandPath = "/usr/bin/systemctl"
const systemctlCommandArgs = "--no-pager status cfs-state-reporter"

var systemctlCommandString = systemctlCommandPath + " " + systemctlCommandArgs

const systemctlSuccessString = "(code=exited, status=0/SUCCESS)"

// We assume the local and remote user are root, and that the local user's
// ssh keys and known_hosts files are in the default locations.
const remoteUser = "root"
const localSshDir = "/root/.ssh"

var localKeyFile = localSshDir + "/" + "id_rsa"
var localKnownHostsFile = localSshDir + "/" + "known_hosts"

// We assume the default ssh port on the remote host
const remoteSshPort = 22

// Generate the ssh.ClientConfig object to use with our subsequent ssh sessions.
// It will use public key authentication.
func getSSHConfig() (config *ssh.ClientConfig, err error) {
	common.Infof("Generating ssh client configuration")
	key, err := ioutil.ReadFile(localKeyFile)
	if err != nil {
		common.Debugf("ioutil.ReadFile(\"%s\") returned error: %v", localKeyFile, err)
		err = fmt.Errorf("unable to read private key: %v", err)
		return
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		common.Debugf("ssh.ParsePrivateKey returned error: %v", err)
		err = fmt.Errorf("unable to parse private key: %v", err)
		return
	}

	hostKeyCallback, err := knownhosts.New(localKnownHostsFile)
	if err != nil {
		common.Debugf("knownhosts.New(\"%s\") returned error: %v", localKnownHostsFile, err)
		return
	}

	config = &ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}
	common.Infof("Successfully Generated ssh client configuration")
	return
}

// Return the ssh.Client object needed to create ssh Sessions to the specified remoteHost.
func getSSHClient(remoteHost string, config *ssh.ClientConfig) (client ssh.Client, err error) {
	remoteHostPort := fmt.Sprintf("%s:%d", remoteHost, remoteSshPort)
	common.Debugf("Creating ssh client for %s", remoteHostPort)
	client, err = ssh.Dial("tcp", remoteHostPort, config)
	if err != nil {
		common.Debugf("ssh.Dial returned error: %v", err)
		err = fmt.Errorf("Failed to dial: ", err)
	}
	common.Debugf("Successfully created ssh client for %s", remoteHostPort)
	return
}

// Return the ssh.Session object needed to run commands on the specified client
func getSSHSession(client ssh.Client) (session ssh.Session, err error) {
	common.Debugf("Creating ssh session for client %v", client)
	session, err := client.NewSession()
	if err != nil {
		common.Debugf("client.NewSession returned error: %v", err)
		err = fmt.Errorf("Failed to create session: ", err)
	}
	common.Debugf("Successfully created ssh session for client %v", client)
	return
}

// Run the specified remote command via ssh on the specified remote host.
// If the remote command runs but returns a non-0 return code, this does not
// generate an error. It is left to the caller to decide how to deal with that.
// Errors are only returned in the case that there were problems with ssh itself.
func runRemoteCommand(remoteHost, commandString string, config *ssh.ClientConfig) (outString, errString string, rc int, err error) {
	client, err := getSSHClient(remoteHost, config)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := getSSHSession(client)
	if err != nil {
		return err
	}
	defer session.Close()

	var outBuffer, errBuffer bytes.Buffer
	session.Stdout = &outBuffer
	session.Stderr = &errBuffer
	err := session.Wait()
	outString = outBuffer.String()
	if len(outString) > 0 {
		common.Debugf("Command stdout: %s", outString)
	} else {
		common.Debugf("No stdout from command")
	}
	errString = errBytes.String()
	if len(errString) > 0 {
		common.Warnf("Command stderr: %s", errString)
	} else {
		common.Debugf("No stderr from command")
	}
	if err != nil {
		if exiterr, ok := err.(*ssh.ExitError); ok {
			// This means the remote command failed
			rc = exiterr.ExitStatus()
			common.Debugf("Command return code: %d", rc)
			// We do not have this function return an error in this case -- we will let the caller
			// decide how to handle the fact that the remote command failed
			err = nil
		} else {
			// Some other error happened running the remote command
			err = fmt.Errorf("Error trying to execute command via ssh: %v", err)
			return
		}
	} else {
		// This means the remote command passed
		common.Debugf("Command return code: 0")
		rc = 0
	}
	return
}

// Checks the systemctl command output to see if our expected success string is in there.
func validateSystemctlOutput(outString) error {
	if len(outString) == 0 {
		return fmt.Errorf("Command gave no output: %s", systemctlCommandString)
	} else if !strings.Contains(outString, successString) {
		return fmt.Errorf("Expected string not found in \"%s\" output: %s", systemctlCommandString, systemctlSuccessString)
	}
	return nil
}

// Runs our systemctl command via ssh on the remote host, and validates the output.
// If ssh errors are encountered, a warning is logged, but the function returns success.
func checkCfsStateReporterOnRemoteHost(remoteHost string, config *ssh.ClientConfig) bool {
	common.Infof("Checking cfs-state-reporter status on %s", remoteHost)
	outString, errString, rc, err := runRemoteCommand(remoteHost, systemctlCommandString, config)
	if err != nil {
		common.Warnf("Unable to check cfs-state-reporter status on %s: %v", remoteHost, err)
		// We do not consider this to be a failure
		return true
	}
	err = validateSystemctlOutput(outString)
	if err != nil {
		common.Errorf("cfs-state-reporter check failed on %s: %v", remoteHost, err)
		return false
	}
	common.Infof("cfs-state-reporter status looks good on %s", remoteHost)
	return true
}

// Extract from /etc/hosts the list of all master and worker NCNs, except for the current host
func getMastersAndWorkers() (ncns []string, err error) {
	result, err := common.RunPath("/bin/bash", "-c",
		"grep -o -E \"ncn-[mw][0-9][0-9][0-9]([[:space:]]|$)\" /etc/hosts | grep -Ev \"^$HOSTNAME$\"")
	if err != nil {
		err = fmt.Errorf("Error getting NCN hostnames from /etc/hosts: %v", err)
		return
	} else if result.Rc != 0 {
		err = fmt.Errorf("Error getting NCN hostnames from /etc/hosts: bash command returned return code %d", result.Rc)
		return
	}
	ncns = strings.Fields(result.OutBytes.String())
	return
}

// We run our systemctl command on the local host and on all other master and worker nodes, validating
// that our expected success string shows up in the output.
func verifyCfsStateReporterOnMasterAndWorkers() (ok bool) {
	// Let's be optimistic! ok starts out as true
	ok = true

	// First, let's check the cfs-state-reporter status on this host
	common.Infof("Checking status of cfs-state-reporter on this host")
	cmdResult, err := common.RunPath(systemctlCommandPath, strings.Fields(systemctlCommandArgs)...)
	if err != nil {
		common.Errorf("Error running \"%s\": %v", systemctlCommandString, err)
		ok = false
	} else {
		err = validateSystemctlOutput(outString)
		if err != nil {
			common.Errorf("cfs-state-reporter check failed on this host: %v", err)
			ok = false
		} else {
			common.Infof("cfs-state-reporter status looks good on this host")
		}
	}

	common.Infof("Checking status of cfs-state-reporter on other master and worker NCNs")

	common.Debugf("Getting list of other master and worker NCNs from /etc/hosts")
	ncnList, err := getMastersAndWorkers()
	if err != nil {
		common.Error(err)
		// Nothing more to do if we cannot get the list
		ok = false
		return
	} else if len(ncnList) == 0 {
		common.Errorf("No other master or worker NCNs found in /etc/hosts -- this should never be the case")
		ok = false
		return
	}

	// Get ssh configuration
	config, err := getSSHConfig()
	if err != nil {
		common.Warnf("Unable to generate ssh configuration: %v", err)
		common.Warnf("Unable to verify cfs-state-reporter status on other master and worker NCNs")
		// Nothing more we can do without the ssh configuration
		return
	}

	for _, ncnName := range ncnList {
		if !checkCfsStateReporterOnRemoteHost(ncnName, config) {
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
