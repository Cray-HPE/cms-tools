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
 * bos_sessiontemplates_api.go
 *
 * BOS sessiontemplates API.
 *
 */
package bos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	resty "gopkg.in/resty.v1"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

/*
 * bos_helper.go
 *
 * BOS test helper file
 *
 */

var DEFAULT_PLAYBOOK = "compute_nodes.yml"
var archMap = map[string]string{
	"X86": "x86_64",
	"ARM": "aarch64",
}

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

func GetLatestImageIdFromCsmProductCatalog(arch string) (string, error) {
	csmProductCatalogData, err := GetCsmProductCatalogData()
	if err != nil {
		return "", err
	}

	// Get the latest CSM version data
	csmVersion, err := GetCSMLatestVersion(csmProductCatalogData)
	if err != nil {
		return "", err
	}

	lastestCSMDataRaw := csmProductCatalogData[csmVersion]

	common.Debugf("Latest CSM data: %v", lastestCSMDataRaw)

	latestCSMData := lastestCSMDataRaw.(map[interface{}]interface{})

	// Access the "images" field from the unmarshaled map
	csmImages, ok := latestCSMData["images"].(map[interface{}]interface{})
	if !ok {
		return "", fmt.Errorf("failed to retrieve 'images' field from latest CSM data")
	}
	common.Infof("CSM images: %v", csmImages)
	for key := range csmImages {
		// Return the first key found
		if strings.Contains(key.(string), archMap[arch]) {
			common.Debugf("Found image ID for architecture %s: %s", archMap[arch], csmImages[key])
			return csmImages[key].(map[interface{}]interface{})["id"].(string), nil
		}
	}
	return "", fmt.Errorf("no image found in the latest CSM data")
}

func GetImageRecord(imageID string) (imagerecord ImsImage, ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return ImsImage{}, false
	}

	url := common.BASEURL + endpoints["ims"]["images"].Url + endpoints["ims"]["images"].Uri + "/" + imageID
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return ImsImage{}, false
	}

	// Unmarshal the response into the ImsImage struct
	if err := json.Unmarshal(resp.Body(), &imagerecord); err != nil {
		common.Error(err)
		return ImsImage{}, false
	}

	ok = true
	return
}

func GetCreateBOSSessionTemplatePayload(cfsConfigName string, enableCFS bool, arch string, imageId string) (bosSessionTemplatePayload string, ok bool) {
	// Get the latest image ID from the CSM product catalog
	// and the image record from the IMS API based on architecture
	// imageId, err := GetLatestImageIdFromCsmProductCatalog(arch)
	// if err != nil {
	// 	common.Error(err)
	// 	return "", false
	// }
	// common.Infof("Using image ID %s with arch: %s", imageId, arch)
	imageRecord, ok := GetImageRecord(imageId)
	if !ok {
		return "", false
	}
	kernelParameters :=
		"console=ttyS0,115200 bad_page=panic crashkernel=512M hugepagelist=2m-2g " +
			"intel_iommu=off intel_pstate=disable iommu.passthrough=on " +
			"modprobe.blacklist=amdgpu numa_interleave_omit=headless oops=panic pageblock_order=14 " +
			"rd.neednet=1 rd.retry=10 rd.shell split_lock_detect=off " +
			"systemd.unified_cgroup_hierarchy=1 ip=dhcp quiet spire_join_token=${SPIRE_JOIN_TOKEN} " +
			fmt.Sprintf("root=live:s3://boot-images/%s/rootfs ", imageRecord.ImageID) +
			fmt.Sprintf("nmd_data=url=s3://boot-images/%s/rootfs,etag=%s", imageRecord.ImageID, imageRecord.Link.S3_Etag)

	computeSet := map[string]interface{}{
		"etag":              imageRecord.Link.S3_Etag,
		"kernel_parameters": kernelParameters,
		"node_roles_groups": []string{"Compute"},
		"path":              imageRecord.Link.S3_Path,
		"type":              "s3",
		"arch":              arch,
	}

	cfsConfig := map[string]string{
		"configuration": cfsConfigName,
	}

	bosParams := map[string]interface{}{
		"cfs":        cfsConfig,
		"enable_cfs": enableCFS,
		"boot_sets": map[string]interface{}{
			"compute": computeSet,
		},
	}

	jsonPayload, err := json.Marshal(bosParams)
	if err != nil {
		common.Error(err)
		return "", false
	}

	return string(jsonPayload), true
}

