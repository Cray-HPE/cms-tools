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
 * bos_sessiontemplate_api_tests.go
 *
 * BOS sessiontemplate API tests
 *
 */
package bos

import (
	"fmt"
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

func TestSessionTemplatesCRUDOperations() (passed bool) {
	// Create session template with X86 architecture image
	sessionTemplateRecordX86, ok := TestSessionTemplatesCreate("X86")
	if !ok {
		common.Errorf("Session template creation failed")
		return false
	}

	// Create session template with ARM architecture image
	sessionTemplateRecordARM, success := TestSessionTemplatesCreate("ARM")
	if !success {
		common.Errorf("Session template creation failed with ARM architecture")
	}

	// Update session template
	updated := TestSessionTemplatesUpdate(sessionTemplateRecordX86.Name)

	//delete session template
	deleted := TestSessionTemplatesDelete(sessionTemplateRecordX86.Name)

	// Delete session template with ARM architecture image
	if success {
		deleted = deleted && TestSessionTemplatesDelete(sessionTemplateRecordARM.Name)
	}

	// Get all session templates
	getAll := TestSessionTemplatesGetAll()

	return updated && deleted && getAll
}

func TestSessionTemplatesCreate(imageArch string) (sessionTemplateRecord BOSSessionTemplate, passed bool) {
	templateName := "BOS_SessionTemplate_" + string(common.GetRandomString(10))
	common.PrintLog(fmt.Sprintf("Creating BOS session template %s with arch %s", templateName, imageArch))

	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	// create sessiontemplates payload
	payload, success := GetCreateBOSSessionTemplatePayload(cfgName, false, imageArch)
	if !success {
		return BOSSessionTemplate{}, false
	}
	common.Debugf("BOS Session template create payload: %s", payload)
	// Create session template
	sessionTemplateRecord, success = CreateUpdateBOSSessiontemplatesAPI(payload, templateName, "PUT")
	if !success {
		return BOSSessionTemplate{}, false
	}

	// Verify sessiontemplate
	if !VerifyBOSSessionTemplate(sessionTemplateRecord, payload, templateName) {
		common.Errorf("Session template %s verification failed", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// Get the created session template
	_, success = GetBOSSessionTemplatesAPI(sessionTemplateRecord.Name, http.StatusOK)
	if !success {
		common.Errorf("Unable to get BOS session template %s", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// verify session template in list of session templates
	sessionTemplateRecords, success := GetAllBOSSessionTemplatesAPI()
	if !success {
		common.Errorf("Unable to get all session templates")
		return BOSSessionTemplate{}, false
	}

	if !BOSSessionTemplateExists(sessionTemplateRecord.Name, sessionTemplateRecords) {
		common.Errorf("BOS session template %s not found in list of session templates", sessionTemplateRecord.Name)
		return BOSSessionTemplate{}, false
	}

	// Validate session template
	if !ValidateBOSSessionTemplateAPI(sessionTemplateRecord.Name) {
		return BOSSessionTemplate{}, false
	}

	common.Infof("Session template %s created successfully", sessionTemplateRecord.Name)
	return sessionTemplateRecord, true

}

func TestSessionTemplatesUpdate(templateName string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Updating session template %s", templateName))
	cfgName := "CFS_Configuration_" + string(common.GetRandomString(10))

	sessionTemplate, success := GetBOSSessionTemplatesAPI(templateName, http.StatusOK)
	if !success {
		common.Errorf("Unable to get session template %s", templateName)
		return false
	}

	payload, success := GetCreateBOSSessionTemplatePayload(cfgName, true, sessionTemplate.Boot_sets.Compute.Arch)
	if !success {
		return false
	}
	common.Debugf("Session template update payload: %s", payload)
	sessionTemplateRecord, success := CreateUpdateBOSSessiontemplatesAPI(payload, templateName, "PATCH")
	if !success {
		common.Errorf("Session template %s update failed", templateName)
		return false
	}

	// Verify sessiontemplate
	if !VerifyBOSSessionTemplate(sessionTemplateRecord, payload, templateName) {
		common.Errorf("Session template %s verification failed", sessionTemplateRecord.Name)
		return false
	}

	// Get the created session template
	_, success = GetBOSSessionTemplatesAPI(sessionTemplateRecord.Name, http.StatusOK)
	if !success {
		common.Errorf("Unable to get BOS session template %s", sessionTemplateRecord.Name)
		return false
	}

	// Validate session template
	if !ValidateBOSSessionTemplateAPI(sessionTemplateRecord.Name) {
		return false
	}

	common.Infof("Session template %s updated successfully", templateName)
	return true

}

func TestSessionTemplatesDelete(templateName string) (passed bool) {
	common.PrintLog(fmt.Sprintf("Deleting session template %s", templateName))
	// Delete session template
	if !DeleteBOSSessionTemplatesAPI(templateName) {
		common.Errorf("Unable to delete sessiontemplate %s", templateName)
		return false
	}

	// Get the deleted session template
	_, success := GetBOSSessionTemplatesAPI(templateName, http.StatusNotFound)
	if !success {
		common.Errorf("BOS sessiontemplate %s was not deleted", templateName)
		return false
	}

	// verify session template in list of session templates
	sessionTemplateRecords, success := GetAllBOSSessionTemplatesAPI()
	if !success {
		common.Errorf("Unable to get all session templates")
		return false
	}

	if BOSSessionTemplateExists(templateName, sessionTemplateRecords) {
		common.Errorf("BOS session template %s not deleted, found in list of session templates", templateName)
		return false
	}

	common.Infof("Deleted sessiontemplate %s", templateName)
	return true
}

func TestSessionTemplatesGetAll() (passed bool) {
	// Get all session templates
	sessionTemplateRecords, success := GetAllBOSSessionTemplatesAPI()
	if !success {
		common.Errorf("Unable to get all session templates")
		return false
	}

	common.Infof("Found %d BOS sessiontemplates", len(sessionTemplateRecords))
	return true
}
