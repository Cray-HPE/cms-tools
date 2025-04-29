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

	apiVersion := common.GetIMSAPIVersion()
	url := constructIMSURL("recipes", apiVersion)
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

	apiVersion := common.GetIMSAPIVersion()
	url := constructIMSURL("recipes", apiVersion) + "/" + recipeId

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

	apiVersion := common.GetIMSAPIVersion()
	url := constructIMSURL("recipes", apiVersion) + "/" + recipeId

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

	apiVersion := common.GetIMSAPIVersion()
	baseURL := constructIMSURL("recipes", apiVersion)
	// getting the base uri needed for getting deleted recipe record
	uri := strings.Split(baseURL, "/recipes")
	url := uri[0] + "/deleted/recipes" + "/" + recipeId

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

func GetDeletedIMSRecipeRecordsAPI() (recordList []IMSRecipeRecord, ok bool) {
	common.Infof("Getting list of all deleted recipe records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil {
		return
	}

	apiVersion := common.GetIMSAPIVersion()
	baseURL := constructIMSURL("recipes", apiVersion)
	// getting the base uri needed for getting all deleted recipe record
	uri := strings.Split(baseURL, "/recipes")
	url := uri[0] + "/deleted/recipes"

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

	apiVersion := common.GetIMSAPIVersion()
	baseURL := constructIMSURL("recipes", apiVersion)
	// getting the base uri needed for restoring deleted recipe record
	uri := strings.Split(baseURL, "/recipes")
	url := uri[0] + "/deleted/recipes" + "/" + recipeId

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

	apiVersion := common.GetIMSAPIVersion()
	baseURL := constructIMSURL("recipes", apiVersion)
	// getting the base uri needed for getting deleted recipe record
	uri := strings.Split(baseURL, "/recipes")
	url := uri[0] + "/deleted/recipes" + "/" + recipeId

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

	apiVersion := common.GetIMSAPIVersion()
	url := constructIMSURL("recipes", apiVersion) + "/" + recipeId

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

	apiVersion := common.GetIMSAPIVersion()
	url := constructIMSURL("recipes", apiVersion)

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

func RecipeRecordExists(recipeId string, recipeRecords []IMSRecipeRecord) (exists bool) {
	for _, recipeRecord := range recipeRecords {
		if recipeRecord.Id == recipeId {
			return true
		}
	}
	common.Infof("Recipe %s was not found in the list of recipes", recipeId)
	return false
}

func VerifyIMSRecipeRecord(recipeRecord IMSRecipeRecord, existingRecipeRecord IMSRecipeRecord) (ok bool) {
	ok = true
	if recipeRecord.Name != existingRecipeRecord.Name {
		common.Errorf("Recipe name %s does not match expected name %s", recipeRecord.Name, existingRecipeRecord.Name)
		ok = false
	}

	if recipeRecord.Linux_distribution != existingRecipeRecord.Linux_distribution {
		common.Errorf("Recipe linux distribution %s does not match expected linux distribution %s", recipeRecord.Linux_distribution, existingRecipeRecord.Linux_distribution)
		ok = false
	}

	if recipeRecord.Arch != existingRecipeRecord.Arch {
		common.Errorf("Recipe arch %s does not match expected arch %s", recipeRecord.Arch, existingRecipeRecord.Arch)
		ok = false
	}

	if recipeRecord.Recipe_type != existingRecipeRecord.Recipe_type {
		common.Errorf("Recipe type %s does not match expected recipe type %s", recipeRecord.Recipe_type, existingRecipeRecord.Recipe_type)
		ok = false
	}

	if recipeRecord.Require_dkms != existingRecipeRecord.Require_dkms {
		common.Errorf("Recipe require_dkms %t does not match expected require_dkms %t", recipeRecord.Require_dkms, existingRecipeRecord.Require_dkms)
		ok = false
	}

	if !common.CompareSlicesOfMaps(recipeRecord.Template_dictionary, existingRecipeRecord.Template_dictionary) {
		common.Errorf("Template dictionary %v does not match expected template dictionary %v", recipeRecord.Template_dictionary, existingRecipeRecord.Template_dictionary)
		ok = false
	}

	ok = true
	return
}
