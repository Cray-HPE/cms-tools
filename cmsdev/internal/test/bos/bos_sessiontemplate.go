package bos

/*
 * bos_sessiontemplate.go
 *
 * bos sessiontemplate tests
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

func getFirstSessionTemplateId(listCmdOut []byte) (sessionTemplateId string, err error) {
	var m interface{}
	var idFound bool
	sessionTemplateId = ""

	err = json.Unmarshal(listCmdOut, &m)
	if err != nil {
		return
	}
	p, ok := m.([]interface{})
	if !ok {
		err = fmt.Errorf("JSON response object is not a list")
		return
	} else if len(p) == 0 {
		// List is empty -- no session template ID to find
		return
	}

	// The list is not empty -- get name field from the first entry
	idFound = false
	for k, v := range p[0].(map[string]interface{}) {
		if reflect.TypeOf(k).String() == "string" && k == "name" {
			sessionTemplateId = v.(string)
			idFound = true
			break
		}
	}
	if !idFound {
		err = fmt.Errorf("No name field found in first list item")
	} else if len(sessionTemplateId) == 0 {
		err = fmt.Errorf("name field of first list item is blank")
	}
	return
}

// sessiontemplate tests
func sessionTemplateTestsAPI() bool {
	var baseurl string = common.BASEURL
	const totalNumTests int = 2

	numTests, numTestsFailed := 0, 0
	params := test.GetAccessTokenParams()
	if params == nil {
		return false
	}

	// test #1, list session templates
	url := baseurl + endpoints["bos"]["sessiontemplate"].Url
	numTests++
	test.RestfulTestHeader("GET sessiontemplate", numTests, totalNumTests)
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	// TODO: deeper validation of returned response

	// test #2, list sessiontemplate with session_template_id
	// use results from previous tests, grab the first sessiontemplate
	sessionTemplateId, err := getFirstSessionTemplateId(resp.Body())
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionTemplateId) == 0 {
		common.VerbosePrintDivider()
		common.Infof("skipping test GET /sessiontemplate/{session_template_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// a session_template_id is available
	url = baseurl + endpoints["bos"]["sessiontemplate"].Url + "/" + sessionTemplateId
	numTests++
	test.RestfulTestHeader("GET session_template_id", numTests, totalNumTests)
	_, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	}
	// TODO: deeper validation of returned response

	return numTestsFailed == 0
}

func sessionTemplateTestsCLI(vnum int) bool {
	var cmdString, verString string

	if vnum == 0 {
		verString = "bos"
	} else if vnum > 0 {
		verString = fmt.Sprintf("bos v%d", vnum)
	} else {
		common.Errorf("PROGRAMMING LOGIC ERROR: sessionTestCLI: Negative vnum value (%d)", vnum)
	}

	// test #1, list session templates
	common.Infof("Getting list of all BOS session templates via CLI")
	cmdString = fmt.Sprintf("%s sessiontemplate list --format json", verString)
	cmdOut := test.RunCLICommand(cmdString)
	if cmdOut == nil {
		return false
	}

	// TODO: deeper validation of returned response

	// test #2, list sessiontemplate with session_template_id
	// use results from previous tests, grab the first sessiontemplate
	sessionTemplateId, err := getFirstSessionTemplateId(cmdOut)
	if err != nil {
		common.Error(err)
		return false
	} else if len(sessionTemplateId) == 0 {
		common.VerbosePrintDivider()
		common.Infof("skipping test CLI describe sessiontemplate {session_template_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// a session_template_id is available
	common.Infof("Describing BOS session template %s via CLI", sessionTemplateId)
	cmdString = fmt.Sprintf("%s sessiontemplate describe %s --format json", verString, sessionTemplateId)
	cmdOut = test.RunCLICommand(cmdString)
	if cmdOut == nil {
		return false
	}

	// TODO: deeper validation of returned response
	return true
}