func CreateUpdateBOSSessiontemplatesAPI(bosSessionTemplatePayload string, sessionTemplateName string, method string) (bosSessionTemplate BOSSessionTemplate, ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return BOSSessionTemplate{}, false
	}

	// Create the BOS session template payload
	if method == "PUT" {
		// Create session template payload
		params.JsonStr = bosSessionTemplatePayload
	} else {
		// Update session template payload
		params.JsonStrArray = []byte(bosSessionTemplatePayload)
	}

	url := common.BASEURL + endpoints["bos"]["sessiontemplates"].Url +
		"/" + endpoints["bos"]["sessiontemplates"].Version +
		endpoints["bos"]["sessiontemplates"].Uri + "/" + sessionTemplateName

	resp, err := VerifyRestStatusWithTenant(method, url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return BOSSessionTemplate{}, false
	}

	// Decode the response body into the BOSSessionTemplate struct
	if err := json.Unmarshal(resp.Body(), &bosSessionTemplate); err != nil {
		common.Error(err)
		return BOSSessionTemplate{}, false
	}

	ok = true
	return
}

func DeleteBOSSessionTemplatesAPI(sessionTemplateName string) (ok bool) {
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Error(fmt.Errorf("Unable to get access token params"))
		return false
	}

	url := common.BASEURL + endpoints["bos"]["sessiontemplates"].Url +
		"/" + endpoints["bos"]["sessiontemplates"].Version +
		endpoints["bos"]["sessiontemplates"].Uri + "/" + sessionTemplateName
	_, err := VerifyRestStatusWithTenant("DELETE", url, *params, http.StatusNoContent)
	if err != nil {
		common.Error(err)
		return false
	}

	ok = true
	return
}

func ValidateBOSSessionTemplateAPI(templateName string) (ok bool) {
	common.Infof("Validating BOS sessiontemplate '%s'", templateName)
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Errorf("Unable to get access token params")
		return false
	}

	url := common.BASEURL + endpoints["bos"]["sessiontemplatesvalid"].Url +
		"/" + endpoints["bos"]["sessiontemplatesvalid"].Version +
		endpoints["bos"]["sessiontemplatesvalid"].Uri + "/" + templateName
	resp, err := VerifyRestStatusWithTenant("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	if strings.Contains(string(resp.Body()), "Valid") {
		common.Infof("Session template %s is valid", templateName)
		ok = true
	} else {
		common.Errorf("Session template %s is not valid", templateName)
		ok = false
	}

	return

}

func GetBOSSessionTemplatesAPI(templateName string, httpStatus int) (bosSessionTemplate BOSSessionTemplate, ok bool) {
	common.Infof("Getting BOS sessiontemplate '%s'", templateName)
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Errorf("Unable to get access token params")
		return BOSSessionTemplate{}, false
	}

	url := common.BASEURL + endpoints["bos"]["sessiontemplates"].Url +
		"/" + endpoints["bos"]["sessiontemplates"].Version +
		endpoints["bos"]["sessiontemplates"].Uri + "/" + templateName
	resp, err := VerifyRestStatusWithTenant("GET", url, *params, httpStatus)
	if err != nil {
		common.Error(err)
		return BOSSessionTemplate{}, false
	}

	// Decode the response body into the BOSSessionTemplate struct
	if err := json.Unmarshal(resp.Body(), &bosSessionTemplate); err != nil {
		common.Error(err)
		return BOSSessionTemplate{}, false
	}

	ok = true
	return
}

