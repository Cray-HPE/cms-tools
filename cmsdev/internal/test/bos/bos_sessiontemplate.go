// MIT License
//
// (C) Copyright 2019-2023 Hewlett Packard Enterprise Development LP
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
 * bos sessiontemplate tests
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

const bosV1SessionTemplatesUri = bosV1BaseUri + "/sessiontemplate"
const bosV2SessionTemplatesUri = bosV2BaseUri + "/sessiontemplates"

const bosV1SessionTemplateTemplateUri = bosV1BaseUri + "/sessiontemplatetemplate"
const bosV2SessionTemplateTemplateUri = bosV2BaseUri + "/sessiontemplatetemplate"

const bosV1SessionTemplatesCLI = "sessiontemplate"
const bosV2SessionTemplatesCLI = "sessiontemplates"
const bosDefaultSessionTemplatesCLI = bosV2SessionTemplatesCLI

const bosV1SessionTemplateTemplateCLI = "sessiontemplatetemplate"
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

// The sessionTemplatesTestsURI and sessionTemplatesTestsCLICommand functions define the API and CLI versions of the BOS session template subtests.
// They both do the same thing:
// 1. List all session templates
// 2. Verify that this succeeds and returns something of the right general form
// 3. If the list returned is empty, then the subtest is over. Otherwise, select the first element of the list and extract the "name" field
// 4. Do a GET/describe on that particular session template
// 5. Verify that this succeeds and returns something of the right general form

func sessionTemplatesTestsAPI(params *common.Params, tenantList []string) (passed bool) {
	passed = true

	// session template template API tests
	// Just do a GET of the sessiontemplatetemplate endpoint and make sure that the response has
	// 200 status and a dictionary object

	// v1
	if !basicGetUriVerifyStringMapTest(bosV1SessionTemplateTemplateUri, params) {
		passed = false
	}

	// v2
	if !basicGetUriVerifyStringMapTest(bosV2SessionTemplateTemplateUri, params) {
		passed = false
	}

	// session template API tests

	// v1
	if !v1SessionTemplatesTestsURI(bosV1SessionTemplatesUri, params) {
		passed = false
	}

	// v2
	if !v2SessionTemplatesTestsURI(params, tenantList) {
		passed = false
	}

	return
}

