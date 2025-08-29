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

/*
* cfs_configurations_api.go
*
* cfs configurations api functions
*
 */
package cfs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	resty "gopkg.in/resty.v1"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

func GetCsmProductCatalogData() (data map[string]interface{}, err error) {
	resp, err := k8s.GetConfigMapDataField(common.NAMESPACE, common.CSMPRODCATALOGCMNAME, "csm")
	if err != nil {
		return nil, err
	}

	//convert YAML
	respMap, err := common.DecodeYAMLIntoStringMap(resp)
	if err != nil {
		return nil, err
	}

	return respMap, nil
}

func GetCSMLatestVersion(prodCatlogData map[string]interface{}) (version string, err error) {
	// Sort the keys in the JSON object to ensure consistent ordering
	keys := make([]string, 0, len(prodCatlogData))
	for key := range prodCatlogData {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return "", fmt.Errorf("no keys found in the configuration map")
	}
	//Get the last key from the sorted keys where version has both configuration and images as keys
	latestCSMVersion := ""
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		if _, ok := prodCatlogData[key].(map[interface{}]interface{})["configuration"]; ok {
			if _, ok := prodCatlogData[key].(map[interface{}]interface{})["images"]; ok {
				latestCSMVersion = key
				break
			}
		}
	}

	common.Infof("Latest CSM version: %s", latestCSMVersion)
	return latestCSMVersion, nil
}

func GetProdCatalogConfigData() (cfsConfigLayerData CsmProductCatalogConfiguration, err error) {
	pordCatalogData, err := GetCsmProductCatalogData()
	if err != nil {
		return CsmProductCatalogConfiguration{}, err
	}

	// Get the latest CSM version data
	csmVersion, err := GetCSMLatestVersion(pordCatalogData)
	if err != nil {
		return CsmProductCatalogConfiguration{}, err
	}
	latestCSMDataRaw := pordCatalogData[csmVersion]
	common.Infof("Latest CSM data: %v", latestCSMDataRaw)
	latestCSMData := latestCSMDataRaw.(map[interface{}]interface{})
	return CsmProductCatalogConfiguration{
		Clone_url: latestCSMData["configuration"].(map[interface{}]interface{})["clone_url"].(string),
		Commit:    latestCSMData["configuration"].(map[interface{}]interface{})["commit"].(string),
	}, nil

}

func GetCreateCFGConfigurationPayload(apiVersion string, addTenant bool) (payload string, ok bool) {
	common.Infof("Getting product catalog configuration layer data")
	configData, err := GetProdCatalogConfigData()
	if err != nil {
		return "", false
	}

	cfgLayerName := "Configuration_Layer_" + string(common.GetRandomString(10))

	// Create the CFS configuration payload
	cfsPayload := map[string]interface{}{
		"layers": []map[string]string{
			{
				"commit":   configData.Commit,
				"playbook": DEFAULT_PLAYBOOK,
				"name":     cfgLayerName,
			},
		},
	}

	// add the tenant in the payload body if the configuration is created by admin
	if addTenant {
		cfsPayload["tenant_name"] = common.GetTenantName()
	}

	// Setting the clone_url in the payload based on the API version
	// As per API spec For v3, use "clone_url" key, for others, use "cloneUrl" key
	layers, ok := cfsPayload["layers"].([]map[string]string)
	if !ok {
		common.Error(fmt.Errorf("failed to type-assert layers as []map[string]string"))
		return "", false
	}

	if apiVersion == "v3" {
		layers[0]["clone_url"] = configData.Clone_url
	} else {
		layers[0]["cloneUrl"] = configData.Clone_url
	}

	jsonPayload, err := json.Marshal(cfsPayload)
	if err != nil {
		common.Error(err)
		return "", false
	}

	common.Infof("CFS configuration payload: %s", string(jsonPayload))
	return string(jsonPayload), true
}

