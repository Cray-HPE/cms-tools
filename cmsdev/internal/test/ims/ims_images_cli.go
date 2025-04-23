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
 * ims_images_cli.go
 *
 * ims images cli functions
 *
 */
import (
	"encoding/json"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func GetDeletedIMSImageRecordCLI(imageId string) (imageRecord IMSImageRecord, ok bool) {
	common.Infof("Getting deleted image record %s in IMS via CLI", imageId)
	if cmdOut := runCLICommand("deleted", "images", "describe", imageId); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &imageRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetDeletedIMSImageRecordsCLI() (imageRecords []IMSImageRecord, ok bool) {
	common.Infof("Getting all deleted image records in IMS via CLI")
	if cmdOut := runCLICommand("deleted", "images", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &imageRecords); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func CreateIMSImageRecordCLI(imageName string) (imageRecord IMSImageRecord, ok bool) {
	common.Infof("Creating image %s in IMS via CLI", imageName)
	if cmdOut := runCLICommand("images", "create", "--name", imageName,
		"--metadata-key", "name", "--metadata-value", imageName); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &imageRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func UpdateIMSImageRecordCLI(imageId string, arch string) (imageRecord IMSImageRecord, ok bool) {
	common.Infof("Updating image %s with arch %s", imageId, arch)
	if cmdOut := runCLICommand("images", "update", imageId, "--arch", arch, "--metadata-operation", "set",
		"--metadata-key", "project", "--metadata-value", "csm"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &imageRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func DeleteIMSImageRecordCLI(imageId string) (ok bool) {
	common.Infof("Soft deleting image %s", imageId)
	return runCLICommand("images", "delete", imageId) != nil
}

func UndeleteIMSImageRecordCLI(imageId string) (ok bool) {
	common.Infof("Restoring image %s", imageId)
	return runCLICommand("deleted", "images", "update", imageId, "--operation", "undelete") != nil
}

func PermanentDeleteIMSImageRecordCLI(imageId string) (ok bool) {
	common.Infof("Permanently deleting image %s", imageId)
	return runCLICommand("deleted", "images", "delete", imageId) != nil
}
