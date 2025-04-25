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

import (
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

/*
 * ims_images_test.go
 *
 * ims images test api functions
 *
 */

// Test image CRUD operations using all supported API versions
func TestImageCRUDOperationUsingAPIVersions() (passed bool) {
	passed = true

	for _, version := range common.IMSAPIVERSIONS {
		common.Infof("Testing image CRUD operations using API version: %s", version)
		common.SetIMSAPIVersion(version)
		passed = passed && TestImageCRUDOperation(version)
	}

	// default API version
	common.SetIMSAPIVersion("")
	passed = passed && TestImageCRUDOperation(common.GetIMSAPIVersion())
	return passed
}

func TestImageCRUDOperation(apiVersion string) (passed bool) {
	// Test creating an image
	imageRecord, success := TestImageCreate()
	if !success {
		return false
	}

	if apiVersion == "v2" {

		// Test updating the image
		updated := TestImageUpdate(imageRecord.Id)

		// Test deleting the image
		deleted := TestImageDeleteV2(imageRecord.Id)

		getAll := TestGetAllImages()

		return updated && deleted && getAll

	} else {
		// Test updating the image
		updated := TestImageUpdate(imageRecord.Id)

		// Test soft deleting the image
		deleted := TestImageDelete(imageRecord.Id)

		// Test undeleting the image
		undeleted := TestImageUndelete(imageRecord.Id)

		// Test hard deleting the image
		permDeleted := TestImagePermanentDelete(imageRecord.Id)

		// Test get all images
		getAll := TestGetAllImages()

		return updated && deleted && undeleted && permDeleted && getAll
	}
}

func TestImagePermanentDelete(imageId string) (passed bool) {
	// Soft delete the image
	if success := DeleteIMSImageRecordAPI(imageId); !success {
		return false
	}

	// Permanently delete the image
	if success := PermanentDeleteIMSImageRecordAPI(imageId); !success {
		return false
	}

	// Verify the image is permanently deleted
	if _, success := GetDeletedIMSImageRecordAPI(imageId, http.StatusNotFound); !success {
		common.Errorf("Image %s was not permanently deleted", imageId)
		return false
	}
	// Verify the image is not in the list of images
	if _, success := GetIMSImageRecordAPI(imageId, http.StatusNotFound); !success {
		common.Errorf("Image %s was not permanently deleted", imageId)
		return false
	}

	// Verify the image is not in the list of all images
	imageRecords, success := GetIMSImageRecordsAPI()
	if !success {
		return false
	}

	if ImageRecordExists(imageId, imageRecords) {
		common.Errorf("Image %s was not permanently deleted", imageId)
		return false
	}

	// Verify the image is not in the list of deleted images
	deletedImageRecords, success := GetDeletedIMSImageRecordsAPI()
	if !success {
		return false
	}

	if ImageRecordExists(imageId, deletedImageRecords) {
		common.Errorf("Image %s was not permanently deleted", imageId)
		return false
	}

	common.Infof("Image %s was permanently deleted", imageId)
	return true
}

func TestImageUndelete(imageId string) (passed bool) {
	if success := UndeleteIMSImageRecordAPI(imageId); !success {
		return false
	}

	// Verify the image is undeleted
	if _, success := GetIMSImageRecordAPI(imageId, http.StatusOK); !success {
		common.Errorf("Image %s was not restored", imageId)
		return false
	}
	common.Infof("Image %s was restored", imageId)
	return true
}

func TestImageDelete(imageId string) (passed bool) {
	if success := DeleteIMSImageRecordAPI(imageId); !success {
		return false
	}

	// Verify the image is deleted
	if _, success := GetDeletedIMSImageRecordAPI(imageId, http.StatusOK); !success {
		common.Errorf("Image %s was not soft deleted", imageId)
		return false
	}

	// Verify the image is not in the list of images
	if _, success := GetIMSImageRecordAPI(imageId, http.StatusNotFound); !success {
		common.Errorf("Image %s was not soft deleted", imageId)
		return false
	}

	// Verify the image is not in the list of all images
	imageRecords, success := GetIMSImageRecordsAPI()
	if !success {
		return false
	}

	if ImageRecordExists(imageId, imageRecords) {
		common.Errorf("Image %s was not soft deleted", imageId)
		return false
	}

	// Verify the image is not in the list of deleted images
	deletedImageRecords, success := GetDeletedIMSImageRecordsAPI()
	if !success {
		return false
	}

	if !ImageRecordExists(imageId, deletedImageRecords) {
		common.Errorf("Image %s was not found in the list of deleted images", imageId)
		return false
	}

	common.Infof("Image %s was deleted", imageId)
	return true
}

func TestImageUpdate(imageId string) (passed bool) {
	// getting the existing metedata info for the image and updating it as per the test
	existingImageRecord, success := GetIMSImageRecordAPI(imageId, http.StatusOK)
	if !success {
		common.Errorf("Unable to fetch Image %s ", imageId)
		return false
	}
	expectedMetadata := existingImageRecord.Metadata
	expectedMetadata["project"] = "csm"
	common.Infof("Expected metadata %v", expectedMetadata)

	arch := "aarch64"
	metadata := map[string]string{
		"operation": "set",
		"key":       "project",
		"value":     "csm",
	}
	imageRecord, success := UpdateIMSImageRecordAPI(imageId, arch, metadata)
	if !success || imageRecord.Id == "" {
		return false
	}

	if imageRecord.Arch != arch {
		common.Errorf("Expected architecture %s, got %s", arch, imageRecord.Arch)
		return false
	}

	if !common.CompareMaps(imageRecord.Metadata, expectedMetadata) {
		common.Errorf("Expected metadata %v, got %v", expectedMetadata, imageRecord.Metadata)
		return false
	}
	return true
}

func TestImageCreate() (imageRecord IMSImageRecord, passed bool) {
	// Create a new image
	imageName := "image_" + string(common.GetRandomString(10))
	metadata := map[string]string{
		"key":   "name",
		"value": imageName,
	}
	expectedMetadata := GetCreateImageExpectedMetadata(metadata)

	imageRecord, success := CreateIMSImageRecordAPI(imageName, metadata)
	if !success || imageRecord.Id == "" {
		return IMSImageRecord{}, false
	}

	// Get the created image details
	if imageRecord, success = GetIMSImageRecordAPI(imageRecord.Id, http.StatusOK); !success {
		return IMSImageRecord{}, false
	}
	if imageRecord.Name != imageName {
		common.Errorf("Expected image name %s, got %s", imageName, imageRecord.Name)
		return IMSImageRecord{}, false
	}
	if !common.CompareMaps(imageRecord.Metadata, expectedMetadata) {
		common.Errorf("Expected metadata %v, got %v", expectedMetadata, imageRecord.Metadata)
		return IMSImageRecord{}, false
	}

	// Verify the image is in the list of images
	imageRecords, success := GetIMSImageRecordsAPI()
	if !success {
		return IMSImageRecord{}, false
	}

	if !ImageRecordExists(imageRecord.Id, imageRecords) {
		common.Errorf("Image %s was not found in the list of images", imageRecord.Id)
		return IMSImageRecord{}, false
	}

	common.Infof("Image %s was created with id %s", imageName, imageRecord.Id)
	return imageRecord, true
}

func TestGetAllImages() (passed bool) {
	_, success := GetIMSImageRecordsAPI()
	if !success {
		return false
	}
	common.Infof("All images were retrieved successfully")
	return true
}

func TestImageDeleteV2(imageId string) (passed bool) {
	if success := DeleteIMSImageRecordAPI(imageId); !success {
		return false
	}

	// Verify the image is not in the list of images
	if _, success := GetIMSImageRecordAPI(imageId, http.StatusNotFound); !success {
		common.Errorf("Image %s was not deleted", imageId)
		return false
	}

	// Verify the image is not in the list of all images
	imageRecords, success := GetIMSImageRecordsAPI()
	if !success {
		return false
	}

	if ImageRecordExists(imageId, imageRecords) {
		common.Errorf("Image %s was not deleted", imageId)
		return false
	}

	common.Infof("Image %s was deleted", imageId)
	return true
}
