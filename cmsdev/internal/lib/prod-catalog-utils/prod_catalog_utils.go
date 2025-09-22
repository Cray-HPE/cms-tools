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
 * prod_catalog_utils.go
 *
 * CSM Product Catalog Utils
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
	Initialized   bool              `json:"-"` // internal use only to indicate if this struct has been initialized
}

var prodCatError error

// var LatestProdCatEntry map[interface{}]interface{}
var LatestProdCatEntry ProdCatalogEntry

// Used to indicate if dummy data is being used for testing
var prodCatalogDummyData = false

func SetDummyDataFlag(value bool) {
	prodCatalogDummyData = value
}

func IsUsingDummyData() bool {
	return prodCatalogDummyData
}

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

// Initialize the ProdCatalogEntry with dummy data
func UseProdCatalogEntryDummyData() error {
	common.Error(fmt.Errorf("%v. Using dummy data for testing.", prodCatError))
	dummyData := map[interface{}]interface{}{
		"configuration": map[interface{}]interface{}{
			"clone_url":     "https://vcs.cmn.wasp.hpc.amslabs.hpecorp.net/vcs/cray/dummy-csm-config-management.git",
			"commit":        "f5e2ffc9560c19858c3dd4423708a70da80d4999",
			"import_branch": "cray/csm/1.48.0",
			"import_date":   "2025-08-19 10:56:34.133195",
			"ssh_url":       "",
		},
		"images": map[interface{}]interface{}{
			"compute-csm-1.7-7.1.37-aarch64": map[interface{}]interface{}{
				"id": "8ceeac2a-221e-470d-ae49-18e648b96999",
			},
			"compute-csm-1.7-7.1.37-x86_64": map[interface{}]interface{}{
				"id": "870d6bce-a626-48a2-98d3-4fe628e90999",
			},
		},
		"recipes": map[interface{}]interface{}{
			"cray-shasta-csm-sles15sp6-barebones-csm-1.7-aarch64": map[interface{}]interface{}{
				"id": "5e1d9136-49cd-4551-b359-7edc1f270999",
			},
			"cray-shasta-csm-sles15sp6-barebones-csm-1.7-x86_64": map[interface{}]interface{}{
				"id": "df002ba2-5ffc-40f0-b39b-d4e1d49e0999",
			},
		},
	}
	entry, err := MapToProdCatalogEntry(dummyData)
	if err != nil {
		return fmt.Errorf("Failed to convert dummy data: %v", err)
	} else if !entry.Initialized {
		return fmt.Errorf("Dummy data not properly initialized")
	}

	SetDummyDataFlag(true)
	LatestProdCatEntry = entry
	return nil
}

// Helper function to convert map[interface{}]interface{} to ProdCatalogEntry
// Example input data:
//
//	configuration:
//	  clone_url: https://vcs.cmn.wasp.hpc.amslabs.hpecorp.net/vcs/cray/csm-config-management.git
//	  commit: f5e2ffc9560c19858c3dd4423708a70da80d4006
//	  import_branch: cray/csm/1.48.0
//	  import_date: 2025-08-19 10:56:34.133195
//	  ssh_url: git@vcs.cmn.wasp.hpc.amslabs.hpecorp.net:cray/csm-config-management.git
//	images:
//	  compute-csm-1.7-7.1.37-aarch64:
//	    id: 8ceeac2a-221e-470d-ae49-18e648b96371
//	  compute-csm-1.7-7.1.37-x86_64:
//	    id: 870d6bce-a626-48a2-98d3-4fe628e90585
//	recipes:
//	  cray-shasta-csm-sles15sp6-barebones-csm-1.7-aarch64:
//	    id: 5e1d9136-49cd-4551-b359-7edc1f27061e
//	  cray-shasta-csm-sles15sp6-barebones-csm-1.7-x86_64:
//	    id: df002ba2-5ffc-40f0-b39b-d4e1d49e090e
//
// Example mapped ProdCatalogEntry:
//
//	ProdCatalogEntry{
//	  Configuration: Configuration{
//	    CloneURL:     "https://vcs.cmn.wasp.hpc.amslabs.hpecorp.net/vcs/cray/csm-config-management.git",
//	    Commit:       "f5e2ffc9560c19858c3dd4423708a70da80d4006",
//	    ImportBranch: "cray/csm/1.48.0",
//	    ImportDate:   "2025-08-19 10:56:34.133195",
//	    SSHURL:       "git@vcs.cmn.wasp.hpc.amslabs.hpecorp.net:cray/csm-config-management.git",
//	  },
//	  Images: map[string]Image{
//	    "compute-csm-1.7-7.1.37-aarch64": {ID: "8ceeac2a-221e-470d-ae49-18e648b96371"},
//	    "compute-csm-1.7-7.1.37-x86_64":  {ID: "870d6bce-a626-48a2-98d3-4fe628e90585"},
//	  },
//	  Recipes: map[string]Recipe{
//	    "cray-shasta-csm-sles15sp6-barebones-csm-1.7-aarch64": {ID: "5e1d9136-49cd-4551-b359-7edc1f27061e"},
//	    "cray-shasta-csm-sles15sp6-barebones-csm-1.7-x86_64":  {ID: "df002ba2-5ffc-40f0-b39b-d4e1d49e090e"},
//	  },
//	}
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
	entry.Initialized = true
	return entry, nil
}

