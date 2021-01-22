package ims

/*
 * ims.go
 * 
 * ims commons file  
 *
 * Copyright 2019-2020, Cray Inc.
 */

import (
	"fmt"
	"regexp"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"gopkg.in/yaml.v2"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
	coreV1 "k8s.io/api/core/v1"
)

const DEFAULT_INIT_IMS_YAML string = "/opt/cray/crayctl/ansible_framework/roles/cray_init_ims/defaults/main.yml"

type RecipesYAML struct {
	Cray_init_ims_initial_recipes map[string]map[string]string
}

type Recipe struct {
	Name, Distro string
}

type IMSImageRecord struct {
	Created, Id, Name string
	Link map[string]string
}

type IMSConnectionInfoRecord struct {
	Host string
	Port int
}

type IMSSSHContainerRecord struct {
	Connection_info map[string]IMSConnectionInfoRecord
	Jail bool
	Name, Status string
}

type IMSJobRecord struct {
	Artifact_id, Created, Id, Image_root_archive_name, Initrd_file_name,
		Job_type, Kernel_file_name, Kubernetes_configmap, Kubernetes_job,
		Kubernetes_namespace, Kubernetes_service, Public_key_id,
		Resultant_image_id, Status string
	Build_env_size int
	Enable_debug bool
	Ssh_containers []IMSSSHContainerRecord
}

type IMSPublicKeyRecord struct {
	Created, Id, Name, Public_key string
}

type IMSRecipeRecord struct {
	Id, Created, Recipe_type, Linux_distribution, Name string
	Link map[string]string
}

type IMSVersionRecord struct {
	Version string
}

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()

func IsIMSRunning(local, smoke, ct bool, crayctlStage string) bool {
	var failed, found bool
	switch crayctlStage {
	case "1", "2", "3":
		common.Infof("Nothing to run for this stage")
		return true
	case "4", "5":
		// check service pod status
		if !test.CheckServicePodStatsByPrefixKey("ims", 1, 1) {
			return false
		}

		// check ims pvc pod status
		if !test.CheckPVCStatusByPrefixKey("imsPvc") {
			return false
		}

		failed = false
		// Get list of all expected default recipes
		recipesFromYAML, err := getDefaultImsRecipes()
		if err != nil {
			common.Error(err)
			failed = true
		}

		// Verify any default ims recipe pods are Succeeded
		common.Infof("Getting list of cray-init-recipe pods")
		pods, err := k8s.GetPods(common.NAMESPACE, "cray-init-recipe")
		if err != nil {
			common.Error(err)
			failed = true
		} else if len(pods) > 0 {
			for _, pod := range pods {
				if len(pod.Status.Message) > 0 {
					common.Infof("Found pod %s with phase %s (message: %s)", pod.GetName(), pod.Status.Phase, pod.Status.Message)
				} else {
					common.Infof("Found pod %s with phase %s", pod.GetName(), pod.Status.Phase)
				}
				if pod.Status.Phase == "Succeeded" {
					continue
				}
				// Check if this is one of our default pods
				matchingRecipesFromYAML := findMatchingRecipes(pod.GetName(), recipesFromYAML)
				if len(matchingRecipesFromYAML) == 0 {
					common.Warnf("This pod has not Succeeded, but it is not for one of our default recipes")
					continue
				}
				recFromPod, okay := getRecipeEnvVars(pod)
				if !okay {
					failed = true
					continue
				}
				found = false
				for _, matchingRec := range matchingRecipesFromYAML {
					if recFromPod.Name == matchingRec.Name && recFromPod.Distro == matchingRec.Distro {
						found = true
					} else if recFromPod.Name == matchingRec.Name {
						common.Warnf("The recipe name matches one of our default recipes, but with a different distro")
					}
				}
				if found {
					common.Errorf("Default recipe pod %s should have phase Succeeded", pod.GetName())
					failed = true
				} else {
					common.Warnf("This pod is not for one of our default recipes, so we do not fail even though it has not Succeeded")
				}
			}
		} else {
			// We don't consider this a failure because as long as the recipes exist in
			// IMS, it is okay if the pod isn't there
			common.Warnf("No cray-init-recipe pods found")
		}

		// Verify that all expected base recipes are available in IMS
		// Get all recipe records from IMS
		imsRecipeList := getIMSRecipeRecordsAPI()
		if imsRecipeList == nil { return false }
		common.Infof("Found %d recipe records in IMS", len(imsRecipeList))

		common.Infof("Verifying that all expected IMS recipes are available")
		for _, recipeFromYAML := range recipesFromYAML {
			found = false
			common.Infof("Checking for recipe %s, distro %s", recipeFromYAML.Name, recipeFromYAML.Distro)
			for _, irec := range imsRecipeList {
				if irec.Name == recipeFromYAML.Name {
					if irec.Linux_distribution == recipeFromYAML.Distro {
						common.Infof("Found this recipe in IMS with id %s", irec.Id)
						if found {
							common.Warnf("Multiple IMS recipe records found matching this name and distro")
						}
						found = true
					} else {
						common.Warnf("Found an IMS record (id %s) for this recipe name, but different distro (%s)", irec.Id, irec.Linux_distribution)
					}
				}
			}
			if !found {
				common.Errorf("No IMS recipe record found with name %s and distro %s", recipeFromYAML.Name, recipeFromYAML.Distro)
				failed = true
			}
		}

		// Do a few basic API tests
		checkIMSLivenessProbe()
		checkIMSReadinessProbe()
		ver := getIMSVersion()
		common.Infof("IMS version is reported to be %s", ver)
		getIMSImageRecordsAPI()
		getIMSJobRecordsAPI()
		getIMSPublicKeyRecordsAPI()
		// We don't get the recipes via API as we have done so already earlier in this test

		getIMSImageRecordsCLI()
		getIMSJobRecordsCLI()
		getIMSPublicKeyRecordsCLI()
		getIMSRecipeRecordsCLI()

		if failed { return false }
		return true
	default:
		common.Errorf("Invalid stage for this test")
		return false
	}
	common.Errorf("Programming logic error: this line should never be reached")
	return false
}

