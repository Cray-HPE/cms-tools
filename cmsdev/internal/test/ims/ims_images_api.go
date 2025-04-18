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
 * ims_images_api.go
 *
 * ims images api functions
 *
 */
import (
	"encoding/json"
	"net/http"
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func CreateIMSImageRecordAPI(imageName string) (imageRecord IMSImageRecord, ok bool) {
	common.Infof("Creating image %s in IMS via API", imageName)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}

	// setting the payload
	payload := map[string]string{
		"name": imageName,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return
	}

	params.JsonStr = string(jsonPayload)
	url := common.BASEURL + endpoints["ims"]["images"].Url
	resp, err := test.RestfulVerifyStatus("POST", url, *params, http.StatusCreated)
	if err != nil {
		common.Error(err)
		return
	}

	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &imageRecord); err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func UpdateIMSImageRecordAPI(imageId string, arch string) (imageRecord IMSImageRecord, ok bool) {
	common.Infof("Updating image %s with arch %s", imageId, arch)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}

	// setting the payload
	payload := map[string]string{
		"arch": arch,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return
	}

	params.JsonStrArray = jsonPayload
	url := common.BASEURL + endpoints["ims"]["images"].Url + "/" + imageId
	resp, err := test.RestfulVerifyStatus("PATCH", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &imageRecord); err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func DeleteIMSImageRecordAPI(imageId string) (ok bool) {
	common.Infof("Soft deleting image %s", imageId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := common.BASEURL + endpoints["ims"]["images"].Url + "/" + imageId
	_, err := test.RestfulVerifyStatus("DELETE", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func UndeleteIMSImageRecordAPI(imageId string) (ok bool) {
	common.Infof("Restoring image %s", imageId)
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

	// getting the base uri needed for undelete
	uri := strings.Split(endpoints["ims"]["images"].Url, "/images")
	url := common.BASEURL + uri[0] + "/deleted/images" + "/" + imageId
	_, err = test.RestfulVerifyStatus("PATCH", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func HardDeleteIMSImageRecordAPI(imageId string) (ok bool) {
	common.Infof("Permanently deleting image %s", imageId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	// getting the base uri needed for hard delete
	uri := strings.Split(endpoints["ims"]["images"].Url, "/images")
	url := common.BASEURL + uri[0] + "/deleted/images" + "/" + imageId
	_, err := test.RestfulVerifyStatus("DELETE", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}

	ok = true
	return
}

func GetDeletedIMSImageRecordAPI(imageId string) (imageRecord IMSImageRecord, ok bool) {
	common.Infof("Getting deleted image record %s in IMS via API", imageId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return IMSImageRecord{}, false
	}
	// getting the base uri needed for hard delete
	uri := strings.Split(endpoints["ims"]["images"].Url, "/images")
	url := common.BASEURL + uri[0] + "/deleted/images" + "/" + imageId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return IMSImageRecord{}, false
	}
	// Extract image record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &imageRecord); err != nil {
		common.Error(err)
		return IMSImageRecord{}, false
	}

	ok = true
	return
}

// Return specific image record in IMS via API
func GetIMSImageRecordAPI(imageId string) (imageRecord IMSImageRecord, ok bool) {
	common.Infof("Getting image record %s in IMS via API", imageId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return IMSImageRecord{}, false
	}
	url := common.BASEURL + endpoints["ims"]["images"].Url + "/" + imageId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return IMSImageRecord{}, false
	}

	// Extract image record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &imageRecord); err != nil {
		common.Error(err)
		return IMSImageRecord{}, false
	}
	ok = true

	return
}

// Return a list of all image records in IMS via API
func GetIMSImageRecordsAPI() (recordList []IMSImageRecord, ok bool) {
	common.Infof("Getting list of all image records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return []IMSImageRecord{}, false
	}
	url := common.BASEURL + endpoints["ims"]["images"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return []IMSImageRecord{}, false
	}

	// Extract list of image records from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return []IMSImageRecord{}, false
	}
	ok = true

	return
}
