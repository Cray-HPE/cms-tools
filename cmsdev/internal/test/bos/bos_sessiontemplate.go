// MIT License
//
// (C) Copyright 2019-2024 Hewlett Packard Enterprise Development LP
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
package bos

/*
 * bos_sessiontemplate.go
 *
 * BOS sessiontemplate tests utility functions
 *
 */

import (
	"errors"
	"fmt"
	resty "gopkg.in/resty.v1"
	"net/http"
	"reflect"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

const bosV2SessionTemplatesUri = bosV2BaseUri + "/sessiontemplates"

const bosV2SessionTemplateTemplateUri = bosV2BaseUri + "/sessiontemplatetemplate"

const bosV2SessionTemplatesCLI = "sessiontemplates"
const bosDefaultSessionTemplatesCLI = bosV2SessionTemplatesCLI

const bosV2SessionTemplateTemplateCLI = "sessiontemplatetemplate"
const bosDefaultSessionTemplateTemplateCLI = bosV2SessionTemplateTemplateCLI

// A BOS v2 session template is uniquely identified by its name and tenant. The tenant may null, blank,
// or, equivalently, not present in the actual record. In all such cases, the
// following struct uses an empty string value for Tenant.
type v2TemplateData struct {
	Name, Tenant string
}

// Returns true if the Name field of the object is "", false otherwise
func (templateData v2TemplateData) IsNil() bool {
	return templateData.Name == ""
}

// Returns a string representation of the object
func (templateData v2TemplateData) String() string {
	if len(templateData.Tenant) > 0 {
		return fmt.Sprintf("name: '%s', tenant: '%s'", templateData.Name, templateData.Tenant)
	}
	return fmt.Sprintf("name: '%s', no tenant", templateData.Name)
}

// Compares two objects. Returns true if both fields match, false otherwise. If false,
// appropriate errors are also logged noting the discrepancies
func (templateData v2TemplateData) HasExpectedValues(expectedTemplateData v2TemplateData) (passed bool) {
	// Let's be optimistic and assume that they will match
	passed = true

	// Validate that name fields match
	if templateData.Name != expectedTemplateData.Name {
		common.Errorf("BOS session template name '%s' does not match expected name '%s'", templateData.Name, expectedTemplateData.Name)
		passed = false
	} else {
		common.Debugf("BOS session template name '%s' matches expected value", templateData.Name)
	}

	// Validate the tenant name
	if templateData.Tenant == expectedTemplateData.Tenant {
		if len(templateData.Tenant) == 0 {
			common.Debugf("Session template does not belong to a tenant, which matches expectations")
		} else {
			common.Debugf("Session template belongs to the expected tenant")
		}
		return
	}
	passed = false
	if len(templateData.Tenant) == 0 {
		common.Errorf("Session template does not belong to a tenant, but it should belong to '%s'", expectedTemplateData.Tenant)
	} else {
		common.Errorf("Session template belongs to tenant '%s', but it should not belong to any tenant", templateData.Tenant)
	}
	return
}

// Takes as input a mapping from strings to arbitrary objects.
// Parses it to extract the values of the 'name' and 'tenant' fields (although the
// latter is allowed to be absent). Validates that 'name' (and 'tenant', if present and non-null)
// have string values and that 'name' is non-0 length. Returns a v2TemplateData object
// populated from those fields. Returns an error with any problems encountered.
func parseV2TemplateData(sessionTemplateDict map[string]interface{}) (templateData v2TemplateData, totalErr error) {
	var err error
	var fieldFound, fieldIsString bool
	var sessionTemplateTenantField interface{}

	common.Debugf("Getting name of session template")
	templateData.Name, err = common.GetStringFieldFromMapObject("name", sessionTemplateDict)
	if err != nil {
		err = fmt.Errorf("%w; Error getting 'name' field of BOS session template", err)
	} else if len(templateData.Name) == 0 {
		err = fmt.Errorf("BOS session template has a 0-length name")
	} else {
		common.Debugf("Name of session template is '%s'", templateData.Name)
	}
	totalErr = errors.Join(totalErr, err)

	common.Debugf("Checking for 'tenant' field of session template '%s'", templateData.Name)
	sessionTemplateTenantField, fieldFound = sessionTemplateDict["tenant"]
	if !fieldFound {
		// This session template has no tenant field, which is equivalent to a 0-length string tenant field
		templateData.Tenant = ""
		common.Debugf("Session template '%s' has no 'tenant' field", templateData.Name)
		return
	}
	if sessionTemplateTenantField == nil {
		// This session template has null-value tenant field, which is equivalent to a 0-length string tenant field
		templateData.Tenant = ""
		common.Debugf("Session template '%s' has null 'tenant' field", templateData.Name)
		return
	}

	// If it is present and non-null, it should be a string value
	templateData.Tenant, fieldIsString = sessionTemplateTenantField.(string)
	if !fieldIsString {
		// tenant field has non-string value
		err = fmt.Errorf("Session template '%s' has a non-null 'tenant' field but its value is type %s, not string",
			templateData.Name, reflect.TypeOf(sessionTemplateTenantField).String())
		totalErr = errors.Join(totalErr, err)
	}
	common.Debugf("Session template: %s", templateData.String())
	return
}

// Performs an API query to get a particular BOS v2 session template (possibly belonging to a tenant)
// Parses the result to convert it to a dictionary with string keys
// Returns the result and error (if any)
func getV2TemplateApi(params *common.Params, templateData v2TemplateData) (sessionTemplateDict map[string]interface{}, err error) {
	var resp *resty.Response
	uri := bosV2SessionTemplatesUri + "/" + templateData.Name
	if len(templateData.Tenant) == 0 {
		common.Infof("GET %s test scenario", uri)
		resp, err = bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	} else {
		common.Infof("GET %s (tenant: %s) test scenario", uri, templateData.Tenant)
		resp, err = bosTenantRestfulVerifyStatus("GET", uri, templateData.Tenant, params, http.StatusOK)
	}
	if err != nil {
		return
	}
	// Decode JSON into a string map
	sessionTemplateDict, err = common.DecodeJSONIntoStringMap(resp.Body())
	return
}

// Performs an API query to list BOS v2 templates (possibly with a tenant specified)
// Parses the result to convert it to a list of dictionaries with string keys
// Returns the result and error (if any)
func listV2TemplatesApi(params *common.Params, tenantName string) (dictList []map[string]interface{}, err error) {
	var resp *resty.Response
	if len(tenantName) == 0 {
		common.Infof("GET %s test scenario", bosV2SessionTemplatesUri)
		resp, err = bosRestfulVerifyStatus("GET", bosV2SessionTemplatesUri, params, http.StatusOK)
	} else {
		common.Infof("GET %s (tenant: %s) test scenario", bosV2SessionTemplatesUri, tenantName)
		resp, err = bosTenantRestfulVerifyStatus("GET", bosV2SessionTemplatesUri, tenantName, params, http.StatusOK)
	}
	if err != nil {
		return
	}
	// Decode JSON into a list of string maps
	dictList, err = common.DecodeJSONIntoStringMapList(resp.Body())
	return
}

// Get a particular BOS v2 session template (possibly belonging to a tenant) using getV2TemplateApi,
// parse that dictionary into a v2TemplateData struct.
// Validates that it matches the expected session template name and (if any) tenant name.
// Returns that struct and error (if any)
func getV2TemplateDataApi(params *common.Params, templateData v2TemplateData) (templateDataFromApi v2TemplateData, err error) {
	var sessionTemplateDict map[string]interface{}

	sessionTemplateDict, err = getV2TemplateApi(params, templateData)
	if err != nil {
		return
	}
	templateDataFromApi, err = parseV2TemplateData(sessionTemplateDict)
	if err != nil {
		return
	}
	if !templateDataFromApi.HasExpectedValues(templateData) {
		err = fmt.Errorf("SessionTemplate returned by API query (%s) does not match session template requested (%s)",
			templateDataFromApi.String(), templateData.String())
	}
	return
}

// Get a particular BOS v2 session template (possibly belonging to a tenant) using bosTenantDescribeCli,
// parse that dictionary into a v2TemplateData struct.
// Validates that it matches the expected session template name and (if any) tenant name.
// Returns that struct and error (if any)
func describeV2TemplateDataCli(templateData v2TemplateData, cmdArgs ...string) (templateDataFromCli v2TemplateData, passed bool) {
	var sessionTemplateDict map[string]interface{}
	var ok bool
	var err error

	passed = false
	sessionTemplateDict, ok = bosTenantDescribeCli(templateData.Tenant, templateData.Name, cmdArgs...)
	if !ok {
		return
	}
	templateDataFromCli, err = parseV2TemplateData(sessionTemplateDict)
	if err != nil {
		return
	}
	if !templateDataFromCli.HasExpectedValues(templateData) {
		err = fmt.Errorf("SessionTemplate returned by CLI command (%s) does not match session template requested (%s)",
			templateDataFromCli.String(), templateData.String())
	} else {
		passed = true
	}
	return
}

// Parses list of string-key dictionaries into a list of v2TemplateData structs.
// If a tenant was specified, validate that every template belongs to that tenant.
// Returns that list of structs and error (if any)
func dictListToTemplateDataList(dictList []map[string]interface{}, tenantName string) (templateDataList []v2TemplateData, err error) {
	var templateData v2TemplateData
	templateDataList = make([]v2TemplateData, 0, len(dictList))
	for templateIndex, templateDict := range dictList {
		templateData, err = parseV2TemplateData(templateDict)
		if err != nil {
			err = fmt.Errorf("%w; Error parsing template #%d in list", err, templateIndex)
			return
		}
		// If a tenant was specified, validate that template belongs to the expected tenant
		if len(tenantName) > 0 && templateData.Tenant != tenantName {
			err = fmt.Errorf("Template #%d in the list (%s) does not belong to expected tenant '%s'",
				templateIndex, templateData.String(), tenantName)
			return
		}
		templateDataList = append(templateDataList, templateData)
	}
	return
}

// Gets a list of session template dictionary objects using listV2TemplatesApi,
// converts that list using dictListToTemplateDataList. Returns that list of structs and error (if any)
func listV2TemplateDataApi(params *common.Params, tenantName string) (templateDataList []v2TemplateData, err error) {
	var dictList []map[string]interface{}

	dictList, err = listV2TemplatesApi(params, tenantName)
	if err != nil {
		return
	}
	templateDataList, err = dictListToTemplateDataList(dictList, tenantName)
	return
}

// Gets a list of session template dictionary objects using bosTenantListCli,
// converts that list using dictListToTemplateDataList. Returns resulting list and boolean
// value indicating whether the function passed or failed (an error will have been logged in the case
// of failure)
func listTemplateDataCli(tenantName string, cmdArgs ...string) (templateDataList []v2TemplateData, passed bool) {
	var dictList []map[string]interface{}
	var err error

	dictList, passed = bosTenantListCli(tenantName, cmdArgs...)
	if !passed {
		return
	}
	templateDataList, err = dictListToTemplateDataList(dictList, tenantName)
	if err != nil {
		common.Error(err)
		passed = false
	}
	return
}

// Given an array of bytes, validate that:
// 1. It resolves to a JSON dictonary
// 2. The resulting dictionary can be successfully parsed as a V2 template object (using parseV2TemplateData function)
// 3. The resulting V2 template object has the expected ID values (using HasExpectedValues method)
//
// Return true if all of the above is true. Otherwise, log an appropriate error and return false.
func ValidateTemplateData(mapCmdOut []byte, expectedId v2TemplateData) bool {
	common.Infof("Validating that BOS session template has expected values for name and tenant")

	// This function should always receive a non-0-length expected name
	if len(expectedId.Name) == 0 {
		common.Errorf("Programming logic error: ValidateTemplateData function received 0-length string for expected name")
		return false
	}

	// Should be a dictionary object mapping strings to values
	common.Debugf("Parsing session template as a dictionary")
	templateDict, err := common.DecodeJSONIntoStringMap(mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}

	templateData, err := parseV2TemplateData(templateDict)
	if err != nil {
		common.Error(err)
		return false
	}

	return templateData.HasExpectedValues(expectedId)
}
