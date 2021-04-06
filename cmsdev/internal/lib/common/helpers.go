/*
Copyright 2021 Hewlett Packard Enterprise Development LP
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
