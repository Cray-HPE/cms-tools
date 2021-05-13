/*
 * Copyright 2021 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 * 
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 * 
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 * 
 * (MIT License)
 */

package common

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func GetStringFieldFromFirstItem(fieldName string, listJsonBytes []byte) (fieldValue string, err error) {
	var m interface{}
	fieldValue = ""

	Infof("Getting value of \"%s\" field from first element of list in JSON object", fieldName)
	err = json.Unmarshal(listJsonBytes, &m)
	if err != nil {
		return
	}
	listObject, ok := m.([]interface{})
	if !ok {
		err = fmt.Errorf("JSON response object is not a list")
		return
	} else if len(listObject) == 0 {
		// List is empty
		Infof("List is empty")
		return
	}

	firstItem, ok := listObject[0].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("First list item is not a dictionary")
		return
	}

	fieldRawValue, ok := firstItem[fieldName]
	if !ok {
		err = fmt.Errorf("First list item does not have \"%s\" field", fieldName)
		return
	}

	fieldValue, ok = fieldRawValue.(string)
	if !ok {
		err = fmt.Errorf(
			"First list item has \"%s\" field but its value is type %s, not string",
			fieldName, reflect.TypeOf(fieldRawValue).String())
		return
	}

	if len(fieldValue) == 0 {
		err = fmt.Errorf("First list item has empty value for \"%s\" field", fieldName)
		return
	}
	Infof("Value of \"%s\" field in first list item is \"%s\"", fieldName, fieldValue)
	return
}
