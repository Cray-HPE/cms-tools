package ims

/*
 * ims.go
 *
 * ims test file
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
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

import (
	coreV1 "k8s.io/api/core/v1"
	"os"
	"regexp"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/cms"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
	"strings"
)

func IsIMSRunning() (passed bool) {
	var ok, found, artifactsCollected bool
	var expectedRecipes []Recipe
	passed = true
	artifactsCollected = false
	// check service pod status
	podNames, ok := test.GetPodNamesByPrefixKey("ims", 1, 1)
	if !ok {
		passed = false
	}
	common.Infof("Found %d ims pods", len(podNames))
	if !test.CheckPodListStats(podNames) {
		passed = false
	}

	for _, pvcName := range pvcNames {
		if !test.CheckPVCStatus(pvcName) {
			passed = false
		}
	}

	if !passed {
		common.ArtifactsPodsPvcs(podNames, pvcNames)
		artifactsCollected = true
	}

	// Until CASMCMS-6027 is resolved, this test does not verify the existence & correctness of any IMS recipes.
	// However, the user can manually include a check for a recipe by using the following environment variables.
	// If set, IMS_RECIPE_NAME specifies the expected name of the the default IMS recipe to verify.
	// If set, IMS_RECIPE_DISTRO specifies the expected distro of the default IMS recipe to verify.
	defaultRecipeName := os.Getenv("IMS_RECIPE_NAME")
	if len(defaultRecipeName) > 0 {
		common.Infof("IMS_RECIPE_NAME set to \"%s\"", defaultRecipeName)

		defaultRecipeDistro := os.Getenv("IMS_RECIPE_DISTRO")
		if len(defaultRecipeDistro) > 0 {
			common.Infof("IMS_RECIPE_DISTRO set to \"%s\"", defaultRecipeDistro)
			expectedRecipes = append(expectedRecipes, Recipe{Name: defaultRecipeName, Distro: defaultRecipeDistro})
		} else {
			// Default to RECIPE_DISTRO_DEFAULT
			common.Infof("IMS_RECIPE_DISTRO not set. Default IMS recipe distro = \"%s\"", RECIPE_DISTRO_DEFAULT)
			expectedRecipes = append(expectedRecipes, Recipe{Name: defaultRecipeName, Distro: RECIPE_DISTRO_DEFAULT})
		}
	} else {
		common.Infof("IMS_RECIPE_NAME not set. Skipping default recipe checks.")
	}

	// Verify any default ims recipe pods are Succeeded
	common.Infof("Getting list of cray-init-recipe pods")
	pods, err := k8s.GetPods(common.NAMESPACE, "cray-init-recipe")
	if err != nil {
		common.Error(err)
		passed = false
	} else if len(pods) > 0 {
		for _, pod := range pods {
			podName := pod.GetName()
			podPhase := pod.Status.Phase
			if len(pod.Status.Message) > 0 {
				common.Infof("Found pod %s with phase %s (message: %s)", podName, podPhase, pod.Status.Message)
			} else {
				common.Infof("Found pod %s with phase %s", podName, podPhase)
			}
			if podPhase == "Succeeded" {
				continue
			}
			// Check if this is one of our default pods
			matchingRecipes := findMatchingRecipes(podName, expectedRecipes)
			if len(matchingRecipes) == 0 {
				common.Warnf("Pod %s has not Succeeded (phase=%s), but it is not for one of our default recipes", podName, podPhase)
				continue
			}
			recFromPod, okay := getRecipeEnvVars(pod)
			if !okay {
				passed = false
				continue
			}
			found = false
			podRecName := recFromPod.Name
			podRecDistro := recFromPod.Distro
			for _, matchingRec := range matchingRecipes {
				if podRecName == matchingRec.Name && podRecDistro == matchingRec.Distro {
					found = true
				} else if podRecName == matchingRec.Name {
					common.Warnf("Recipe name (%s) in this pod (%s) matches a default recipe, but with different distro (pod %s, default %s)", podRecName, podName, podRecDistro, matchingRec.Distro)
				}
			}
			if found {
				common.Errorf("Default recipe pod %s should have phase Succeeded but has phase=%s", podName, podPhase)
				passed = false
			} else {
				common.Warnf("Pod %s is not for a default recipe (pod recipe=%s distro=%s), so we only warn that its phase %s != Succeeded", podName, podRecName, podRecDistro, podPhase)
			}
		}
	} else {
		// We don't consider this a failure because as long as the recipes exist in
		// IMS, it is okay if the pod isn't there
		common.Warnf("No cray-init-recipe pods found")
	}

	if !passed && !artifactsCollected {
		common.ArtifactsPodsPvcs(podNames, pvcNames)
		artifactsCollected = true
	}

	// Verify S3 (from an IMS perspective)
	if !verifyS3(len(expectedRecipes)) {
		passed = false
	}

	// Get all recipe records from IMS
	imsRecipeList, ok := getIMSRecipeRecordsAPI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d recipe records in IMS", len(imsRecipeList))
		if len(expectedRecipes) > 0 {
			// Verify that all expected base recipes are available in IMS
			if !verifyDefaultRecipes(expectedRecipes, imsRecipeList) {
				passed = false
			}

			// Validate that we can retrieve a specific recipe via API
			_, ok = getIMSRecipeRecordAPI(imsRecipeList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	// Do a few basic API and CLI tests
	if !checkIMSLivenessProbe() {
		passed = false
	}
	if !checkIMSReadinessProbe() {
		passed = false
	}
	ver, ok := getIMSVersion()
	if !ok {
		passed = false
	} else {
		common.Infof("IMS version is reported to be %s", ver)
	}

	imsImageList, ok := getIMSImageRecordsAPI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d IMS image records via API", len(imsImageList))
		if len(imsImageList) > 0 {
			_, ok = getIMSImageRecordAPI(imsImageList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	imsJobList, ok := getIMSJobRecordsAPI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d IMS job records via API", len(imsJobList))
		if len(imsJobList) > 0 {
			_, ok = getIMSJobRecordAPI(imsJobList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	imsPkeyList, ok := getIMSPublicKeyRecordsAPI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d IMS public key records via API", len(imsPkeyList))
		if len(imsPkeyList) > 0 {
			_, ok = getIMSPublicKeyRecordAPI(imsPkeyList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	// We don't get the recipes via API as we have done so already earlier in this test

	imsImageList, ok = getIMSImageRecordsCLI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d IMS image records via CLI", len(imsImageList))
		if len(imsImageList) > 0 {
			_, ok = getIMSImageRecordCLI(imsImageList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	imsJobList, ok = getIMSJobRecordsCLI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d IMS job records via CLI", len(imsJobList))
		if len(imsJobList) > 0 {
			_, ok = getIMSJobRecordCLI(imsJobList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	imsPkeyList, ok = getIMSPublicKeyRecordsCLI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d IMS public key records via CLI", len(imsPkeyList))
		if len(imsPkeyList) > 0 {
			_, ok = getIMSPublicKeyRecordCLI(imsPkeyList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	imsRecipeList, ok = getIMSRecipeRecordsCLI()
	if !ok {
		passed = false
	} else {
		common.Infof("Found %d IMS recipe records via CLI", len(imsRecipeList))
		if len(imsRecipeList) > 0 {
			_, ok = getIMSRecipeRecordCLI(imsRecipeList[0].Id)
			if !ok {
				passed = false
			}
		}
	}

	if !passed && !artifactsCollected {
		common.ArtifactsPodsPvcs(podNames, pvcNames)
		artifactsCollected = true
	}

	return
}

func verifyDefaultRecipes(expectedRecipes []Recipe, imsRecipeList []IMSRecipeRecord) (passed bool) {
	passed = true
	common.Infof("Verifying that all expected IMS recipes are available")
	for _, expectedRecipe := range expectedRecipes {
		defaultRecName := expectedRecipe.Name
		defaultRecDistro := expectedRecipe.Distro
		imsid := ""
		common.Infof("Checking for recipe %s, distro %s", defaultRecName, defaultRecDistro)
		for _, irec := range imsRecipeList {
			if irec.Name == defaultRecName {
				if irec.Linux_distribution == defaultRecDistro {
					common.Infof("Found this recipe in IMS with id %s", irec.Id)
					if len(imsid) > 0 {
						common.Warnf("Multiple IMS recipe records found matching recipe name %s and distro %s (IMS ids %s, %s)", defaultRecName, defaultRecDistro, imsid, irec.Id)
					} else {
						imsid = irec.Id
					}
					if irec.Link == nil {
						common.Errorf("Recipe link should not be null, but it is (IMS id %s)", irec.Id)
						passed = false
					}
				} else {
					common.Warnf("Found an IMS record (id %s) for default recipe %s, but different distro (IMS %s, default %s)", irec.Id, defaultRecName, irec.Linux_distribution, defaultRecDistro)
				}
			}
		}
		if len(imsid) == 0 {
			common.Errorf("No IMS recipe record found with name %s and distro %s", defaultRecName, defaultRecDistro)
			passed = false
		}
	}
	return
}

func verifyS3(numExpectedRecipes int) bool {
	bucketList := cms.GetBuckets()
	if bucketList == nil {
		return false
	}

	common.Infof("Found the following S3 buckets: %s", strings.Join(bucketList, " "))
	imsBucketsFound := 0
	for _, bucketName := range bucketList {
		if bucketName == "ims" {
			imsBucketsFound += 1
		}
	}
	if imsBucketsFound == 0 {
		common.Errorf("No S3 bucket named 'ims' found")
		return false
	} else if imsBucketsFound > 1 {
		common.Errorf("%d S3 buckets named 'ims' found, but there should be exactly 1", imsBucketsFound)
		return false
	}

	common.Infof("S3 bucket named 'ims' found")
	artifactList := cms.GetArtifactsInBucket("ims")
	if artifactList == nil {
		return false
	}

	if len(artifactList) < numExpectedRecipes {
		common.InfoOverridef("Found %d IMS S3 artifact(s):", len(artifactList))
		for _, s3Artifact := range artifactList {
			common.InfoOverridef("Key=%s, Etag: %s, Modified: %s", s3Artifact.Key, s3Artifact.ETag, s3Artifact.LastModified)
		}
		common.Errorf("# S3 IMS artifacts (%d) should be >= # of default IMS recipes (%d)", len(artifactList), numExpectedRecipes)
		return false
	}

	common.Infof("Found %d IMS S3 artifact(s)", len(artifactList))
	for _, s3Artifact := range artifactList {
		common.Infof("Key=%s, Etag: %s, Modified: %s", s3Artifact.Key, s3Artifact.ETag, s3Artifact.LastModified)
	}
	return true
}

// Checks if the specified podName matches one of our default recipes
// e.g. is of the form cray-init-recipe-<recipename>-...
// Returns a list of default recipes which match this pod name
func findMatchingRecipes(podName string, expectedRecipes []Recipe) (matchingRecipes []Recipe) {
	matchingRecipes = make([]Recipe, 0, len(expectedRecipes))
	for _, recipe := range expectedRecipes {
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
func getRecipeEnvVars(pod coreV1.Pod) (recipe Recipe, okay bool) {
	var foundContainer bool
	var recipeName, recipeDistro string

	okay = true
	foundContainer = false
	recipeName = ""
	recipeDistro = ""
	podName := pod.GetName()
	for _, container := range k8s.GetContainers(pod) {
		if container.Name != "init-ims" {
			continue
		} else if foundContainer {
			common.Errorf("Multiple init-ims containers found in pod %s", podName)
			okay = false
			break
		} else {
			common.Infof("Found init-ims container in %s", podName)
			foundContainer = true
		}
		for _, evar := range k8s.GetEnvVars(container) {
			if evar.Name == "RECIPE_NAME" {
				if len(evar.Value) == 0 {
					common.Errorf("Pod %s: RECIPE_NAME variable in init-ims container exists but is blank", podName)
					okay = false
				} else {
					common.Infof("Pod %s: Found RECIPE_NAME variable set to %s", podName, evar.Value)
					if len(recipeName) > 0 {
						common.Errorf("Pod %s: RECIPE_NAME variable in init-ims container is set multiple times (%s, %s)", podName, recipeName, evar.Value)
						okay = false
					} else {
						recipeName = evar.Value
					}
				}
			} else if evar.Name == "RECIPE_LINUX_DISTRIBUTION" {
				if len(evar.Value) == 0 {
					common.Errorf("Pod %s: RECIPE_LINUX_DISTRIBUTION variable in init-ims container exists but is blank", podName)
					okay = false
				} else {
					common.Infof("Pod %s: Found RECIPE_LINUX_DISTRIBUTION variable in init-ims container set to %s", podName, evar.Value)
					if len(recipeDistro) > 0 {
						common.Errorf("Pod %s: RECIPE_LINUX_DISTRIBUTION variable in init-ims container is set multiple times (%s, %s)", podName, recipeDistro, evar.Value)
						okay = false
					} else {
						recipeDistro = evar.Value
					}
				}
			}
		}
		if len(recipeName) == 0 {
			common.Errorf("Pod %s: RECIPE_NAME variable not found in init-ims container", podName)
			okay = false
		}
		if len(recipeDistro) == 0 {
			common.Errorf("Pod %s: RECIPE_LINUX_DISTRIBUTION variable not found in init-ims container", podName)
			okay = false
		}
	}
	if !foundContainer {
		common.Errorf("No init-ims container found in pod %s", podName)
		okay = false
	} else if okay {
		recipe.Name = recipeName
		recipe.Distro = recipeDistro
	}
	return
}
