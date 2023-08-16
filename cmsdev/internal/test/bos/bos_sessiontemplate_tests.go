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
 * bos_sessiontemplate_tests.go
 *
 * BOS sessiontemplate tests
 *
 */

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"strings"
)

// The sessionTemplatesTestsURI and sessionTemplatesTestsCLICommand functions define the API and CLI versions of the BOS session template subtests.
// They both basically do the same thing, although the V2 API test has some enhancements described later in the file.
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
	if !v1SessionTemplatesTestsURI(params) {
		passed = false
	}

	// v2
	if !v2SessionTemplatesTestsURI(params, tenantList) {
		passed = false
	}

	return
}

func sessionTemplatesTestsCLI(tenantList []string) (passed bool) {
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
	if !v1SessionTemplatesTestsCLICommand("v1", bosV1SessionTemplatesCLI) {
		passed = false
	}

	// v2 sessiontemplates
	if !v2SessionTemplatesTestsCLICommand(tenantList, "v2", bosV2SessionTemplatesCLI) {
		passed = false
	}

	// sessiontemplates
	if !v2SessionTemplatesTestsCLICommand(tenantList, bosDefaultSessionTemplatesCLI) {
		passed = false
	}

	return
}

// v1 session templates API test
func v1SessionTemplatesTestsURI(params *common.Params) bool {
	var untenantedTemplateData v2TemplateData
	var err error
	var templateDataList []v2TemplateData

	// test #1, list templates
	templateDataList, err = listV1TemplateDataApi(params)
	if err != nil {
		common.Error(err)
		return false
	} else if len(templateDataList) == 0 {
		common.Infof("skipping test GET %s/{session_template_id}", bosV1SessionTemplatesUri)
		common.Infof("results from previous test is []")
		return true
	}

	// From the session template ID list, we want to identify:
	// * One session template that has no tenant (untenantedTemplateData)
	for sessionTemplateIndex, TemplateData := range templateDataList {
		common.Debugf("Parsing session #%d in the session list (%s)", sessionTemplateIndex, TemplateData.String())
		if len(TemplateData.Tenant) != 0 {
			continue
		}
		// This is the first session template we have encountered that has no tenant field,
		// so take note of it
		untenantedTemplateData = TemplateData
		common.Infof("Found BOS session template #%d (%s)", sessionTemplateIndex, untenantedTemplateData.String())
		break
	}

	if untenantedTemplateData.IsNil() {
		common.Infof("skipping test GET %s/{session_template_id} with no tenant specified, because all BOS session templates are owned by tenants",
			bosV1SessionTemplatesUri)
		return true
	}
	// test: describe session template using the untenanted session template name we found earlier
	_, err = getV1TemplateDataApi(params, untenantedTemplateData.Name)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
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

	// test #1, list templates
	templateDataList, err = listV2TemplateDataApi(params, "")
	if err != nil {
		common.Error(err)
		return false
	}
	if len(templateDataList) == 0 {
		common.Infof("skipping test GET %s/{session_template_id} because result from previous test is []", bosV2SessionTemplatesUri)

		// However, we can still try to list all of the session templates with a tenant name specified.
		// Since no session templates were found from the un-tenanted query, we expect none to be found once a tenant
		// is specified
		tenant := getAnyTenant(tenantList)
		templateDataList, err = listV2TemplateDataApi(params, tenant)
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
		templateDataList, err = listV2TemplateDataApi(params, tenant)
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
	templateDataList, err = listV2TemplateDataApi(params, tenantedTemplateData.Tenant)
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

// v1 session templates CLI tests
// See comment earlier in the file for a description of this function
func v1SessionTemplatesTestsCLICommand(cmdArgs ...string) bool {
	var untenantedTemplateData v2TemplateData

	// test #1, list session templates
	templateDataList, passed := listTemplateDataCli("", cmdArgs...)
	if !passed {
		return false
	}
	if len(templateDataList) == 0 {
		common.Infof("skipping test CLI %s describe {template_id} because result from previous test is []", strings.Join(cmdArgs, " "))
		return true
	}

	// From the template ID list, we want to identify:
	// * One template that has no tenant (untenantedTemplateData)
	for templateIndex, templateData := range templateDataList {
		common.Debugf("Parsing template #%d in the template list (%s)", templateIndex, templateData.String())
		if len(templateData.Tenant) == 0 {
			// This template has no tenant
			untenantedTemplateData = templateData
			common.Infof("Found BOS template #%d (%s)", templateIndex, untenantedTemplateData.String())
			break
		}
	}

	if untenantedTemplateData.IsNil() {
		common.Infof("skipping test CLI %s describe {template_id} with no tenant specified, because all BOS templates are owned by tenants",
			strings.Join(cmdArgs, " "))
		return true
	}

	// an untenanted session template id is available
	// test #2 describe session templates
	cmdOut := runBosCLIDescribe(untenantedTemplateData.Name, cmdArgs...)
	if cmdOut == nil {
		return false
	}
	return ValidateTemplateData(cmdOut, untenantedTemplateData)
}

// v2 session templates CLI tests
// See comment earlier in the file for a description of this function
func v2SessionTemplatesTestsCLICommand(tenantList []string, cmdArgs ...string) bool {
	var tenantedTemplateData, untenantedTemplateData v2TemplateData
	var tenantedTemplateCount int
	var ok, passed bool
	var templateDataList []v2TemplateData

	// test #1, list templates with no tenant specified
	templateDataList, ok = listTemplateDataCli("", cmdArgs...)
	if !ok {
		return false
	}
	if len(templateDataList) == 0 {
		common.Infof("skipping test CLI %s describe {template_id} because result from previous test is []", strings.Join(cmdArgs, " "))

		// However, we can still try to list all of the session templates with a tenant name specified.
		// Since no session templates were found from the un-tenanted query, we expect none to be found once a tenant
		// is specified
		tenant := getAnyTenant(tenantList)
		templateDataList, ok = listTemplateDataCli(tenant, cmdArgs...)
		if !ok {
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

	passed = true

	if untenantedTemplateData.IsNil() {
		common.Infof("skipping test CLI %s describe {template_id} with no tenant specified, because all BOS v2 session templates are owned by tenants",
			strings.Join(cmdArgs, " "))
	} else {
		// test: describe session template using the untenanted session template name we found earlier
		_, ok = describeV2TemplateDataCli(untenantedTemplateData, cmdArgs...)
		passed = passed && ok
	}

	if tenantedTemplateData.IsNil() {
		common.Infof("No BOS v2 session templates found belonging to any tenants")

		tenant := getAnyTenant(tenantList)
		templateDataList, ok = listTemplateDataCli(tenant, cmdArgs...)
		if !ok {
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
	templateDataList, ok = listTemplateDataCli(tenantedTemplateData.Tenant, cmdArgs...)
	if !ok {
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
	_, ok = describeV2TemplateDataCli(tenantedTemplateData, cmdArgs...)
	passed = passed && ok

	return passed
}
