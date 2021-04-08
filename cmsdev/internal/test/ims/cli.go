package ims

/*
 * ims.go
 *
 * ims cli functions
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
 */

import (
	"encoding/json"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

func runCLICommand(cmdArgs ...string) []byte {
	return test.RunCLICommandJSON("ims", cmdArgs...)
}

// Return specific image record in IMS via CLI
func getIMSImageRecordCLI(imageId string) (imageRecord IMSImageRecord, ok bool) {
	ok = false

	common.Infof("Getting image record %s in IMS via CLI", imageId)
	cmdOut := runCLICommand("images", "describe", imageId)
	if cmdOut == nil {
		return
	}

	// Extract image record from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &imageRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

// Return a list of all image records in IMS via CLI
func getIMSImageRecordsCLI() (recordList []IMSImageRecord, ok bool) {
	ok = false

	common.Infof("Getting list of all image records in IMS via CLI")
	cmdOut := runCLICommand("images", "list")
	if cmdOut == nil {
		return
	}

	// Extract list of image records from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

// Return specific job record in IMS via CLI
func getIMSJobRecordCLI(jobId string) (jobRecord IMSJobRecord, ok bool) {
	ok = false

	common.Infof("Getting job record %s in IMS via CLI", jobId)
	cmdOut := runCLICommand("jobs", "describe", jobId)
	if cmdOut == nil {
		return
	}

	// Extract job records from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &jobRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return a list of all job records in IMS via CLI
func getIMSJobRecordsCLI() (recordList []IMSJobRecord, ok bool) {
	ok = false

	common.Infof("Getting list of all job records in IMS via CLI")
	cmdOut := runCLICommand("jobs", "list")
	if cmdOut == nil {
		return
	}

	// Extract list of job records from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return specific public key record in IMS via CLI
func getIMSPublicKeyRecordCLI(pkeyId string) (pkeyRecord IMSPublicKeyRecord, ok bool) {
	ok = false

	common.Infof("Getting public key record %s in IMS via CLI", pkeyId)
	cmdOut := runCLICommand("public-keys", "describe", pkeyId)
	if cmdOut == nil {
		return
	}

	// Extract public key record from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &pkeyRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return a list of all public key records in IMS via CLI
func getIMSPublicKeyRecordsCLI() (recordList []IMSPublicKeyRecord, ok bool) {
	ok = false

	common.Infof("Getting list of all public key records in IMS via CLI")
	cmdOut := runCLICommand("public-keys", "list")
	if cmdOut == nil {
		return
	}

	// Extract list of public key records from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return a specific recipe record in IMS via CLI
func getIMSRecipeRecordCLI(recipeId string) (recipeRecord IMSRecipeRecord, ok bool) {
	ok = false

	common.Infof("Describing recipe record %s in IMS via CLI", recipeId)
	cmdOut := runCLICommand("recipes", "describe", recipeId)
	if cmdOut == nil {
		return
	}

	// Extract recipe record from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recipeRecord); err != nil {
		return
	}
	ok = true
	return
}

// Return a list of all recipe records in IMS via CLI
func getIMSRecipeRecordsCLI() (recordList []IMSRecipeRecord, ok bool) {
	ok = false

	common.Infof("Getting list of all recipe records in IMS via CLI")
	cmdOut := runCLICommand("recipes", "list")
	if cmdOut == nil {
		return
	}

	// Extract list of recipe records from command output
	common.Infof("Decoding JSON in command output")
	err := json.Unmarshal(cmdOut, &recordList)
	if err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}