// GetCSMLatestVersion returns the latest valid CSM version string from the product catalog config map data.
// For testing purposes, both configuration and images must be present.
// It sorts all available version keys, checks for the presence of both "configuration" and "images" fields,
// and selects the highest version that meets these criteria. Returns the version string and an error if not found.
func GetCSMLatestVersion(prodCatlogData map[string]interface{}) (version string, err error) {
	// Sort the keys in the JSON object to ensure consistent ordering
	keys := make([]string, 0, len(prodCatlogData))
	for key := range prodCatlogData {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return "", fmt.Errorf("No keys found in the configuration map")
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

// GetLatestCSMProductCatalogEntry fetches the latest CSM version entry from the product catalog config map
// and caches it in LatestProdCatEntry. It retrieves the config map data from Kubernetes, determines the
// latest valid CSM version, converts its data to a ProdCatalogEntry, and stores it for future use.
// Returns an error if fetching or conversion fails.
func GetLatestCSMProductCatalogEntry() error {
	prodCatalogData, err := GetCsmProductCatalogData()
	if err != nil {
		return fmt.Errorf("Failed to get product catalog data: %v", err)
	}

	latestVersion, err := GetCSMLatestVersion(prodCatalogData)
	if err != nil {
		return fmt.Errorf("Failed to get latest CSM version: %v", err)
	}

	versionData, ok := prodCatalogData[latestVersion].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("Latest CSM version data not found or invalid format")
	}
	entry, err := MapToProdCatalogEntry(versionData)
	if err != nil {
		return fmt.Errorf("Failed to convert version data: %v", err)
	} else if !entry.Initialized {
		return fmt.Errorf("Programming logic error: No error raised parsing product catalog data, but it is not properly initialized")
	}
	LatestProdCatEntry = entry
	return nil
}

// GetLatestProdCatEntry returns the ProdCatalogEntry for the latest CSM version.
// If a cached entry is available, it is returned immediately.
// If a previous error occurred, that error is returned.
// Otherwise, the function fetches and caches the latest entry from the product catalog config map.
// Returns the latest ProdCatalogEntry and an error if fetching fails.
// Caches the error in prodCatError to avoid repeated fetch attempts in future calls.
func GetLatestProdCatEntry() (ProdCatalogEntry, error) {
	if LatestProdCatEntry.Initialized {
		common.Infof("Using cached product catalog data")
		return LatestProdCatEntry, nil
	}
	if prodCatError != nil {
		return ProdCatalogEntry{}, fmt.Errorf("Product catalog data unavailable due to previous failure: %v", prodCatError)
	}
	if err := GetLatestCSMProductCatalogEntry(); err != nil {
		prodCatError = err
		// On error, use dummy data for testing
		errDummyData := UseProdCatalogEntryDummyData()
		if errDummyData != nil {
			// If dummy data creation fails, return a concatenated error, Also assign concatenated error to prodCatError
			prodCatError = fmt.Errorf("%v; also failed to use dummy product catalog data: %v", prodCatError, errDummyData)
			return ProdCatalogEntry{}, prodCatError
		}
	}
	return LatestProdCatEntry, prodCatError
}
