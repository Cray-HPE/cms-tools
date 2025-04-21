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
)

/*
 * ims_recipes_cli_test.go
 *
 * ims recipes cli test functions
 *
 */

func TestRecipeCRUDOperationUsingCLI() (passed bool) {
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

	// Test hard deleting the recipe
	hardDeleted := TestCLIRecipeHardDelete(recipeRecord.Id)

	// Test get all recipes
	getAll := TestCLIGetAllRecipes()

	return updated && deleted && undeleted && hardDeleted && getAll
}

func TestCLIRecipeCreate() (recipeRecord IMSRecipeRecord, passed bool) {

	// Create the recipe
	recipeName := "recipe_" + string(common.GetRandomString(10))

	recipeRecord, success := CreateIMSRecipeRecordCLI(recipeName)
	if !success {
		return IMSRecipeRecord{}, false
	}
	// Get the recipe record
	recipeRecord, success = getIMSRecipeRecordCLI(recipeRecord.Id)
	if !success {
		return IMSRecipeRecord{}, false
	}
	common.Infof("Created recipe ID %s with name %s", recipeRecord.Id, recipeName)
	return recipeRecord, true
}

func TestCLIRecipeUpdate(recipeId string) (passed bool) {
	arch := "aarch64"
	if _, success := UpdateIMSRecipeRecordCLI(recipeId, arch); !success {
		return false
	}
	// Get the recipe record
	recipeRecord, success := getIMSRecipeRecordCLI(recipeId)
	if !success || recipeRecord.Arch != arch {
		common.Errorf("Recipe %s was not updated with arch %s", recipeId, arch)
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

func TestCLIRecipeHardDelete(recipeId string) (passed bool) {
	// Soft delete the recipe
	if success := DeleteIMSRecipeRecordCLI(recipeId); !success {
		return false
	}

	if success := HardDeleteIMSRecipeRecordCLI(recipeId); !success {
		return false
	}
	// Verify the recipe is hard deleted
	if _, success := GetDeletedIMSRecipeRecordCLI(recipeId); success {
		common.Errorf("Recipe %s was not hard deleted", recipeId)
		return false
	}
	// Verify the recipe is not in the list of recipes
	if _, success := getIMSRecipeRecordCLI(recipeId); success {
		common.Errorf("Recipe %s was not hard deleted", recipeId)
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
