/*
Copyright 2021 Hewlett Packard Enterprise Development LP
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
