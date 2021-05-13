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

func DecodeJSONIntoStringMap(mapJsonBytes []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(mapJsonBytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func DecodeJSONIntoList(listJsonBytes []byte) ([]interface{}, error) {
	var m []interface{}
	err := json.Unmarshal(listJsonBytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func DecodeJSONIntoStringList(listJsonBytes []byte) ([]string, error) {
	var m []string
	err := json.Unmarshal(listJsonBytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func GetStringFieldFromFirstItem(fieldName string, listJsonBytes []byte) (fieldValue string, err error) {
	fieldValue = ""

	Infof("Getting value of \"%s\" field from first element of list in JSON object", fieldName)
	listObject, err := DecodeJSONIntoList(listJsonBytes)
	if err != nil {
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

func GetStringFieldFromMap(fieldName string, mapJsonBytes []byte) (fieldValue string, err error) {

	Infof("Getting value of \"%s\" field from JSON object", fieldName)
	mapObject, err := DecodeJSONIntoStringMap(mapJsonBytes)
	if err != nil {
		return
	}
	fieldRawValue, ok := mapObject[fieldName]
	if !ok {
		err = fmt.Errorf("Map does not have \"%s\" field", fieldName)
		return
	}

	fieldValue, ok = fieldRawValue.(string)
	if !ok {
		err = fmt.Errorf(
			"Map has \"%s\" field but its value is type %s, not string",
			fieldName, reflect.TypeOf(fieldRawValue).String())
		return
	}

	Infof("Value of \"%s\" field in first list item is \"%s\"", fieldName, fieldValue)
	return
}

func ValidateStringFieldValue(objectName, fieldName, expectedFieldValue string, mapJsonBytes []byte) (err error) {
	actualValue, err := GetStringFieldFromMap(fieldName, mapJsonBytes)
	if err != nil {
		return
	} else if actualValue != expectedFieldValue {
		err = fmt.Errorf("should have \"%s\" field value of \"%s\", but it is \"%s\"", objectName, fieldName, expectedFieldValue, actualValue)
		return
	}
	Infof("%s has expected value for \"%s\" field", objectName, fieldName)
	return
}
