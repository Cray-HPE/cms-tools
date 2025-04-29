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

import "stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"

/*
 * ims_public_key_cli_test.go
 *
 * ims public key cli test functions
 *
 */

func TestPublicKeyCRUDOperationUsingCLI() (passed bool) {
	common.PrintLog("Testing Public Key CRUD operations using CLI")
	// Test creating a public key
	publicKeyRecord, success := TestCLIPublicKeyCreate()
	if !success {
		return false
	}

	// Test soft deleting the public key
	deleted := TestCLIPublicKeyDelete(publicKeyRecord.Id)

	// Test undeleting the public key
	undeleted := TestCLIPublicKeyUndelete(publicKeyRecord.Id)

	// Test hard deleting the public key
	permDeleted := TestCLIPublicKeyPermanentDelete(publicKeyRecord.Id)

	// Test get all public keys
	getAll := TestCLIGetAllPublicKeys()

	return deleted && undeleted && permDeleted && getAll
}

func TestCLIPublicKeyCreate() (publicKeyRecord IMSPublicKeyRecord, passed bool) {
	publicKeyName := "public_key_" + string(common.GetRandomString(10))
	publicKeyRecord, success := CreateIMSPublicKeyRecordCLI(publicKeyName)
	if !success {
		return IMSPublicKeyRecord{}, false
	}

	// Get the public key record
	publicKeyRecord, success = getIMSPublicKeyRecordCLI(publicKeyRecord.Id)
	if !success {
		return IMSPublicKeyRecord{}, false
	}

	// Verify the public key exists in the list of public keys
	publickeyRecords, success := getIMSPublicKeyRecordsCLI()
	if !success {
		return IMSPublicKeyRecord{}, false
	}

	if !PublicKeyRecordExists(publicKeyRecord.Id, publickeyRecords) {
		common.Errorf("Public key %s was not found in the list of public keys", publicKeyRecord.Id)
		return IMSPublicKeyRecord{}, false
	}

	return publicKeyRecord, true
}

func TestCLIPublicKeyDelete(publicKeyId string) (passed bool) {
	if success := DeleteIMSPublicKeyRecordCLI(publicKeyId); !success {
		return false
	}
	// Verify the public key is soft deleted
	if _, success := GetDeletedIMSPublicKeyRecordCLI(publicKeyId); !success {
		return false
	}

	// verify the public key is not in the list of public keys
	if _, success := getIMSPublicKeyRecordCLI(publicKeyId); success {
		return false
	}

	// verify the public key is not in the list of all public keys
	publickeyRecords, success := getIMSPublicKeyRecordsCLI()
	if !success {
		return false
	}

	if PublicKeyRecordExists(publicKeyId, publickeyRecords) {
		common.Errorf("Public key %s was not soft deleted", publicKeyId)
		return false
	}

	// verify the public key is in the list of all deleted public keys
	deletedpublicKeyRecords, success := GetDeletedIMSPublicKeyRecordCLIs()
	if !success {
		return false
	}

	if !PublicKeyRecordExists(publicKeyId, deletedpublicKeyRecords) {
		common.Errorf("Public key %s was not found in the list of deleted public keys", publicKeyId)
		return false
	}

	common.Infof("Public key %s successfully soft deleted", publicKeyId)
	return true
}

func TestCLIPublicKeyUndelete(publicKeyId string) (passed bool) {
	if success := UndeleteIMSPublicKeyRecordCLI(publicKeyId); !success {
		return false
	}
	// Verify the public key is restored
	if _, success := getIMSPublicKeyRecordCLI(publicKeyId); !success {
		return false
	}
	common.Infof("Public key %s successfully restored", publicKeyId)
	return true
}

func TestCLIPublicKeyPermanentDelete(publicKeyId string) (passed bool) {
	// soft delete the public key
	if success := DeleteIMSPublicKeyRecordCLI(publicKeyId); !success {
		return false
	}

	// hard delete the public key
	if success := PermanentDeleteIMSPublicKeyRecordCLI(publicKeyId); !success {
		return false
	}
	// Verify the public key is hard deleted
	if _, success := GetDeletedIMSPublicKeyRecordCLI(publicKeyId); success {
		return false
	}
	// Verify the public key is not in the list of public keys
	if _, success := getIMSPublicKeyRecordCLI(publicKeyId); success {
		common.Errorf("Public key %s was not permanently deleted", publicKeyId)
		return false
	}

	// verify the public key is not in the list of all public keys
	publickeyRecords, success := getIMSPublicKeyRecordsCLI()
	if !success {
		common.Errorf("Public key %s was not permanently deleted", publicKeyId)
		return false
	}

	if PublicKeyRecordExists(publicKeyId, publickeyRecords) {
		common.Errorf("Public key %s was not permanently deleted", publicKeyId)
		return false
	}

	// verify the public key is not in the list of all deleted public keys
	deletedpublicKeyRecords, success := GetDeletedIMSPublicKeyRecordCLIs()
	if !success {
		return false
	}

	if PublicKeyRecordExists(publicKeyId, deletedpublicKeyRecords) {
		common.Errorf("Public key %s was not permanently deleted", publicKeyId)
		return false
	}

	common.Infof("Public key %s successfully hard deleted", publicKeyId)
	return true
}

func TestCLIGetAllPublicKeys() (passed bool) {
	// Get all public keys
	_, success := getIMSPublicKeyRecordsCLI()
	if !success {
		return false
	}
	common.Infof("Public key records successfully retrieved")
	return true
}
