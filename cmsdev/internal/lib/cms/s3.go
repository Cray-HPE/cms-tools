package cms

/*
 * s3.go
 *
 * s3 function library
 *
 * © Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"encoding/json"
	"fmt"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

type ListArtifacts struct {
	Artifacts []ArtifactRecord
}

type ArtifactRecord struct {
	LastModified, ETag, StorageClass, Key string
	Owner                                 map[string]string
	Size                                  int
}

// Get a list of S3 buckets via CLI.
// If error, logs it and returns nil.
func GetBuckets() (bucketList []string) {
	common.Infof("Getting list of all S3 buckets via CLI")
	cmdOut := test.RunCLICommand("artifacts buckets list --format json")
	if cmdOut == nil {
		return nil
	}

	// Extract list of buckets from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &bucketList); err != nil {
		common.Error(err)
		return nil
	}
	return
}

// Get a list of artifacts in the specified S3 bucket (via CLI).
// If error, logs it and returns nil.
func GetArtifactsInBucket(bucket string) []ArtifactRecord {
	common.Infof("Getting list of all S3 artifacts in %s bucket via CLI", bucket)
	cmdOut := test.RunCLICommand(fmt.Sprintf("artifacts list %s --format json", bucket))
	if cmdOut == nil {
		return nil
	}

	var listArtifactsObject ListArtifacts

	// Extract object from command output
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &listArtifactsObject); err != nil {
		common.Error(err)
		return nil
	}

	return listArtifactsObject.Artifacts
}