func CreateUpdateCFSConfigurationRecordAPI(cfgName, apiVersion, payload string, httpStatus int) (cfsConfig CFSConfiguration, ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return CFSConfiguration{}, false
	}

	// Set the payload for the request
	params.JsonStr = payload
	url := constructCFSURL("configurations", apiVersion) + "/" + cfgName
	resp, err := VerifyRestStatusWithTenant("PUT", url, *params, httpStatus)

	if err != nil {
		common.Error(err)
		return CFSConfiguration{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsConfig); err != nil {
		common.Error(err)
		return CFSConfiguration{}, false
	}

	ok = true
	return
}

func GetCFSConfigurationRecordAPI(cfgName, apiVersion string, httpStatus int) (cfsConfig CFSConfiguration, ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return CFSConfiguration{}, false
	}

	url := constructCFSURL("configurations", apiVersion) + "/" + cfgName
	resp, err := VerifyRestStatusWithTenant("GET", url, *params, httpStatus)

	if err != nil {
		common.Error(err)
		return CFSConfiguration{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsConfig); err != nil {
		common.Error(err)
		return CFSConfiguration{}, false
	}

	ok = true
	return
}

func GetCFSConfigurationsListAPIV2(apiVersion string) (cfsConfigurations []CFSConfiguration, ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return []CFSConfiguration{}, false
	}

	url := constructCFSURL("configurations", apiVersion)
	resp, err := VerifyRestStatusWithTenant("GET", url, *params, http.StatusOK)

	if err != nil {
		common.Error(err)
		return []CFSConfiguration{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsConfigurations); err != nil {
		common.Error(err)
		return []CFSConfiguration{}, false
	}

	ok = true
	return
}

func GetCFSConfigurationsListAPI(apiVersion string) (cfsConfigurations CFSConfigurationsList, ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return CFSConfigurationsList{}, false
	}

	url := constructCFSURL("configurations", apiVersion)
	resp, err := VerifyRestStatusWithTenant("GET", url, *params, http.StatusOK)

	if err != nil {
		common.Error(err)
		return CFSConfigurationsList{}, false
	}

	// Decoding the response body into the CFSConfiguration struct
	if err := json.Unmarshal(resp.Body(), &cfsConfigurations); err != nil {
		common.Error(err)
		return CFSConfigurationsList{}, false
	}

	ok = true
	return
}

func GetAPIBasedCFSConfigurationRecordList(apiVersion string) (cfsConfig []CFSConfiguration, ok bool) {
	// Get the CFS configurations list based on the API version. The response will be different for v2 and v3.
	// For v3, the response will be a list of CFSConfigurationsList, while for v2, it will be a CFSConfigurations struct.
	if apiVersion == "v3" {
		cfsConfigV3, ok := GetCFSConfigurationsListAPI(apiVersion)
		if !ok {
			return []CFSConfiguration{}, false
		}
		cfsConfig = cfsConfigV3.Configurations
	} else {
		cfsConfig, ok = GetCFSConfigurationsListAPIV2(apiVersion)
		if !ok {
			return []CFSConfiguration{}, false
		}
	}
	return cfsConfig, true
}

func DeleteCFSConfigurationRecordAPI(cfgName, apiVersion string, httpStatus int) (ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return false
	}

	url := constructCFSURL("configurations", apiVersion) + "/" + cfgName
	_, err := VerifyRestStatusWithTenant("DELETE", url, *params, httpStatus)

	if err != nil {
		common.Error(err)
		return false
	}

	ok = true
	return
}

func CFSConfigurationExists(cfsConfigurations []CFSConfiguration, cfgName string) (ok bool) {
	for _, cfsConfig := range cfsConfigurations {
		if cfsConfig.Name == cfgName {
			common.Infof("CFS configuration %s was found in the list of configurations", cfgName)
			return true
		}
	}
	common.Infof("CFS configuration %s was not found in the list of configurations", cfgName)
	return false
}

