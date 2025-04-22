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
package ims

import (
	"net/http"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

/*
 * ims_recipes_test.go
 *
 * ims recipes test api functions
 *
 */

func TestRecipeCRUDOperation() (passed bool) {
	// Test creating a recipe
	recipeRecord, success := TestRecipeCreate()
	if !success {
		return false
	}
	// Test updating the recipe
	updated := TestRecipeUpdate(recipeRecord.Id)

	// Test soft deleting the recipe
	deleted := TestRecipeDelete(recipeRecord.Id)

	// Test undeleting the recipe
	undeleted := TestRecipeUndelete(recipeRecord.Id)

	// Test hard deleting the recipe
	permDeleted := TestRecipePermanentDelete(recipeRecord.Id)

	// Test get all recipes
	getAll := TestGetAllRecipes()

	return updated && deleted && undeleted && permDeleted && getAll
}

func TestRecipePermanentDelete(recipeId string) (passed bool) {
	// Soft delete the recipe
	if success := DeleteIMSRecipeRecordAPI(recipeId); !success {
		return false
	}

	// Permanently delete the recipe
	if success := PermanentDeleteIMSRecipeRecordAPI(recipeId); !success {
		return false
	}
	// Verify the recipe is hard deleted
	if _, success := GetDeletedIMSRecipeRecordAPI(recipeId, http.StatusNotFound); !success {
		common.Errorf("Recipe %s was not permanently deleted", recipeId)
		return false
	}
	// Verify the recipe is not in the list of recipes
	if _, success := GetIMSRecipeRecordAPI(recipeId, http.StatusNotFound); success {
		common.Errorf("Recipe %s was not permanently deleted", recipeId)
		return false
	}
	common.Infof("Recipe %s was permanently deleted", recipeId)
	return true
}

func TestRecipeCreate() (recipeRecord IMSRecipeRecord, passed bool) {
	recipeName := "recipe_" + string(common.GetRandomString(10))
	templatesDict := []map[string]string{
		{
			"key":   "USS_VERSION",
			"value": "1.1.2-1-cos-base-3.1",
		},
		{
			"key":   "FULL_COS_BASE_VERSION",
			"value": "3.1.2-1-sle-15.5",
		},
	}
	requireDKMS := false
	recipeRecord, success := CreateIMSRecipeRecordAPI(recipeName, templatesDict, requireDKMS)
	if !success {
		return IMSRecipeRecord{}, false
	}

	// Verify the recipe is created
	recipeRecord, success = GetIMSRecipeRecordAPI(recipeRecord.Id, http.StatusOK)
	if !success ||
		recipeRecord.Name != recipeName ||
		!common.CompareSlicesOfMaps(recipeRecord.Template_dictionary, templatesDict) ||
		recipeRecord.Require_dkms != requireDKMS {

		common.Errorf("Recipe %s was not created", recipeName)
		return IMSRecipeRecord{}, false
	}
	common.Infof("Recipe %s was created with id %s", recipeName, recipeRecord.Id)
	return recipeRecord, true
}

func TestRecipeUpdate(recipeId string) (passed bool) {
	arch := "aarch64"
	templatesDict := []map[string]string{
		{
			"key":   "USS_VERSION",
			"value": "1.1.2-1-cos-3.1",
		},
	}
	if _, success := UpdateIMSRecipeRecordAPI(recipeId, arch, templatesDict); !success {
		return false
	}

	// Verify the recipe is updated
	recipeRecord, success := GetIMSRecipeRecordAPI(recipeId, http.StatusOK)
	if !success ||
		recipeRecord.Arch != arch ||
		!common.CompareSlicesOfMaps(recipeRecord.Template_dictionary, templatesDict) {
		common.Errorf("Recipe %s was not updated", recipeId)
		return false
	}
	common.Infof("Recipe %s was updated with arch %s", recipeId, arch)
	return true
}

func TestRecipeDelete(recipeId string) (passed bool) {
	if success := DeleteIMSRecipeRecordAPI(recipeId); !success {
		return false
	}

	// Verify the recipe is deleted
	if _, success := GetDeletedIMSRecipeRecordAPI(recipeId, http.StatusOK); !success {
		common.Errorf("Recipe %s was not deleted", recipeId)
		return false
	}
	common.Infof("Recipe %s was deleted", recipeId)
	return true
}

func TestRecipeUndelete(recipeId string) (passed bool) {
	if success := UndeleteIMSRecipeRecordAPI(recipeId); !success {
		return false
	}

	// Verify the recipe is undeleted
	if _, success := GetIMSRecipeRecordAPI(recipeId, http.StatusOK); !success {
		common.Errorf("Recipe %s was not restored", recipeId)
		return false
	}
	common.Infof("Recipe %s was restored", recipeId)
	return true
}

func TestGetAllRecipes() (passed bool) {
	if _, success := GetIMSRecipeRecordsAPI(); !success {
		return false
	}
	common.Infof("All recipes were retrieved")
	return true
}