// Checks if the specified podName matches one of our default recipes
// e.g. is of the form cray-init-recipe-<recipename>-...
// Returns a list of default recipes which match this pod name
func findMatchingRecipes(podName string, recipesFromYAML []Recipe) (matchingRecipes []Recipe) {
	matchingRecipes = make([]Recipe, 0, len(recipesFromYAML))
	for _, recipe := range recipesFromYAML {
		re, _ := regexp.MatchString("^cray-init-recipe-"+recipe.Name+"-", podName)
		if re {
			matchingRecipes = append(matchingRecipes, recipe)
		}
	}
	return
}

// Checks pod for init-ims container, then checks that container for the
// RECIPE_NAME and RECIPE_LINUX_DISTRIBUTION environment variables.
// Returns a Recipe object with those variables if found. Otherwise returns nil.
func getRecipeEnvVars(pod coreV1.Pod)(recipe Recipe, okay bool) {
	var foundContainer, foundName, foundDistro bool
	var recipeName, recipeDistro string

	okay = true
	foundContainer = false
	foundName = false
	foundDistro = false
	for _, container := range k8s.GetContainers(pod) {
		if container.Name != "init-ims" {
			continue
		} else if foundContainer {
			common.Errorf("Multiple init-ims containers found in pod %s", pod.GetName())
			okay = false
			break
		} else {
			common.Infof("Found init-ims container in %s", pod.GetName())
			foundContainer = true
		}
		for _, evar := range k8s.GetEnvVars(container) {
			if evar.Name == "RECIPE_NAME" {
				if len(evar.Value) == 0 {
					common.Errorf("RECIPE_NAME variable exists but is blank")
					okay = false
				} else {
					common.Infof("Found RECIPE_NAME variable set to %s", evar.Value)
					if foundName {
						common.Errorf("RECIPE_NAME variable is set multiple times")
						okay = false
					} else {
						recipeName = evar.Value
						foundName = true
					}
				}
			} else if evar.Name == "RECIPE_LINUX_DISTRIBUTION" {
				if len(evar.Value) == 0 {
					common.Errorf("RECIPE_LINUX_DISTRIBUTION variable exists but is blank")
					okay = false
				} else {
					common.Infof("Found RECIPE_LINUX_DISTRIBUTION variable set to %s", evar.Value)
					if foundDistro {
						common.Errorf("RECIPE_LINUX_DISTRIBUTION variable is set multiple times")
						okay = false
					} else {
						recipeDistro = evar.Value
						foundDistro = true
					}
				}
			}
		}
		if !foundName {
			common.Errorf("RECIPE_NAME variable not found")
			okay = false
		}
		if !foundDistro {
			common.Errorf("RECIPE_LINUX_DISTRIBUTION variable not found")
			okay = false
		}
		if okay {
			recipe.Name = recipeName
			recipe.Distro= recipeDistro
		}
	}
	if !foundContainer {
		common.Errorf("No init-ims container found in pod %s", pod.GetName())
		okay = false
	}
	return
}

