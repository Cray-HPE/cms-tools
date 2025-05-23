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
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
)

/*
 * ims_recipes_cli_test.go
 *
 * ims recipes cli test functions
 *
 */

func TestRecipeCRUDOperationUsingCLI() (passed bool) {
	common.PrintLog("Testing recipe CRUD operations using CLI")
	// Test creating a recipe
	recipeRecord, success := TestCLIRecipeCreate()
	if !success {
		return false
	}
	// Test updating the recipe
	updated := TestCLIRecipeUpdate(recipeRecord.Id)

	// Test soft deleting the recipe
	deleted := TestCLIRecipeDelete(recipeRecord.Id)

	// Test undeleting the recipe
	undeleted := TestCLIRecipeUndelete(recipeRecord.Id)

	// Test permanent deleting the recipe
	permDeleted := TestCLIRecipePermanentDelete(recipeRecord.Id)

	// Test get all recipes
	getAll := TestCLIGetAllRecipes()

	return updated && deleted && undeleted && permDeleted && getAll
}

func TestCLIRecipeCreate() (recipeRecord IMSRecipeRecord, passed bool) {

	// Create the recipe
	recipeName := "recipe_" + string(common.GetRandomString(10))

	expectedTemplatesDict := []map[string]string{
		{
			"key":   "USS_VERSION",
			"value": "1.1.2-1-cos-base-3.1",
		},
		{
			"key":   "FULL_COS_BASE_VERSION",
			"value": "3.1.2-1-sle-15.5",
		},
	}

	recipeRecord, success := CreateIMSRecipeRecordCLI(recipeName)
	if !success {
		return IMSRecipeRecord{}, false
	}
	// Get the recipe record
	recipeRecord, success = getIMSRecipeRecordCLI(recipeRecord.Id)
	if !success || recipeRecord.Name != recipeName {
		common.Errorf("Recipe %s was not created with name %s", recipeRecord.Id, recipeName)
		return IMSRecipeRecord{}, false
	}

	// Verify the recipe metadata
	if !common.CompareSlicesOfMaps(recipeRecord.Template_dictionary, expectedTemplatesDict) {
		common.Errorf("Expected template dictionary %v, got %v", expectedTemplatesDict, recipeRecord.Template_dictionary)
		return IMSRecipeRecord{}, false
	}

	// Verfy the recipe is in the list of recipes
	recipeRecords, success := getIMSRecipeRecordsCLI()
	if !success {
		return IMSRecipeRecord{}, false
	}

	if !RecipeRecordExists(recipeRecord.Id, recipeRecords) {
		common.Errorf("Recipe %s was not found in the list of recipes", recipeRecord.Id)
		return IMSRecipeRecord{}, false
	}

	common.Infof("Created recipe ID %s with name %s", recipeRecord.Id, recipeName)
	return recipeRecord, true
}

func TestCLIRecipeUpdate(recipeId string) (passed bool) {
	arch := "aarch64"
	expectedTemplatesDict := []map[string]string{
		{
			"key":   "USS_VERSION",
			"value": "1.1.2-1-cos-3.1",
		},
	}
	if _, success := UpdateIMSRecipeRecordCLI(recipeId, arch); !success {
		return false
	}
	// Get the recipe record
	recipeRecord, success := getIMSRecipeRecordCLI(recipeId)
	if !success || recipeRecord.Arch != arch {
		common.Errorf("Recipe %s was not updated with arch %s", recipeId, arch)
		return false
	}

	// Verify the recipe metadata
	if !common.CompareSlicesOfMaps(recipeRecord.Template_dictionary, expectedTemplatesDict) {
		common.Errorf("Expected template dictionary %v, got %v", expectedTemplatesDict, recipeRecord.Template_dictionary)
		return false
	}
	common.Infof("Updated recipe ID %s with arch %s", recipeRecord.Id, arch)
	return true
}

