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
		numTestsFailed++
	} else {
		// TODO: deeper validation of returned response

		// test #2, list sessiontemplate with session_template_id
		// use results from previous tests, grab the first sessiontemplate
		var m interface{}
		var sessionTemplateId string

		if err := json.Unmarshal(resp.Body(), &m); err != nil {
			common.Error(err)
			numTestsFailed++
		} else {
			p, ok := m.([]interface{})
			if !ok {
				common.Errorf("JSON response object is not a list")
				numTestsFailed++
			} else if len(p) == 0 {
				common.VerbosePrintDivider()
				common.Infof("skipping test GET /sessiontemplate/{session_template_id}")
				common.Infof("results from previous test is []")
			} else {
				// a session_template_id is available
				for k, v := range p[0].(map[string]interface{}) {
					if reflect.TypeOf(k).String() == "string" && k == "name" {
						sessionTemplateId = v.(string)
						break
					}
				}
				url = baseurl + endpoints["bos"]["sessiontemplate"].Url + "/" + sessionTemplateId
				numTests++
				test.RestfulTestHeader("GET session_template_id", numTests, totalNumTests)
				_, err = test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
				if err != nil {
					common.Error(err)
					numTestsFailed++
				}

				// TODO: deeper validation of returned response
			}
		}
	}

	return numTestsFailed == 0
}

func sessionTemplateTestsCLI() bool {
	// test #1, list session templates
	common.Infof("Getting list of all BOS session templates via CLI")
	cmdOut := test.RunCLICommand("bos v1 sessiontemplate list --format json")
	if cmdOut == nil {
		return false
	}

	// TODO: deeper validation of returned response

	// test #2, list sessiontemplate with session_template_id
	// use results from previous tests, grab the first sessiontemplate
	var m interface{}
	var sessionTemplateId string
	var idFound bool

	if err := json.Unmarshal(cmdOut, &m); err != nil {
		common.Error(err)
		return false
	}

	p, ok := m.([]interface{})
	if !ok {
		common.Errorf("JSON response object is not a list")
		return false
	}
	if len(p) == 0 {
		common.VerbosePrintDivider()
		common.Infof("skipping test CLI describe sessiontemplate {session_template_id}")
		common.Infof("results from previous test is []")
		return true
	}

	// a session_template_id is available
	idFound = false
	for k, v := range p[0].(map[string]interface{}) {
		if reflect.TypeOf(k).String() == "string" && k == "name" {
			sessionTemplateId = v.(string)
			idFound = true
			break
		}
	}
	if !idFound {
		common.Errorf("Unable to find session template name")
		return false
	}

	common.Infof("Describing BOS session template %s via CLI", sessionTemplateId)
	cmdString := fmt.Sprintf("bos v1 sessiontemplate describe %s --format json", sessionTemplateId)
	cmdOut = test.RunCLICommand(cmdString)
	if cmdOut == nil {
		return false
	}

	// TODO: deeper validation of returned response
	return true
}