func sessionTemplatesTestsCLI() (passed bool) {
	passed = true

	// session template template CLI tests
	// Make sure that "sessiontemplatetemplate list" CLI command succeeds and returns a dictionary object.

	// v1 sessiontemplatetemplate list
	if !basicCLIListVerifyStringMapTest("v1", bosV1SessionTemplateTemplateCLI) {
		passed = false
	}

	// v2 sessiontemplatetemplate list
	if !basicCLIListVerifyStringMapTest("v2", bosV2SessionTemplateTemplateCLI) {
		passed = false
	}

	// sessiontemplatetemplate list
	if !basicCLIListVerifyStringMapTest(bosDefaultSessionTemplateTemplateCLI) {
		passed = false
	}

	// session template CLI tests

	// v1 sessiontemplate
	if !sessionTemplatesTestsCLICommand("v1", bosV1SessionTemplatesCLI) {
		passed = false
	}

	// v2 sessiontemplates
	if !sessionTemplatesTestsCLICommand("v2", bosV2SessionTemplatesCLI) {
		passed = false
	}

	// sessiontemplates
	if !sessionTemplatesTestsCLICommand(bosDefaultSessionTemplatesCLI) {
		passed = false
	}

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

// Performs an API query to list BOS v2 sessions (possibly with a tenant specified)
// Parses the result to convert it to a list of dictionaries with string keys
// Returns the result and error (if any)
func listV2SessionTemplatesApi(params *common.Params, tenantName string) (dictList []map[string]interface{}, err error) {
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
// parses that dictionaries into a v2TemplateData struct.
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

// Gets a list of V2 session template dictionary objects using listV2SessionTemplatesApi,
// parses those dictionaries into v2TemplateData structs.
// If a tenant was specified, validate that every session template belongs to that tenant.
// Returns that list of structs and error (if any)
func listV2TemplateDatasApi(params *common.Params, tenantName string) (templateDataList []v2TemplateData, err error) {
	var dictList []map[string]interface{}
	var templateData v2TemplateData

	dictList, err = listV2SessionTemplatesApi(params, tenantName)
	if err != nil {
		return
	}
	templateDataList = make([]v2TemplateData, 0, len(dictList))
	for sessionTemplateIndex, sessionTemplateDict := range dictList {
		templateData, err = parseV2TemplateData(sessionTemplateDict)
		if err != nil {
			err = fmt.Errorf("%w; Error parsing session template #%d in list", err, sessionTemplateIndex)
			return
		}
		// If a tenant was specified, validate that session template belongs to the expected tenant
		if len(tenantName) > 0 && templateData.Tenant != tenantName {
			err = fmt.Errorf("SessionTemplate #%d in the list (%s) does not belong to expected tenant '%s'",
				sessionTemplateIndex, templateData.String(), tenantName)
			return
		}
		templateDataList = append(templateDataList, templateData)
	}
	return
}

// v2 session templates API tests
// See comment earlier in the file for a general description of this function
//
// The v2 API version of this function has had further improvements made in order to test good-path multitenancy
// queries. Specifically, this function now does the following (all using API calls)
// 1. API query to list all BOS v2 session templates (no tenant name specified)
// 2. Parses the result, verifying that it is a JSON list, and verifying that each item in that list:
//    a. Is a JSON dictionary
//    b. Has a "name" field which maps to a non-0-length string
//    c. Either has no "tenant" field or has a "tenant" field that maps to a (possibly 0-length) string
// 3. If the list is empty, then pick a random tenant name (possibly not one which exists on the system)
//    and issue a query for all BOS session templates belonging to that tenant. Verify that the resulting list is empty.
// 4. If the list from #1 is not empty, but no session templates belong to a tenant, then do #3.
// 5. If the list from #1 is not empty and has session templates owned by a tenant, then pick one such tenant, pick above
//    session template owned by that tenant, and also count how many total session templates are owned by that tenant.
//    a. Query for all BOS session templates belonging to that tenant and verify that we get the same total number. Also
//       perform the same validation steps as in #2, and validate that every returned session template belongs to the specified tenant.
//    b. Query for the specific BOS session template belonging to that tenant and verify that it can be retrieved.
// 6. If the list from #1 is not empty and has at least one session template that is not owned by a tenant, then pick such
//    a session template, query BOS for that specific session template, and verify that it can be retrieved.

// Several of these checks could fail if BOS session templates are being created or deleted while this test is running.
// For now, this will be a documented limitation of the test, with the recommendation to re-run just
// the BOS health check in the case that certain failures are seen. Eventually the test could be improved to automatically
// retry the relevant checks a limited number of times, to reduce the likelihood of false failures.

func v2SessionTemplatesTestsURI(params *common.Params, tenantList []string) bool {
	var tenantedTemplateData, untenantedTemplateData v2TemplateData
	var tenantedTemplateCount int
	var err error
	var templateDataList []v2TemplateData

	// test #1, list sessions
	templateDataList, err = listV2TemplateDatasApi(params, "")
	if err != nil {
		common.Error(err)
		return false
	} else if len(templateDataList) == 0 {
		common.Infof("skipping test GET %s/{session_template_id}", bosV2SessionTemplatesUri)
		common.Infof("results from previous test is []")

		// However, we can still try to list all of the session templates with a tenant name specified.
		// Since no session templates were found from the un-tenanted query, we expect none to be found once a tenant
		// is specified
		tenant := getAnyTenant(tenantList)
		templateDataList, err = listV2TemplateDatasApi(params, tenant)
		if err != nil {
			common.Error(err)
			return false
		}

		// Validate that this list is empty
		if len(templateDataList) > 0 {
			common.Errorf("Listing of all session templates empty, but found %d when listing session templates for tenant '%s'",
				len(templateDataList), tenant)
			return false
		}
		common.Infof("Session template list is still empty, as expected")
		return true
	}

	// From the session template ID list, we want to identify:
	// * One session template that has no tenant (untenantedTemplateData)
	// * One session template that belongs to a tenant (tenantedTemplateData), and a count (tenantedTemplateCount)
	//   of how many session templates in total have that same tenant
	for sessionTemplateIndex, TemplateData := range templateDataList {
		common.Debugf("Parsing session #%d in the session list (%s)", sessionTemplateIndex, TemplateData.String())
		if len(TemplateData.Tenant) == 0 {
			// This session template has no tenant
			if untenantedTemplateData.IsNil() {
				// This is the first session template we have encountered that has no tenant field,
				// so take note of it
				untenantedTemplateData = TemplateData
				common.Infof("Found BOS v2 session template #%d (%s)", sessionTemplateIndex, untenantedTemplateData.String())
			}
			continue
		}
		// This session template is owned by a tenant.
		if tenantedTemplateData.IsNil() {
			// This is the first session template we've found that is owned by a tenant, so remember it,
			// and note that we have found 1 session template belonging to this tenant so far
			tenantedTemplateData = TemplateData
			tenantedTemplateCount = 1
			common.Infof("Found BOS v2 session template #%d (%s)", sessionTemplateIndex, tenantedTemplateData.String())
			continue
		}
		// We have already found a session template belonging to a tenant. If it is the tenant we found first,
		// increment our session count
		if TemplateData.Tenant == tenantedTemplateData.Tenant {
			tenantedTemplateCount += 1
		}
	}

	passed := true

	if untenantedTemplateData.IsNil() {
		common.Infof("skipping test GET %s/{session_template_id} with no tenant specified, because all BOS v2 session templates are owned by tenants",
			bosV2SessionTemplatesUri)
	} else {
		// test: describe session template using the untenanted session template name we found earlier
		_, err = getV2TemplateDataApi(params, untenantedTemplateData)
		if err != nil {
			common.Error(err)
			passed = false
		}
	}

	if tenantedTemplateData.IsNil() {
		common.Infof("No BOS v2 session templates found belonging to any tenants")

		tenant := getAnyTenant(tenantList)
		templateDataList, err = listV2TemplateDatasApi(params, tenant)
		if err != nil {
			common.Error(err)
			passed = false
		} else {
			// Validate that this list is empty
			if len(templateDataList) > 0 {
				common.Errorf("Listing all session templates found none owned by tenants, but found %d when listing session templates for tenant '%s'",
					len(templateDataList), tenant)
				passed = false
			}
			common.Infof("List of session templates belonging to tenant '%s' is empty, as expected", tenant)
		}
		return passed
	}
	common.Infof("Counted %d BOS v2 session templates belonging to tenant '%s'", tenantedTemplateCount, tenantedTemplateData.Tenant)
	templateDataList, err = listV2TemplateDatasApi(params, tenantedTemplateData.Tenant)
	if err != nil {
		common.Error(err)
		passed = false
	} else {
		// Validate that this list is expected length
		if len(templateDataList) != tenantedTemplateCount {
			common.Errorf("Listing all session templates found %d owned by tenant '%s', but found %d when listing session templates for that tenant",
				tenantedTemplateCount, tenantedTemplateData.Tenant, len(templateDataList))
			passed = false
		}
		common.Infof("List of session templates belonging to tenant '%s' is the expected length", tenantedTemplateData.Tenant)
	}

	// test: describe session template using the session template name we found owned by a tenant in the earlier loop
	_, err = getV2TemplateDataApi(params, tenantedTemplateData)
	if err != nil {
		common.Error(err)
		return false
	}

	return passed
}

// Given a response object (as an array of bytes), do the following:
// 1. Verify that it is a JSON list object
// 2. If the list object is empty, return a blank string.
// 3. If the list is not empty, verify that its first element is a dictionary.
// 4. Look up the "name" key in that dictionary, and return its value.
// If any of the above does not work, return an appropriate error.
func getFirstTemplateData(listCmdOut []byte) (string, error) {
	return common.GetStringFieldFromFirstItem("name", listCmdOut)
}

// Given a response object (as an array of bytes), validate that:
// 1. It resolves to a JSON dictonary
// 2. That dictionary has a "name" field
// 3. The "name" field of that dictionary has a value which matches our expectedName string
//
// Return true if all of the above is true. Otherwise, log an appropriate error and return false.
func ValidateTemplateData(mapCmdOut []byte, expectedName string) bool {
	err := common.ValidateStringFieldValue("BOS sessiontemplate", "name", expectedName, mapCmdOut)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// session templates API tests
// See comment earlier in the file for a description of this function
func v1SessionTemplatesTestsURI(uri string, params *common.Params) bool {
	// test #1, list session templates
	common.Infof("GET %s test scenario", uri)
	resp, err := bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}

	// use results from previous test, grab the first session template
	templateData, err := getFirstTemplateData(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(templateData) == 0 {
		common.Infof("skipping test GET %s/{session_template_id}", uri)
		common.Infof("results from previous test is []")
		return true
	}

	// a session_template_id is available
	uri += "/" + templateData
	common.Infof("GET %s test scenario", uri)
	resp, err = bosRestfulVerifyStatus("GET", uri, params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	} else if !ValidateTemplateData(resp.Body(), templateData) {
		return false
	}

	return true
}

// session templates CLI tests
// See comment earlier in the file for a description of this function
func sessionTemplatesTestsCLICommand(cmdArgs ...string) bool {
	// test #1, list session templates
	cmdOut := runBosCLIList(cmdArgs...)
	if cmdOut == nil {
		return false
	}

	// use results from previous test, grab the first session template
	templateData, err := getFirstTemplateData(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	} else if len(templateData) == 0 {
		common.Infof("skipping test CLI describe session template {session_template_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// a session template id is available
	// test #2 describe session templates
	cmdOut = runBosCLIDescribe(templateData, cmdArgs...)
	if cmdOut == nil {
		return false
	} else if !ValidateTemplateData(cmdOut, templateData) {
		return false
	}

	return true
}
