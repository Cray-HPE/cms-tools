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
 * ims_recipes_api.go
 *
 * ims recipes api functions
 *
 */
import (
	"encoding/json"
	"net/http"
	"strings"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func CreateIMSRecipeRecordAPI(recipeName string, templateDict []map[string]string, requireDKMS bool) (recipeRecord IMSRecipeRecord, ok bool) {
	common.Infof("Creating recipe %s with templates %v in IMS via API", recipeName, templateDict)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	// setting the payload
	payload := map[string]interface{}{
		"name":                recipeName,
		"recipe_type":         "kiwi-ng",
		"linux_distribution":  "sles15",
		"require_dkms":        requireDKMS,
		"template_dictionary": templateDict,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return
	}
	// setting the payload
	params.JsonStr = string(jsonPayload)
	url := common.BASEURL + endpoints["ims"]["recipes"].Url
	resp, err := test.RestfulVerifyStatus("POST", url, *params, http.StatusCreated)
	if err != nil {
		common.Error(err)
		return
	}
	// Extract recipe record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recipeRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

func UpdateIMSRecipeRecordAPI(recipeId string, arch string, templateDict []map[string]string) (recipeRecord IMSRecipeRecord, ok bool) {
	common.Infof("Updating recipe %s with arch %s and templates %v", recipeId, arch, templateDict)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	// setting the payload
	payload := map[string]interface{}{
		"arch":                arch,
		"template_dictionary": templateDict,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		common.Error(err)
		return
	}

	params.JsonStrArray = jsonPayload

	url := common.BASEURL + endpoints["ims"]["recipes"].Url + "/" + recipeId
	resp, err := test.RestfulVerifyStatus("PATCH", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}
	// Extract recipe record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recipeRecord); err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

func DeleteIMSRecipeRecordAPI(recipeId string) (ok bool) {
	common.Infof("Soft deleting recipe %s", recipeId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := common.BASEURL + endpoints["ims"]["recipes"].Url + "/" + recipeId
	_, err := test.RestfulVerifyStatus("DELETE", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

func GetDeletedIMSRecipeRecordAPI(recipeId string, httpStatus int) (recipeRecord IMSRecipeRecord, ok bool) {
	common.Infof("Getting deleted recipe record %s in IMS via API", recipeId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return IMSRecipeRecord{}, false
	}
	uri := strings.Split(endpoints["ims"]["recipes"].Url, "/recipes")
	url := common.BASEURL + uri[0] + "/deleted/recipes" + "/" + recipeId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, httpStatus)
	if err != nil {
		common.Error(err)
		return IMSRecipeRecord{}, false
	}
	// Extract recipe record from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recipeRecord); err != nil {
		common.Error(err)
		return IMSRecipeRecord{}, false
	}
	ok = true
	return
}

func UndeleteIMSRecipeRecordAPI(recipeId string) (ok bool) {
	common.Infof("Restoring recipe %s", recipeId)
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
	uri := strings.Split(endpoints["ims"]["recipes"].Url, "/recipes")
	url := common.BASEURL + uri[0] + "/deleted/recipes" + "/" + recipeId
	_, err = test.RestfulVerifyStatus("PATCH", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

func PermanentDeleteIMSRecipeRecordAPI(recipeId string) (ok bool) {
	common.Infof("Permanently deleting recipe %s", recipeId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	// getting the base uri needed for hard delete
	uri := strings.Split(endpoints["ims"]["recipes"].Url, "/recipes")
	url := common.BASEURL + uri[0] + "/deleted/recipes" + "/" + recipeId
	_, err := test.RestfulVerifyStatus("DELETE", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return
	}
	ok = true
	return
}

// Return a specific recipe record in IMS via API
func GetIMSRecipeRecordAPI(recipeId string, httpStatus int) (recipeRecord IMSRecipeRecord, ok bool) {
	common.Infof("Getting recipe record %s in IMS via API", recipeId)
	params := test.GetAccessTokenParams()
	if params == nil {
		return IMSRecipeRecord{}, false
	}

	url := common.BASEURL + endpoints["ims"]["recipes"].Url + "/" + recipeId
	resp, err := test.RestfulVerifyStatus("GET", url, *params, httpStatus)
	if err != nil {
		return IMSRecipeRecord{}, false
	}

	// Extract recipe record from command output
	common.Infof("Decoding JSON in command output")
	err = json.Unmarshal(resp.Body(), &recipeRecord)
	if err != nil {
		return IMSRecipeRecord{}, false
	}
	ok = true
	return
}

// Return a list of all recipe records in IMS via API
func GetIMSRecipeRecordsAPI() (recordList []IMSRecipeRecord, ok bool) {
	common.Infof("Getting list of all recipe records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}
	url := common.BASEURL + endpoints["ims"]["recipes"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return
	}

	// Extract list of recipe records from response
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return
	}
	ok = true

	return
}
