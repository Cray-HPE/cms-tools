package vcs

/*
 * vcs.go
 *
 * vcs repo test
 *
 * Copyright 2020-2021 Hewlett Packard Enterprise Development LP
 */

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty"
	"net/http"
	"os/exec"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

const VCSURL = common.BASEURL + "/vcs/api/v1"

// Perform a VCS request of the specified type to the specified uri, with the specified JSON data (if any)
// If the request fails, it is retried with authentication disabled.
// Verify that the request succeeded and that the response had the expected status code
// Log any errors found

// requestOk is set to true if the request worked as expected (either originally or on unauthenticated retry)
// Otherwise it is set to false
// errorsFound is set to false if there were no errors, otherwise it is set to true.
// This is of course the case if the request didn't work,
// but it is also the case if the request only worked after being retried without authentication.
// The latter case means the test can continue to run, since therequest did work, but ultimately the test will
// be logged as a failure.
func vcsRequest(requestType, requestUri, jsonString string, expectedStatusCode int) (requestOk, errorsFound bool) {
	var resp *resty.Response
	var dataArray []byte
	var insecure = false

	requestOk = false
	errorsFound = true

	// Get vcs user and password
	common.Infof("Getting vcs user and password")
	vcsUser, vcsPass, err := k8s.GetVcsUsernamePassword()
	if err != nil {
		common.Error(err)
		return
	}

	client := resty.New()
	client.SetHeader("Content-Type", "application/json")
	client.SetBasicAuth(vcsUser, vcsPass)

	requestUrl := VCSURL + requestUri

	common.Infof("%s %s", requestType, requestUrl)
	if len(jsonString) > 0 {
		dataArray = []byte(jsonString)
		common.Infof("data: %s", jsonString)
	}

	for true {
		if requestType == "DELETE" {
			if len(jsonString) > 0 {
				resp, err = client.R().SetBody(dataArray).Delete(requestUrl)
			} else {
				resp, err = client.R().Delete(requestUrl)
			}
		} else if requestType == "GET" {
			if len(jsonString) > 0 {
				resp, err = client.R().SetBody(dataArray).Get(requestUrl)
			} else {
				resp, err = client.R().Get(requestUrl)
			}
		} else if requestType == "POST" {
			if len(jsonString) > 0 {
				resp, err = client.R().SetBody(dataArray).Post(requestUrl)
			} else {
				resp, err = client.R().Post(requestUrl)
			}
		} else {
			common.Errorf("PROGRAMMING LOGIC ERROR: Invalid request type: %s", requestType)
			return
		}

		if err != nil {
			common.Error(err)
			if insecure {
				// Failed even with verify set to false
				return
			}
			common.Infof("This is a failure, but will retry request with InsecureSkipVerify set to true")
			client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
			insecure = true
			common.Infof("InsecureSkipVerify=true %s %s", requestType, requestUrl)
			continue
		}
		break
	}

	common.PrettyPrintJSON(resp)
	if resp.StatusCode() != expectedStatusCode {
		common.Errorf("%s %s: expected status code %d, got %d", requestType, requestUrl, expectedStatusCode, resp.StatusCode())
		return
	}
	common.Infof("%s %s: expected status code %d and got it", requestType, requestUrl, expectedStatusCode)
	requestOk = true
	if !insecure {
		errorsFound = false
	}
	return
}

// Wrapper for vcs delete calls. In this test, these calls never include JSON data and
// always expect StatusNoContent, so those are set appropriately.
func vcsDelete(requestUri string) (bool, bool) {
	return vcsRequest("DELETE", requestUri, "", http.StatusNoContent)
}

// Wrapper for vcs get calls. In this test, these calls never include JSON data, so
// that is set to an empty string
func vcsGet(requestUri string, expectedStatusCode int) (bool, bool) {
	return vcsRequest("GET", requestUri, "", expectedStatusCode)
}

// Wrapper for vcs delete calls. In this test, these calls always expect StatusCreated,
// so that is set appropriately.
func vcsPost(requestUri, jsonString string) (bool, bool) {
	return vcsRequest("POST", requestUri, jsonString, http.StatusCreated)
}

