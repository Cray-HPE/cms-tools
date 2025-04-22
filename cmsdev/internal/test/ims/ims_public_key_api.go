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
 * ims_public_key_api.go
 *
 * ims public key api functions
 *
 */
import (
	"encoding/json"
	"net/http"
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func GetDeletedIMSPublicKeyRecordAPI(publicKeyId string) (publicKeyRecord IMSPublicKeyRecord, ok bool) {
	common.Infof("Getting public key record %s in IMS via API", publicKeyId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	uri := strings.Split(endpoints["ims"]["public_keys"].Url, "/public-keys")
	url := common.BASEURL + uri[0] + "/deleted/public-keys" + "/" + publicKeyId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return IMSPublicKeyRecord{}, false
	}

	// Extract public key record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &publicKeyRecord); err != nil {
		common.Error(err)
		return IMSPublicKeyRecord{}, false
	}
	ok = true

	return
}

func DeleteIMSPublicKeyRecordAPI(publicKeyId string) (ok bool) {
	common.Infof("Soft deleting public key record %s in IMS via API", publicKeyId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := common.BASEURL + endpoints["ims"]["public_keys"].Url + "/" + publicKeyId
	_, err := test.RestfulVerifyStatus("DELETE", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func UndeleteIMSPublicKeyRecordAPI(publicKeyId string) (ok bool) {
	common.Infof("Restoring public key record %s in IMS via API", publicKeyId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}

	// setting the payload
	payload := map[string]string{
		"operation": "undelete",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return
	}
	params.JsonStrArray = jsonPayload

	uri := strings.Split(endpoints["ims"]["public_keys"].Url, "/public-keys")
	url := common.BASEURL + uri[0] + "/deleted/public-keys" + "/" + publicKeyId
	_, err = test.RestfulVerifyStatus("PATCH", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func PermanentDeleteIMSPublicKeyRecordAPI(publicKeyId string) (ok bool) {
	common.Infof("Permanently deleting public key record %s in IMS via API", publicKeyId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	uri := strings.Split(endpoints["ims"]["public_keys"].Url, "/public-keys")
	url := common.BASEURL + uri[0] + "/deleted/public-keys" + "/" + publicKeyId
	_, err := test.RestfulVerifyStatus("DELETE", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func CreateIMSPublicKeyRecordAPI(publicKeyName string) (publicKeyRecord IMSPublicKeyRecord, ok bool) {
	common.Infof("Creating public key record %s in IMS via API", publicKeyName)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}

	sshPublicKey, err := common.GetSSHPublicKey()
	if err != nil {
		common.Error(err)
		return
	}

	// Create public key payload
	payload := map[string]string{
		"name":       publicKeyName,
		"public_key": sshPublicKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return
	}

	params.JsonStr = string(jsonPayload)
	url := common.BASEURL + endpoints["ims"]["public_keys"].Url
	resp, err := test.RestfulVerifyStatus("POST", url, *params, http.StatusCreated)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract public key record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &publicKeyRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}

// Return specific public key record in IMS via API
func GetIMSPublicKeyRecordAPI(pkeyId string) (pkeyRecord IMSPublicKeyRecord, ok bool) {
	common.Infof("Getting public key record %s in IMS via API", pkeyId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return IMSPublicKeyRecord{}, false
	}
	url := common.BASEURL + endpoints["ims"]["public_keys"].Url + "/" + pkeyId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return IMSPublicKeyRecord{}, false
	}

	// Extract public key record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &pkeyRecord); err != nil {
		common.Error(err)
		return IMSPublicKeyRecord{}, false
	}
	ok = true

	return
}

// Return a list of all public key records in IMS via API
func GetIMSPublicKeyRecordsAPI() (recordList []IMSPublicKeyRecord, ok bool) {
	common.Infof("Getting list of all public key records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := common.BASEURL + endpoints["ims"]["public_keys"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract list of public key records from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}