func VerifyCFSConfigurationRecord(cfsConfig CFSConfiguration, cfsPayload, cfgName, apiVersion string) (ok bool) {
	ok = true
	// Verify the CFS configuration record
	var cfsConfigPayload map[string]interface{}
	var cfsConfiglayerMap map[string]string

	// Decode the cfs configuration payload into a map[string]interface{}
	if err := json.Unmarshal([]byte(cfsPayload), &cfsConfigPayload); err != nil {
		common.Error(err)
		return false
	}

	cfsConfigPayloadLayerInterface := cfsConfigPayload["layers"].([]interface{})
	cfsConfigPayloadLayerInt := cfsConfigPayloadLayerInterface[0].(map[string]interface{})

	// Convert map[string]interface{} to map[string]string
	cfsConfigPayloadLayerMap := make(map[string]string)
	for key, value := range cfsConfigPayloadLayerInt {
		if strValue, ok := value.(string); ok {
			cfsConfigPayloadLayerMap[key] = strValue
		} else {
			common.Warnf("Skipping non-string value for key %s", key)
		}
	}

	//Extract cfs configuration layer data from the cfs configuration record

	if len(cfsConfig.Layers) > 0 {
		// Convert the first layer of cfsConfig to a map[string]string
		cfsConfiglayerMap = map[string]string{
			"commit":   cfsConfig.Layers[0].Commit,
			"playbook": cfsConfig.Layers[0].Playbook,
			"name":     cfsConfig.Layers[0].Name,
		}

		if apiVersion == "v3" {
			cfsConfiglayerMap["clone_url"] = cfsConfig.Layers[0].Clone_url
		} else {
			cfsConfiglayerMap["cloneUrl"] = cfsConfig.Layers[0].CloneURL
		}
	}

	// Verify the cfs configuration layer data
	passed := common.CompareMaps(cfsConfigPayloadLayerMap, cfsConfiglayerMap)
	if !passed {
		ok = false
		common.Errorf("CFS configuration layer data mismatch: expected %v, got %v", cfsConfigPayloadLayerMap, cfsConfiglayerMap)
	}

	if cfsConfig.Name != cfgName {
		common.Errorf("CFS configuration name mismatch: expected %s, got %s", cfgName, cfsConfig.Name)
		ok = false
	}

	return
}

func constructCFSURL(endpoint, apiVersion string) string {
	base := common.BASEURL + endpoints["cfs"][endpoint].Url
	if apiVersion != "" {
		return base + "/" + apiVersion + endpoints["cfs"][endpoint].Uri
	}
	return base + endpoints["cfs"][endpoint].Uri
}

func VerifyRestStatusWithTenant(method, uri string, params common.Params, expectedStatus int) (*resty.Response, error) {
	tenantName := common.GetTenantName()
	// Checking if Tenant name is set or not
	if len(tenantName) == 0 {
		return test.RestfulVerifyStatus(method, uri, params, expectedStatus)
	} else {
		if common.IsDummyTenant(tenantName) {
			// If the tenant name is a dummy tenant, we can skip the verification
			return test.TenantRestfulVerifyStatus(method, uri, tenantName, params, http.StatusBadRequest)
		}
		return test.TenantRestfulVerifyStatus(method, uri, tenantName, params, expectedStatus)
	}
}

func GetExpectedHTTPStatusCode() int {
	tenantName := common.GetTenantName()
	if common.IsDummyTenant(tenantName) {
		// If the tenant name is a dummy tenant, BadRequest is the expected status code
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func GetTenantFromList() string {
	tenantList, err := k8s.GetTenants()
	if err != nil {
		common.Errorf("Error getting tenant list: %s", err.Error())
		return ""
	}

	// Set the tenant name to be used in the tests
	tenantName, err := common.GetRandomStringFromList(tenantList)
	if err != nil {
		common.Debugf("Unable to get random tenant from list: %s", err.Error())
		return ""
	}
	common.Infof("Using tenant: %s", tenantName)
	return tenantName
}

func GetAnotherTenantFromList(currentTenant string) string {
	tenantList, err := k8s.GetTenants()
	if err != nil {
		common.Errorf("Error getting tenant list: %s", err.Error())
		return ""
	}

	// Set the tenant name to be used in the tests
	tenantName, err := common.GetRandomStringFromListExcept(tenantList, currentTenant)
	if err != nil {
		common.Debugf("Unable to get random tenant from list excluding %s: %s", currentTenant, err.Error())
		return ""
	}
	common.Infof("Using another tenant: %s", tenantName)
	return tenantName
}
