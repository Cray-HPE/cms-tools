// MIT License
//
// (C) Copyright 2025 Hewlett Packard Enterprise Development LP
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

/*
 * bos_sessions_cli.go
 *
 * BOS sessions helper function
 *
 */
package bos

import (
	"encoding/json"
	"strconv"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func CreateBOSSessionCLI(staged bool, sessionName, templateName, cliVersion string) (sessionRecord BOSSession, ok bool) {
	common.Infof("Creating session template %s in BOS via CLI", sessionName)
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessions", "create", "--name", sessionName, "--stage", strconv.FormatBool(staged),
		"--template-name", templateName, "--operation", "reboot", "--limit", "fakexname"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &sessionRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetBOSSessionRecordCLI(sessionName, cliVersion string) (sessionRecord BOSSession, ok bool) {
	common.Infof("Getting session %s in BOS via CLI", sessionName)
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessions", "describe", sessionName); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &sessionRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetBOSSessionRecordsCLI(cliVersion string) (sessionRecords []BOSSession, ok bool) {
	common.Infof("Getting all sessions in BOS via CLI")
	if cmdOut := RunVersionedBOSCommand(cliVersion, "sessions", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &sessionRecords); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func DeleteBOSSessionCLI(sessionName, cliVersion string) (ok bool) {
	common.Infof("Deleting session %s in BOS via CLI", sessionName)
	return RunVersionedBOSCommand(cliVersion, "sessions", "delete", sessionName) != nil
}