// Does the following:
// 1) Creates vcs organization via API
// 2) Queries the org via API
// 3) Creates vcs repo in new org via API
// 4) Queries new repo via API
// 5) Performs the cloneTest() function on the new repo
// 6) Delete the new repo via API
// 7) Queries deleted repo via API to verify it is not found
// 8) Performs the cloneTest() function on the repo and verifies it fails to clone
// 9) Deletes the new org via API
// 10) Queries the new org via API to verify it is not found
func repoTest() (passed bool) {
	var repoUrl, vcsUser, vcsPass string
	var requestOk, errorsFound bool

	passed = true

	// Attempt to create a temporary organization
	orgName := "test-cmsdev-" + common.AlnumString(8)
	common.Infof("Create vcs organization named %s", orgName)
	orgDataJsonString := fmt.Sprintf(
		`{ "username": "%s", "visibility": "public", "description": "Test org created by cmsdev" }`,
		orgName)
	requestOk, errorsFound = vcsPost("/orgs", orgDataJsonString)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Org created successfully")
	} else {
		common.Errorf("Failed to create vcs organization")
		return
	}

	// Query new org
	orgUri := "/orgs/" + orgName
	common.Infof("Query new vcs org")
	requestOk, errorsFound = vcsGet(orgUri, http.StatusOK)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Successfully queried new vcs org")
	} else {
		common.Errorf("Failed to query vcs organization")
		// Despite the query failure, we will proceed with the test, since the
		// create request did appear to work
	}

	// Attempt to create a repo in our org
	repoName := "harf-" + common.AlnumString(8)
	common.Infof("Create vcs repository named %s in organization %s", repoName, orgName)
	repoDataJsonString := fmt.Sprintf(
		`{ "name": "%s", "auto_init": null, "description": "Test repo created by cmsdev", "gitignores": null, "license": null, "private": false, "readme": null }`,
		repoName)
	requestOk, errorsFound = vcsPost("/org/"+orgName+"/repos", repoDataJsonString)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Repo created successfully")
	} else {
		common.Errorf("Failed to create vcs repo")
		common.Infof("Try to delete org")
		vcsDelete(orgUri)
		return
	}

	// Verify we can query new repo via API
	repoUri := "/repos/" + orgName + "/" + repoName
	common.Infof("Query new vcs repo")
	requestOk, errorsFound = vcsGet(repoUri, http.StatusOK)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Successfully queried new vcs repo")
	} else {
		common.Errorf("Failed to query vcs repo")
		// Despite the query failure, we will proceed with the test, since the
		// create request did appear to work
	}

	// Get vcs user and password
	common.Infof("Getting vcs user and password")
	vcsUser, vcsPass, err := k8s.GetVcsUsernamePassword()
	if err != nil {
		common.Error(err)
		passed = false
	} else {
		repoUrl = fmt.Sprintf("https://%s:%s@%s/vcs/%s/%s.git", vcsUser, vcsPass, common.BASEHOST, orgName, repoName)
		if !cloneTest(repoName, repoUrl) {
			passed = false
		}
	}

	// Delete repo
	common.Infof("Delete repo %s", repoName)
	requestOk, errorsFound = vcsDelete(repoUri)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Successfully deleted vcs repo")
	} else {
		common.Errorf("Failed to delete vcs repo")
		common.Infof("Try to delete org anyway")
		vcsDelete(orgUri)
		return
	}

	// Try API query to verify that it fails
	common.Infof("Verify that API query of repo now returns not found status")
	requestOk, errorsFound = vcsGet(repoUri, http.StatusNotFound)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Deleted repo not found by API query, as expected")
	} else {
		common.Errorf("Deleted repo was found by API query")
		common.Infof("Try to delete org anyway")
		vcsDelete(orgUri)
		return
	}

	// Try a clone to verify that it now fails
	if !cloneShouldFail(repoName, repoUrl) {
		passed = false
	}

	// Delete vcs org
	common.Infof("Delete org %s", orgName)
	requestOk, errorsFound = vcsDelete(orgUri)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Successfully deleted vcs org")
	} else {
		common.Errorf("Failed to delete vcs org")
		return
	}

	// Try API query to verify that it fails
	common.Infof("Verify that API query of org now returns not found status")
	requestOk, errorsFound = vcsGet(orgUri, http.StatusNotFound)
	passed = passed && requestOk && !errorsFound
	if requestOk {
		common.Infof("Deleted org not found by API query, as expected")
	} else {
		common.Errorf("Deleted org was found by API query")
	}

	return
}