// Returns information on recipes defined in DEFAULT_INIT_IMS_YAML
func getDefaultImsRecipes() (recipes []Recipe, err error) {
	var recipesYaml RecipesYAML

	common.Infof("Reading %s", DEFAULT_INIT_IMS_YAML)
	source, err := ioutil.ReadFile(DEFAULT_INIT_IMS_YAML)
	if err != nil { return }

	common.Infof("Parsing file to extract cray_init_ims_initial_recipes map")
	err = yaml.Unmarshal(source, &recipesYaml)
	if err != nil { return }

	numRecipes := len(recipesYaml.Cray_init_ims_initial_recipes)
	if numRecipes <= 0 {
		err = fmt.Errorf("No recipes found in file")
		return
	}
	
	recipes = make([]Recipe, 0, numRecipes)
	for name := range recipesYaml.Cray_init_ims_initial_recipes {
		common.Infof("Found recipe in file: %s", name)
		if distro, ok := recipesYaml.Cray_init_ims_initial_recipes[name]["distribution"]; ok {
			common.Infof("Distro is %s", distro)
			recipes = append(recipes, Recipe{ Name: name, Distro: distro })
		} else {
			if err == nil {
				err = fmt.Errorf("Recipe %s does not have a distribution specified in %s", name, DEFAULT_INIT_IMS_YAML)
			}
			recipes = append(recipes, Recipe{ Name: name })
		}
	}
	return
}

// Return a list of all image records in IMS
func getIMSImageRecordsAPI() []IMSImageRecord {
	var baseurl string = common.BASEURL

	common.Infof("Getting list of all image records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil { return nil }
	url := baseurl + endpoints["ims"]["images"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return nil
	}
	
	// Extract list of image records from response
	var recordList []IMSImageRecord
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return nil
	}

	return recordList
}

// Return a list of all image records in IMS
func getIMSImageRecordsCLI() []IMSImageRecord {
	common.Infof("Getting list of all image records in IMS via CLI")
	cmdOut := test.RunCLICommand("cray ims images list --format json")
	if cmdOut == nil {
		return nil
	}
	
	// Extract list of image records from command output
	var recordList []IMSImageRecord
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recordList); err != nil {
		common.Error(err)
		return nil
	}
	return recordList
}

// Return a list of all job records in IMS
func getIMSJobRecordsAPI() []IMSJobRecord {
	var baseurl string = common.BASEURL

	common.Infof("Getting list of all job records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil { return nil }
	url := baseurl + endpoints["ims"]["jobs"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return nil
	}
	
	// Extract list of job records from response
	var recordList []IMSJobRecord
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return nil
	}

	return recordList
}

