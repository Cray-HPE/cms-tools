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
package ims

/*
 * ims_public_key_cli.go
 *
 * ims public key cli functions
 *
 */
import (
	"encoding/json"
	"os"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func CreateIMSPublicKeyRecordCLI(publicKeyName string) (publicKeyRecord IMSPublicKeyRecord, ok bool) {
	common.Infof("Creating public key %s in IMS via CLI", publicKeyName)
	// Get the home directory of the current user
	homeDir, err := os.UserHomeDir()
	if err != nil {
		common.Error(err)
		return
	}

	filePath := homeDir + "/.ssh/id_rsa.pub"
	if cmdOut := runCLICommand("public-keys", "create", "--name", publicKeyName, "--public-key", filePath); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &publicKeyRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetDeletedIMSPublicKeyRecordCLI(publicKeyId string) (publicKeyRecord IMSPublicKeyRecord, ok bool) {
	common.Infof("Getting deleted public key record %s in IMS via CLI", publicKeyId)
	if cmdOut := runCLICommand("deleted", "public-keys", "describe", publicKeyId); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &publicKeyRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetDeletedIMSPublicKeyRecordCLIs() (publicKeyRecords []IMSPublicKeyRecord, ok bool) {
	common.Infof("Getting list of all deleted public key records in IMS via CLI")
	if cmdOut := runCLICommand("deleted", "public-keys", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &publicKeyRecords); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func DeleteIMSPublicKeyRecordCLI(publicKeyId string) (ok bool) {
	common.Infof("Deleting public key %s in IMS via CLI", publicKeyId)
	return runCLICommand("public-keys", "delete", publicKeyId) != nil
}

func UndeleteIMSPublicKeyRecordCLI(publicKeyId string) (ok bool) {
	common.Infof("Restoring public key %s in IMS via CLI", publicKeyId)
	return runCLICommand("deleted", "public-keys", "update", publicKeyId, "--operation", "undelete") != nil
}

func PermanentDeleteIMSPublicKeyRecordCLI(publicKeyId string) (ok bool) {
	common.Infof("Hard deleting public key %s in IMS via CLI", publicKeyId)
	return runCLICommand("deleted", "public-keys", "delete", publicKeyId) != nil
}
