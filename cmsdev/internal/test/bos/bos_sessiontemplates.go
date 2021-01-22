package bos 

/*
 * bos_sessiontemplate.go
 * 
 * bos sessiontemplate tests  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */

import (
	"encoding/json"
	"net/http"
	"reflect"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
   	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
)

// sessiontemplate tests
func sessionTemplateTests(local bool) {
	var baseurl string = common.BASEURL
	const totalNumTests int = 2

	numTests, numTestsFailed := 0, 0
	params := test.GetAccessTokenParams()
	if params == nil { return }

	// test #1, list session templates
	url := baseurl + endpoints["bos"]["sessiontemplate"].Url
	numTests++
	test.RestfulTestHeader("GET sessiontemplate", numTests, totalNumTests)
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		numTestsFailed++
	}

	// TODO: deeper validation of returned response

	// test #2, list sessiontemplate with session_template_id 
	// use results from previous tests, grab the first sessiontemplate
	var m interface{}
	var skip bool = false
	var sessionTemplateId string

	if err := json.Unmarshal(resp.Body(), &m); err != nil {
		common.VerbosePrintDivider()
		common.Infof("skipping test GET /sessiontemplate/{session_template_id}")
		common.Infof("resultes from previous test is []")
		skip = true
	}

	// a session_template_id is available
	if skip == false {
		p := m.([]interface{})
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
	}

	// TODO: deeper validation of returned response

	test.RestfulTestResultSummary(numTestsFailed, numTests)
}