func GetAllBOSSessionTemplatesAPI() (bosSessionTemplates []BOSSessionTemplate, ok bool) {
	common.Infof("Getting all BOS sessiontemplates")
	params := test.GetAccessTokenParams()
	if params == nil {
		common.Errorf("Unable to get access token params")
		return []BOSSessionTemplate{}, false
	}

	url := common.BASEURL + endpoints["bos"]["sessiontemplates"].Url +
		"/" + endpoints["bos"]["sessiontemplates"].Version +
		endpoints["bos"]["sessiontemplates"].Uri
	resp, err := VerifyRestStatusWithTenant("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return []BOSSessionTemplate{}, false
	}

	// Decode the response body into the BOSSessionTemplate struct
	if err := json.Unmarshal(resp.Body(), &bosSessionTemplates); err != nil {
		common.Error(err)
		return []BOSSessionTemplate{}, false
	}

	ok = true
	return
}

func BOSSessionTemplateExists(sessionTemplateName string, templateList []BOSSessionTemplate) (ok bool) {
	for _, template := range templateList {
		if template.Name == sessionTemplateName {
			return true
		}
	}
	common.Infof("Sessiontemplate %s was not found in the list of sessiontemplates", sessionTemplateName)
	return false
}

func VerifyBOSSessionTemplate(sessionTemplate BOSSessionTemplate, sessionTemplatePayload string, templateName string) (ok bool) {
	// Unmarshal the session template payload into a BOSSessionTemplate struct
	var expectedSessionTemplate BOSSessionTemplate
	if err := json.Unmarshal([]byte(sessionTemplatePayload), &expectedSessionTemplate); err != nil {
		common.Errorf("Error unmarshaling session template payload: %v", err)
		return false
	}

	if sessionTemplate.Name != templateName {
		common.Errorf("Session template name does not match. Expected: %s, Got: %s", templateName, sessionTemplate.Name)
		return false
	}

	if sessionTemplate.Enable_cfs != expectedSessionTemplate.Enable_cfs {
		common.Errorf("Enable CFS does not match. Expected: %v, Got: %v", expectedSessionTemplate.Enable_cfs, sessionTemplate.Enable_cfs)
		return false
	}

	if sessionTemplate.Boot_sets.Compute.Kernel_parameters != expectedSessionTemplate.Boot_sets.Compute.Kernel_parameters {
		common.Errorf("Kernel parameters do not match. Expected: %s, Got: %s", expectedSessionTemplate.Boot_sets.Compute.Kernel_parameters, sessionTemplate.Boot_sets.Compute.Kernel_parameters)
		return false
	}

	if sessionTemplate.Boot_sets.Compute.Etag != expectedSessionTemplate.Boot_sets.Compute.Etag {
		common.Errorf("Etag does not match. Expected: %s, Got: %s", expectedSessionTemplate.Boot_sets.Compute.Etag, sessionTemplate.Boot_sets.Compute.Etag)
		return false
	}

	if sessionTemplate.Boot_sets.Compute.Path != expectedSessionTemplate.Boot_sets.Compute.Path {
		common.Errorf("Path does not match. Expected: %s, Got: %s", expectedSessionTemplate.Boot_sets.Compute.Path, sessionTemplate.Boot_sets.Compute.Path)
		return false
	}

	if sessionTemplate.Boot_sets.Compute.Type != expectedSessionTemplate.Boot_sets.Compute.Type {
		common.Errorf("Type does not match. Expected: %s, Got: %s", expectedSessionTemplate.Boot_sets.Compute.Type, sessionTemplate.Boot_sets.Compute.Type)
		return false
	}

	if sessionTemplate.Boot_sets.Compute.Node_roles_groups[0] != expectedSessionTemplate.Boot_sets.Compute.Node_roles_groups[0] {
		common.Errorf("Node roles groups do not match. Expected: %s, Got: %s", expectedSessionTemplate.Boot_sets.Compute.Node_roles_groups[0], sessionTemplate.Boot_sets.Compute.Node_roles_groups[0])
		return false
	}

	return true
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
		// If the tenant name is a dummy tenant, the expected response is StatusBadRequest
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
		common.Errorf("Error getting random tenant from list: %s", err.Error())
		return ""
	}
	common.Infof("Using tenant: %s", tenantName)
	return tenantName
}