func TestCLIRecipeDelete(recipeId string) (passed bool) {
	if success := DeleteIMSRecipeRecordCLI(recipeId); !success {
		return false
	}
	// Verify the recipe is soft deleted
	if _, success := GetDeletedIMSRecipeRecordCLI(recipeId); !success {
		common.Errorf("Recipe %s was not soft deleted", recipeId)
		return false
	}

	// Set the CLI execution return code to 2. Since the recipe is deleted, the command should return 2.
	test.SetCliExecreturnCode(2)

	// Verify the recipe is not in the list of recipes
	if _, success := getIMSRecipeRecordCLI(recipeId); success {
		common.Errorf("Recipe %s was not soft deleted", recipeId)
		return false
	}

	// Set the CLI execution return code to 0.
	test.SetCliExecreturnCode(0)

	// Verify the recipe is not in the list of all recipes
	recipeRecords, success := getIMSRecipeRecordsCLI()
	if !success {
		return false
	}

	if RecipeRecordExists(recipeId, recipeRecords) {
		common.Errorf("Recipe %s was not soft deleted", recipeId)
		return false
	}

	// verify the recipe is in the list of all deleted recipes
	deletedRecipeRecords, success := GetDeletedIMSRecipeRecordsCLI()
	if !success {
		return false
	}

	if !RecipeRecordExists(recipeId, deletedRecipeRecords) {
		common.Errorf("Recipe %s was not found in the list of deleted recipes", recipeId)
		return false
	}

	common.Infof("Soft deleted recipe ID %s", recipeId)
	return true
}

func TestCLIRecipeUndelete(recipeId string) (passed bool) {
	if success := UndeleteIMSRecipeRecordCLI(recipeId); !success {
		return false
	}
	// Verify the recipe is undeleted
	if _, success := getIMSRecipeRecordCLI(recipeId); !success {
		common.Errorf("Recipe %s was not restored", recipeId)
		return false
	}
	common.Infof("Restored recipe ID %s", recipeId)
	return true
}

func TestCLIRecipePermanentDelete(recipeId string) (passed bool) {
	// Soft delete the recipe
	if success := DeleteIMSRecipeRecordCLI(recipeId); !success {
		return false
	}

	if success := PermanentDeleteIMSRecipeRecordCLI(recipeId); !success {
		return false
	}

	// Set the CLI execution return code to 2. Since the recipe is deleted, the command should return 2.
	test.SetCliExecreturnCode(2)

	// Verify the recipe is hard deleted
	if _, success := GetDeletedIMSRecipeRecordCLI(recipeId); success {
		common.Errorf("Recipe %s was not permanently deleted", recipeId)
		return false
	}

	// Verify the recipe is not in the list of recipes
	if _, success := getIMSRecipeRecordCLI(recipeId); success {
		common.Errorf("Recipe %s was not permanently deleted", recipeId)
		return false
	}

	// Set the CLI execution return code to 0.
	test.SetCliExecreturnCode(0)

	// Verify the recipe is not in the list of all recipes
	recipeRecords, success := getIMSRecipeRecordsCLI()
	if !success {
		return false
	}

	if RecipeRecordExists(recipeId, recipeRecords) {
		common.Errorf("Recipe %s was not permanently deleted", recipeId)
		return false
	}

	// verify the recipe is in the list of all deleted recipes
	deletedRecipeRecords, success := GetDeletedIMSRecipeRecordsCLI()
	if !success {
		return false
	}

	if RecipeRecordExists(recipeId, deletedRecipeRecords) {
		common.Errorf("Recipe %s was not permanently deleted", recipeId)
		return false
	}

	common.Infof("Permanently deleted recipe ID %s", recipeId)
	return true
}

func TestCLIGetAllRecipes() (passed bool) {
	if _, success := getIMSRecipeRecordsCLI(); !success {
		return false
	}
	common.Infof("Got all recipes")
	return true
}
