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
 * ims_images_cli_test.go
 *
 * ims images cli test functions
 *
 */
import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func TestImageCRUDOperationUsingCLI() (passed bool) {
	// Test creating an image
	imageRecord, success := TestCLIImageCreate()
	if !success {
		return false
	}
	// Test updating the image
	updated := TestCLIImageUpdate(imageRecord.Id)

	// Test soft deleting the image
	deleted := TestCLIImageDelete(imageRecord.Id)

	// Test undeleting the image
	undeleted := TestCLIImageUndelete(imageRecord.Id)

	// Test hard deleting the image
	permDeleted := TestCLIImagePermanentDelete(imageRecord.Id)

	// Test get all images
	getAll := TestCLIGetAllImages()

	return updated && deleted && undeleted && permDeleted && getAll
}

func TestCLIImageCreate() (imageRecord IMSImageRecord, passed bool) {

	// Create the image
	imageName := "image_" + string(common.GetRandomString(10))

	expectedMetadata := map[string]string{
		"name": imageName,
	}

	imageRecord, success := CreateIMSImageRecordCLI(imageName)
	if !success {
		return IMSImageRecord{}, false
	}
	// Get the image record
	imageRecord, success = getIMSImageRecordCLI(imageRecord.Id)
	if !success {
		return IMSImageRecord{}, false
	}
	if imageRecord.Name != imageName {
		common.Errorf("Image %s was not created with name %s", imageRecord.Id, imageName)
		return IMSImageRecord{}, false
	}

	if !common.CompareMaps(imageRecord.Metadata, expectedMetadata) {
		common.Errorf("Expected metadata %v, got %v", expectedMetadata, imageRecord.Metadata)
		return IMSImageRecord{}, false
	}

	// Verify the image is in the list of images
	imageRecords, success := getIMSImageRecordsCLI()
	if !success {
		return IMSImageRecord{}, false
	}

	if !ImageRecordExists(imageRecord.Id, imageRecords) {
		common.Errorf("Image %s was not found in the list of images", imageRecord.Id)
		return IMSImageRecord{}, false
	}

	common.Infof("Created image %s with ID %s", imageName, imageRecord.Id)
	return imageRecord, true
}

func TestCLIImageUpdate(imageId string) (passed bool) {
	// build the expected metadata
	existingImageRecord, success := getIMSImageRecordCLI(imageId)
	expectedMetadata := existingImageRecord.Metadata
	expectedMetadata["project"] = "csm"
	common.Infof(("Expected metadata: %v"), expectedMetadata)
	// Update the image
	arch := "aarch64"
	if _, success := UpdateIMSImageRecordCLI(imageId, arch); !success {
		return false
	}
	// Verify the image is updated
	imageRecord, success := getIMSImageRecordCLI(imageId)

	if !success || imageRecord.Arch != arch {
		common.Errorf("Image %s was not updated with arch %s", imageId, arch)
		return false
	}

	if !common.CompareMaps(imageRecord.Metadata, expectedMetadata) {
		common.Errorf("Expected metadata %v, got %v", expectedMetadata, imageRecord.Metadata)
		return false
	}
	common.Infof("Updated image %s with arch %s", imageId, arch)
	return true
}

func TestCLIImageDelete(imageId string) (passed bool) {
	// Soft delete the image
	if success := DeleteIMSImageRecordCLI(imageId); !success {
		return false
	}
	// Verify the image is soft deleted
	if _, success := GetDeletedIMSImageRecordCLI(imageId); !success {
		common.Errorf("Image %s was not soft deleted", imageId)
		return false
	}

	// Verify the image is not in the list of images
	if _, success := getIMSImageRecordCLI(imageId); success {
		common.Errorf("Image %s was not soft deleted", imageId)
		return false
	}

	// Verify the image is not in the list of images
	imageRecords, success := getIMSImageRecordsCLI()
	if !success {
		return false
	}

	if ImageRecordExists(imageId, imageRecords) {
		common.Errorf("Image %s was not soft deleted", imageId)
		return false
	}

	// Verify the image is not in the list of deleted images
	deletedImageRecords, success := GetDeletedIMSImageRecordsCLI()
	if !success {
		return false
	}

	if !ImageRecordExists(imageId, deletedImageRecords) {
		common.Errorf("Image %s was not found in the list of deleted images", imageId)
		return false
	}

	common.Infof("Image %s was soft deleted", imageId)
	return true
}

func TestCLIImageUndelete(imageId string) (passed bool) {
	// Undelete the image
	if success := UndeleteIMSImageRecordCLI(imageId); !success {
		return false
	}
	// Verify the image is undeleted
	if _, success := getIMSImageRecordCLI(imageId); !success {
		common.Errorf("Image %s was not restored", imageId)
		return false
	}
	common.Infof("Image %s was restored", imageId)
	return true
}

func TestCLIImagePermanentDelete(imageId string) (passed bool) {
	// Soft delete the image
	if success := DeleteIMSImageRecordCLI(imageId); !success {
		return false
	}

	// Permanently delete the image
	if success := PermanentDeleteIMSImageRecordCLI(imageId); !success {
		return false
	}
	// Verify the image is hard deleted
	if _, success := GetDeletedIMSImageRecordCLI(imageId); success {
		common.Errorf("Image %s was not  permanently deleted", imageId)
		return false
	}
	// Verify the image is not in the list of images
	if _, success := getIMSImageRecordCLI(imageId); success {
		common.Errorf("Image %s was not permanently deleted", imageId)
		return false
	}

	// Verify the image is not in the list of images
	imageRecords, success := getIMSImageRecordsCLI()
	if !success {
		return false
	}

	if ImageRecordExists(imageId, imageRecords) {
		common.Errorf("Image %s was not permanently deleted", imageId)
		return false
	}

	// Verify the image is not in the list of deleted images
	deletedImageRecords, success := GetDeletedIMSImageRecordsCLI()
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

func TestCLIGetAllImages() (passed bool) {
	// Get all images
	if _, success := getIMSImageRecordsCLI(); !success {
		return false
	}
	common.Infof("Got all images")
	return true
}
