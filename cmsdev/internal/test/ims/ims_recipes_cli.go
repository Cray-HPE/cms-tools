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

func GetDeletedIMSRecipeRecordCLI(recipeId string) (recipeRecord IMSRecipeRecord, ok bool) {
	common.Infof("Getting deleted recipe record %s in IMS via CLI", recipeId)
	if cmdOut := runCLICommand("deleted", "recipes", "describe", recipeId); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &recipeRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func GetDeletedIMSRecipeRecordsCLI() (recipeRecords []IMSRecipeRecord, ok bool) {
	common.Infof("Getting all deleted recipe records in IMS via CLI")
	if cmdOut := runCLICommand("deleted", "recipes", "list"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &recipeRecords); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func CreateIMSRecipeRecordCLI(recipeName string) (recipeRecord IMSRecipeRecord, ok bool) {
	common.Infof("Creating recipe %s in IMS via CLI", recipeName)
	if cmdOut := runCLICommand("recipes", "create", "--name",
		recipeName, "--linux-distribution", "sles15",
		"--recipe-type", "kiwi-ng", "--template-dictionary-key", "USS_VERSION,FULL_COS_BASE_VERSION",
		"--template-dictionary-value", "1.1.2-1-cos-base-3.1,3.1.2-1-sle-15.5"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &recipeRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func UpdateIMSRecipeRecordCLI(recipeId string, arch string) (recipeRecord IMSRecipeRecord, ok bool) {
	common.Infof("Updating recipe %s with arch %s", recipeId, arch)
	if cmdOut := runCLICommand("recipes", "update", recipeId, "--arch", arch,
		"--template-dictionary-key", "USS_VERSION", "--template-dictionary-value", "1.1.2-1-cos-3.1"); cmdOut != nil {
		common.Infof("Decoding JSON in command output")
		if err := json.Unmarshal(cmdOut, &recipeRecord); err == nil {
			ok = true
		} else {
			common.Error(err)
		}
	}
	return
}

func DeleteIMSRecipeRecordCLI(recipeId string) (ok bool) {
	common.Infof("Deleting recipe record %s in IMS via CLI", recipeId)
	return runCLICommand("recipes", "delete", recipeId) != nil
}

func UndeleteIMSRecipeRecordCLI(recipeId string) bool {
	common.Infof("Restoring recipe record %s in IMS via CLI", recipeId)
	return runCLICommand("deleted", "recipes", "update", recipeId, "--operation", "undelete") != nil
}

func PermanentDeleteIMSRecipeRecordCLI(recipeId string) bool {
	common.Infof("Permanently deleting recipe record %s in IMS via CLI", recipeId)
	return runCLICommand("deleted", "recipes", "delete", recipeId) != nil
}
