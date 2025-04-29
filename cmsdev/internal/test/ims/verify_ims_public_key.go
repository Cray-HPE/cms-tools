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
	"fmt"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

/*
 * ims_public_key_test.go
 *
 * ims public key test api functions
 *
 */

func TestPublicKeyCRUDOperationUsingAPIVersions() (passed bool) {
	passed = true

	for _, apiVersion := range common.IMSAPIVERSIONS {
		common.PrintLog(fmt.Sprintf("Testing IMS Public Key CRUD operations using API version: %s", apiVersion))
		common.SetIMSAPIVersion(apiVersion)
		passed = passed && TestPublicKeyCRUDOperation(apiVersion)
	}

	// Test the default API version
	common.PrintLog("Testing IMS Public Key CRUD operations using default API version.")
	common.SetIMSAPIVersion("")
	passed = passed && TestPublicKeyCRUDOperation(common.GetIMSAPIVersion())

	return passed
}

func TestPublicKeyCRUDOperation(apiVersion string) (passed bool) {
	// Test creating a public key
	publicKeyRecord, success := TestPublicKeyCreate()
	if !success {
		return false
	}

	// Test get all public keys
	getAll := TestGetAllPublicKeys()

	if apiVersion == "v3" {
		// Test soft deleting the public key
		deleted := TestPublicKeyDelete(publicKeyRecord.Id)

		// Test undeleting the public key
		undeleted := TestPublicKeyUndelete(publicKeyRecord.Id)

		// Test permanent deleting the public key
		permDeleted := TestPublicKeyPermanentDelete(publicKeyRecord.Id)

		return deleted && undeleted && permDeleted && getAll
	}
	// Test deleting the public key
	deleted := TestPublicKeyDeleteV2(publicKeyRecord.Id)

	return deleted && getAll
}

func TestPublicKeyDelete(publicKeyId string) (passed bool) {
	if success := DeleteIMSPublicKeyRecordAPI(publicKeyId); !success {
		return false
	}

	// Verify the public key is soft deleted
	if _, success := GetDeletedIMSPublicKeyRecordAPI(publicKeyId, http.StatusOK); !success {
		return false
	}

	// Verify the public key is not in the list of public keys
	if _, success := GetIMSPublicKeyRecordAPI(publicKeyId, http.StatusNotFound); !success {
		common.Errorf("Public key %s was not soft deleted", publicKeyId)
		return false
	}

	// Verify the public key is in the list of all deleted public keys
	deletedPublicKeyRecords, success := GetDeletedIMSPublicKeyRecordsAPI()
	if !success {
		return false
	}

	if !PublicKeyRecordExists(publicKeyId, deletedPublicKeyRecords) {
		common.Errorf("Public key %s was not deleted", publicKeyId)
		return false
	}

	// Verify the public key is not in the list of all public keys
	publicKeyRecords, success := GetIMSPublicKeyRecordsAPI()
	if !success {
		return false
	}

	if PublicKeyRecordExists(publicKeyId, publicKeyRecords) {
		common.Errorf("Public key %s was not deleted", publicKeyId)
		return false
	}

	common.Infof("Public key %s successfully soft deleted", publicKeyId)
	return true
}

func TestPublicKeyUndelete(publicKeyId string) (passed bool) {
	if success := UndeleteIMSPublicKeyRecordAPI(publicKeyId); !success {
		return false
	}

	// Verify the public key is not soft deleted
	if _, success := GetIMSPublicKeyRecordAPI(publicKeyId, http.StatusOK); !success {
		return false
	}
	common.Infof("Public key %s successfully restored", publicKeyId)
	return true
}

func TestPublicKeyPermanentDelete(publicKeyId string) (passed bool) {
	// Soft delete the public key
	if success := DeleteIMSPublicKeyRecordAPI(publicKeyId); !success {
		return false
	}

	// Permanently delete the public key
	if success := PermanentDeleteIMSPublicKeyRecordAPI(publicKeyId); !success {
		return false
	}

	// Verify the public key is hard deleted
	if _, success := GetDeletedIMSPublicKeyRecordAPI(publicKeyId, http.StatusNotFound); !success {
		common.Errorf("Public key %s was not permanently deleted", publicKeyId)
		return false
	}
	// Verify the public key is not in the list of public keys
	if _, success := GetIMSPublicKeyRecordAPI(publicKeyId, http.StatusNotFound); !success {
		common.Errorf("Public key %s was not permanently deleted", publicKeyId)
		return false
	}

	// Verify the public key is not in the list of all deleted public keys
	deletedPublicKeyRecords, success := GetDeletedIMSPublicKeyRecordsAPI()
	if !success {
		return false
	}

	if PublicKeyRecordExists(publicKeyId, deletedPublicKeyRecords) {
		common.Errorf("Public key %s was not deleted", publicKeyId)
		return false
	}

	// Verify the public key is not in the list of all public keys
	publicKeyRecords, success := GetIMSPublicKeyRecordsAPI()
	if !success {
		return false
	}

	if PublicKeyRecordExists(publicKeyId, publicKeyRecords) {
		common.Errorf("Public key %s was not deleted", publicKeyId)
		return false
	}

	common.Infof("Public key %s was permanently deleted", publicKeyId)
	return true
}

func TestGetAllPublicKeys() (passed bool) {
	// Get all public keys
	_, success := GetIMSPublicKeyRecordsAPI()
	if !success {
		return false
	}
	common.Infof("Public key records were successfully retrieved")
	return true
}

func TestPublicKeyCreate() (publicKeyRecord IMSPublicKeyRecord, passed bool) {
	publicKeyName := "public_key_" + string(common.GetRandomString(10))
	publicKeyRecord, success := CreateIMSPublicKeyRecordAPI(publicKeyName)
	if !success {
		return IMSPublicKeyRecord{}, false
	}

	// Verify the public key is created
	publicKeyRecord, success = GetIMSPublicKeyRecordAPI(publicKeyRecord.Id, http.StatusOK)
	if !success || publicKeyRecord.Name != publicKeyName {
		common.Errorf("Public key %s was not created", publicKeyName)
		return IMSPublicKeyRecord{}, false
	}

	// Verify the public key is in the list of public keys
	publicKeyRecords, success := GetIMSPublicKeyRecordsAPI()
	if !success {
		return IMSPublicKeyRecord{}, false
	}

	if !PublicKeyRecordExists(publicKeyRecord.Id, publicKeyRecords) {
		common.Errorf("Public key %s was not found in the list of public keys", publicKeyRecord.Id)
		return IMSPublicKeyRecord{}, false
	}

	common.Infof("Public key %s was created with id %s", publicKeyName, publicKeyRecord.Id)
	return publicKeyRecord, true
}

func TestPublicKeyDeleteV2(publicKeyId string) (passed bool) {
	if success := DeleteIMSPublicKeyRecordAPI(publicKeyId); !success {
		return false
	}

	// Verify the public key is not in the list of public keys
	if _, success := GetIMSPublicKeyRecordAPI(publicKeyId, http.StatusNotFound); !success {
		common.Errorf("Public key %s was not deleted", publicKeyId)
		return false
	}

	// Verify the public key is not in the list of all public keys
	publicKeyRecords, success := GetIMSPublicKeyRecordsAPI()
	if !success {
		return false
	}

	if PublicKeyRecordExists(publicKeyId, publicKeyRecords) {
		common.Errorf("Public key %s was not deleted", publicKeyId)
		return false
	}

	common.Infof("Public key %s successfully deleted", publicKeyId)
	return true
}