// Return a list of all job records in IMS
func getIMSJobRecordsCLI() []IMSJobRecord {
	common.Infof("Getting list of all job records in IMS via CLI")
	cmdOut := test.RunCLICommand("cray ims jobs list --format json")
	if cmdOut == nil {
		return nil
	}
	
	// Extract list of job records from command output
	var recordList []IMSJobRecord
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recordList); err != nil {
		common.Error(err)
		return nil
	}

	return recordList
}

// Return a list of all public key records in IMS
func getIMSPublicKeyRecordsAPI() []IMSPublicKeyRecord {
	var baseurl string = common.BASEURL

	common.Infof("Getting list of all public key records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil { return nil }
	url := baseurl + endpoints["ims"]["public_keys"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return nil
	}
	
	// Extract list of public key records from response
	var recordList []IMSPublicKeyRecord
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return nil
	}

	return recordList
}

// Return a list of all public key records in IMS
func getIMSPublicKeyRecordsCLI() []IMSPublicKeyRecord {
	common.Infof("Getting list of all public key records in IMS via CLI")
	cmdOut := test.RunCLICommand("cray ims public-keys list --format json")
	if cmdOut == nil {
		return nil
	}
	
	// Extract list of public key records from command output
	var recordList []IMSPublicKeyRecord
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recordList); err != nil {
		common.Error(err)
		return nil
	}

	return recordList
}

// Return a list of all recipe records in IMS
func getIMSRecipeRecordsAPI() []IMSRecipeRecord {
	var baseurl string = common.BASEURL

	common.Infof("Getting list of all recipe records in IMS via API")
	params := test.GetAccessTokenParams()
	if params == nil { return nil }
	url := baseurl + endpoints["ims"]["recipes"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return nil
	}
	
	// Extract list of recipe records from response
	var recordList []IMSRecipeRecord
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &recordList); err != nil {
		common.Error(err)
		return nil
	}

	return recordList
}

// Return a list of all recipe records in IMS
func getIMSRecipeRecordsCLI() []IMSRecipeRecord {
	common.Infof("Getting list of all recipe records in IMS via CLI")
	cmdOut := test.RunCLICommand("cray ims recipes list --format json")
	if cmdOut == nil {
		return nil
	}

	// Extract list of recipe records from command output
	var recordList []IMSRecipeRecord
	common.Infof("Decoding JSON in command output")
	if err := json.Unmarshal(cmdOut, &recordList); err != nil {
		common.Error(err)
		return nil
	}

	return recordList
}

// Check IMS liveness probe. Returns True if live, False otherwise
func checkIMSLivenessProbe() bool {
	var baseurl string = common.BASEURL

	common.Infof("Checking IMS Liveness Probe")
	params := test.GetAccessTokenParams()
	if params == nil { return false }
	url := baseurl + endpoints["ims"]["live"].Url
	_, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Check IMS readiness probe. Returns True if ready, False otherwise
func checkIMSReadinessProbe() bool {
	var baseurl string = common.BASEURL

	common.Infof("Checking IMS Readiness Probe")
	params := test.GetAccessTokenParams()
	if params == nil { return false }
	url := baseurl + endpoints["ims"]["ready"].Url
	_, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return false
	}
	return true
}

// Return IMS version. Returns an empty string if error.
func getIMSVersion() string {
	var baseurl string = common.BASEURL

	common.Infof("Getting IMS version")
	params := test.GetAccessTokenParams()
	if params == nil { return "" }
	url := baseurl + endpoints["ims"]["version"].Url
	resp, err := test.RestfulVerifyStatus("GET", url, *params, http.StatusOK)
	if err != nil {
		common.Error(err)
		return ""
	}
	
	// Extract version record from response
	var record IMSVersionRecord
	common.Infof("Decoding JSON in response body")
	if err := json.Unmarshal(resp.Body(), &record); err != nil {
		common.Error(err)
		return ""
	}

	return record.Version
}