// Looks up the path of the specified command
// Logs errors, if any
// Returns the path and true if successful, empty string and false otherwise
func getPath(cmdName string) (string, bool) {
	path, err := exec.LookPath(cmdName)
	if err != nil {
		common.Error(err)
		return "", false
	} else if len(path) == 0 {
		common.Errorf("Empty path found for %s", cmdName)
		return "", false
	}
	return path, true
}

// Wrapper to run a command and verify it passes or fails
// Logs errors if any
// Returns true if no errors, false otherwise
func runCmd(shouldPass bool, cmdName string, cmdArgs ...string) bool {
	var cmd *exec.Cmd
	var cmdOut []byte
	var cmdPath string
	var err error
	var ok bool

	cmdPath, ok = getPath(cmdName)
	if !ok {
		return false
	}

	cmd = exec.Command(cmdPath, cmdArgs...)
	if shouldPass {
		common.Infof("Running command: %s", cmd)
	} else {
		common.Infof("RUNNING COMMAND WE EXPECT TO FAIL: %s", cmd)
	}
	cmdOut, err = cmd.CombinedOutput()
	if len(cmdOut) > 0 {
		common.Infof("Command output: %s", cmdOut)
	} else {
		common.Infof("No output from command")
	}
	if err != nil {
		if shouldPass {
			common.Error(err)
			return false
		} else {
			common.Infof("Command error: %s", err.Error())
			return true
		}
	} else if shouldPass {
		return true
	} else {
		common.Errorf("Command passed but we expected it to fail")
		return false
	}
}

// Does the following:
// 1) Clones repo
// 2) Sets git user.email and user.name to avoid errors/warnings
// 3) Adds, commits, and pushes a file
// 4) Deletes local copy of repo
// 5) Re-clones repo
// 6) Verifies that committed file is present and identical to the original
// 7) Deletes the local copy of repo
// Logs errors if any
// Returns true if no errors, false otherwise
func cloneTest(repoName, repoUrl string) bool {
	var baseLsPath, repoDir, repoLsPath string
	var ok bool

	repoDir = fmt.Sprintf("/tmp/%s", repoName)

	if !runCmd(true, "git", "clone", repoUrl, repoDir) {
		return false
	}

	// set user.email and user.name in cloned repo, to avoid errors/warnings
	if !runCmd(true, "git", "-C", repoDir, "config", "user.email", "catfood@dogfood.gov") {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	if !runCmd(true, "git", "-C", repoDir, "config", "user.name", "Joseph Catfood") {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	baseLsPath, ok = getPath("ls")
	if !ok {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}
	// Copy ls file into repo
	if !runCmd(true, "cp", baseLsPath, repoDir) {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	// git add
	if !runCmd(true, "git", "-C", repoDir, "add", ".") {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	// git commit
	if !runCmd(true, "git", "-C", repoDir, "commit", "-m", "Test commit") {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	// git push
	if !runCmd(true, "git", "-C", repoDir, "push") {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	// Remove the cloned directory
	if !runCmd(true, "rm", "-r", repoDir) {
		return false
	}

	// Clone into new directory
	repoDir = fmt.Sprintf("/tmp/%s-take2", repoName)
	if !runCmd(true, "git", "clone", repoUrl, repoDir) {
		return false
	}

	// Compare the existing ls file to the one in the repo
	repoLsPath = fmt.Sprintf("%s/ls", repoDir)
	if !runCmd(true, "cmp", baseLsPath, repoLsPath) {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	// Remove the repo dir
	if !runCmd(true, "rm", "-r", repoDir) {
		return false
	}
	return true
}

func cloneShouldFail(repoName, repoUrl string) bool {
	var repoDir string

	repoDir = fmt.Sprintf("/tmp/%s-shouldfail", repoName)
	if runCmd(false, "git", "clone", repoUrl, repoDir) {
		return true
	}

	// Cleanup the repo dir, since the clone was unexpectedly successful
	runCmd(true, "rm", "-r", repoDir)
	return false
}
