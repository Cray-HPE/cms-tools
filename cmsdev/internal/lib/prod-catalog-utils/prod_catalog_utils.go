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
 * bos_sessiontemplates_api.go
 *
 * BOS sessiontemplates API.
 *
 */
package prod_catalog_utils

import (
	"fmt"
	"sort"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
)

type Configuration struct {
	CloneURL     string `json:"clone_url"`
	Commit       string `json:"commit"`
	ImportBranch string `json:"import_branch"`
	ImportDate   string `json:"import_date"`
	SSHURL       string `json:"ssh_url"`
}

type Image struct {
	ID string `json:"id"`
}

type Recipe struct {
	ID string `json:"id"`
}

type ProdCatalogEntry struct {
	Configuration Configuration     `json:"configuration"`
	Images        map[string]Image  `json:"images"`
	Recipes       map[string]Recipe `json:"recipes"`
}

var prodCatError error

// var LatestProdCatEntry map[interface{}]interface{}
var LatestProdCatEntry ProdCatalogEntry

// GetCsmProductCatalogData is used to fetch the CSM product catalog config map data from Kubernetes
// It returns the config map data as a map[string]interface{} and error if any
// The config map data contains the CSM versions and their corresponding configuration and image details
func GetCsmProductCatalogData() (data map[string]interface{}, err error) {
	resp, err := k8s.GetConfigMapDataField(common.NAMESPACE, common.CSMPRODCATALOGCMNAME, "csm")
	if err != nil {
		return nil, err
	}

	//convert YAML
	respMap, err := common.DecodeYAMLIntoStringMap(resp)
	if err != nil {
		return nil, err
	}

	return respMap, nil
}

// Helper function to convert map[interface{}]interface{} to ProdCatalogEntry
func MapToProdCatalogEntry(data map[interface{}]interface{}) (ProdCatalogEntry, error) {
	var entry ProdCatalogEntry

	// Configuration
	if configMap, ok := data["configuration"].(map[interface{}]interface{}); ok {
		entry.Configuration = Configuration{
			CloneURL:     fmt.Sprintf("%v", configMap["clone_url"]),
			Commit:       fmt.Sprintf("%v", configMap["commit"]),
			ImportBranch: fmt.Sprintf("%v", configMap["import_branch"]),
			ImportDate:   fmt.Sprintf("%v", configMap["import_date"]),
			SSHURL:       fmt.Sprintf("%v", configMap["ssh_url"]),
		}
	}

	// Images
	entry.Images = make(map[string]Image)
	if imagesMap, ok := data["images"].(map[interface{}]interface{}); ok {
		for k, v := range imagesMap {
			if imgMap, ok := v.(map[interface{}]interface{}); ok {
				entry.Images[fmt.Sprintf("%v", k)] = Image{
					ID: fmt.Sprintf("%v", imgMap["id"]),
				}
			}
		}
	}

	// Recipes
	entry.Recipes = make(map[string]Recipe)
	if recipesMap, ok := data["recipes"].(map[interface{}]interface{}); ok {
		for k, v := range recipesMap {
			if recMap, ok := v.(map[interface{}]interface{}); ok {
				entry.Recipes[fmt.Sprintf("%v", k)] = Recipe{
					ID: fmt.Sprintf("%v", recMap["id"]),
				}
			}
		}
	}

	return entry, nil
}

// GetCSMLatestVersion returns the latest CSM version from the product catalog config map data
func GetCSMLatestVersion(prodCatlogData map[string]interface{}) (version string, err error) {
	// Sort the keys in the JSON object to ensure consistent ordering
	keys := make([]string, 0, len(prodCatlogData))
	for key := range prodCatlogData {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return "", fmt.Errorf("no keys found in the configuration map")
	}
	//Get the last key from the sorted keys where version has both configuration and images as keys
	latestCSMVersion := ""
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		if _, ok := prodCatlogData[key].(map[interface{}]interface{})["configuration"]; ok {
			if _, ok := prodCatlogData[key].(map[interface{}]interface{})["images"]; ok {
				latestCSMVersion = key
				break
			}
		}
	}

	common.Infof("Latest CSM version: %s", latestCSMVersion)
	return latestCSMVersion, nil
}

// GetAndCacheLatestCSMVersionData fetches and caches the latest CSM version and its data.
// It only fetches and initializes if LatestCSMVersionData is nil or LatestCSMVersion is empty.
func GetLatestCSMProductCatalogEntry() error {
	prodCatalogData, err := GetCsmProductCatalogData()
	if err != nil {
		return fmt.Errorf("failed to get product catalog data: %v", err)
	}

	latestVersion, err := GetCSMLatestVersion(prodCatalogData)
	if err != nil {
		return fmt.Errorf("failed to get latest CSM version: %v", err)
	}

	versionData, ok := prodCatalogData[latestVersion].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("latest CSM version data not found or invalid format")
	}
	entry, err := MapToProdCatalogEntry(versionData)
	if err != nil {
		return fmt.Errorf("failed to convert version data: %v", err)
	}
	LatestProdCatEntry = entry
	return nil
}

// Helper function to check if ProdCatalogEntry is empty
func isProdCatalogEntryEmpty(entry ProdCatalogEntry) bool {
	return entry.Configuration.CloneURL == "" &&
		entry.Configuration.Commit == "" &&
		entry.Configuration.ImportBranch == "" &&
		entry.Configuration.ImportDate == "" &&
		entry.Configuration.SSHURL == "" &&
		len(entry.Images) == 0 &&
		len(entry.Recipes) == 0
}

// GetLatestProdCatEntry returns the ProdCatalogEntry for latest CSM version.
// Returns cached LatestProdCatEntry if available.
// It returns an error if prodCatError is not nil or if fetching fails.
// Fetches and caches the latest entry if not already cached and returns it.
func GetLatestProdCatEntry() (ProdCatalogEntry, error) {
	if !isProdCatalogEntryEmpty(LatestProdCatEntry) {
		common.Infof("Using cached product catalog data")
		return LatestProdCatEntry, nil
	}
	if prodCatError != nil {
		return ProdCatalogEntry{}, fmt.Errorf("Product catalog data unavailable due to previous failure: %v", prodCatError)
	}
	if err := GetLatestCSMProductCatalogEntry(); err != nil {
		prodCatError = err
		return ProdCatalogEntry{}, err
	}
	return LatestProdCatEntry, nil
}
