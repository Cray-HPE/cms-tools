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
// Verify that the request succeeded and that the response had the expected status code
// Log any errors found
// If no errors are found, passed is set to true and the value of fatal is meaningless
// If the request fails but works if tried without authentication, then passed is
// set to false, but fatal is also set to false. This means the test can continue to run, since the
// request did work, but ultimately the test will be logged as a failure.
// Otherwise passed is set to false and fatal is set to true
func vcsRequest(requestType, requestUri, jsonString string, expectedStatusCode int) (passed, fatal bool) {
	var resp *resty.Response
	var dataArray []byte
	var insecure = false

	passed = false
	fatal = true

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
	} else {
		common.Infof("%s %s: expected status code %d and got it", requestType, requestUrl, expectedStatusCode)
	}
	if insecure {
		fatal = false
	} else {
		passed = true
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

	passed = true

	// Attempt to create a temporary organization
	orgName := "test-cmsdev-" + common.AlnumString(8)
	common.Infof("Create vcs organization named %s", orgName)
	orgDataJsonString := fmt.Sprintf(
		`{ "username": "%s", "visibility": "public", "description": "Test org created by cmsdev" }`,
		orgName)
	if ok, fatal := vcsPost("/orgs", orgDataJsonString); !ok {
		passed = false
		if fatal {
			return
		}
	}
	common.Infof("Org created successfully")

	// Query new org
	orgUri := "/orgs/" + orgName
	common.Infof("Query new vcs org")
	if ok, fatal := vcsGet(orgUri, http.StatusOK); !ok {
		passed = false
		if fatal {
			return
		}
	}
	common.Infof("Successfully queried new vcs org")

	// Attempt to create a repo in our org
	repoName := "harf-" + common.AlnumString(8)
	common.Infof("Create vcs repository named %s in organization %s", repoName, orgName)
	repoDataJsonString := fmt.Sprintf(
		`{ "name": "%s", "auto_init": null, "description": "Test repo created by cmsdev", "gitignores": null, "license": null, "private": false, "readme": null }`,
		repoName)
	if ok, fatal := vcsPost("/org/"+orgName+"/repos", repoDataJsonString); !ok {
		passed = false
		if fatal {
			return
		}
	}
	common.Infof("Repo created successfully")

	// Verify we can query new repo via API
	repoUri := "/repos/" + orgName + "/" + repoName
	common.Infof("Query new vcs repo")
	if ok, fatal := vcsGet(repoUri, http.StatusOK); !ok {
		passed = false
		if fatal {
			return
		}
	}
	common.Infof("Successfully queried new vcs repo")

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
	if ok, fatal := vcsDelete(repoUri); !ok {
		passed = false
		if fatal {
			common.Infof("Try to delete org")
			vcsDelete(orgUri)
			return
		}
	}
	common.Infof("Repo deleted successfully")

	// Try API query to verify that it fails
	common.Infof("Verify that API query of repo now returns not found status")
	if ok, fatal := vcsGet(repoUri, http.StatusNotFound); !ok && fatal {
		passed = false
	} else {
		common.Infof("Deleted repo not found by API query, as expected")
		if !ok {
			passed = false
		}
	}

	// Try a clone to verify that it now fails
	if !cloneShouldFail(repoName, repoUrl) {
		passed = false
	}

	// Delete vcs org
	common.Infof("Delete org %s", orgName)
	if ok, fatal := vcsDelete(orgUri); !ok {
		passed = false
		if fatal {
			return
		}
	}
	common.Infof("Org deleted successfully")

	// Try API query to verify that it fails
	common.Infof("Verify that API query of org now returns not found status")
	if ok, fatal := vcsGet(orgUri, http.StatusNotFound); !ok {
		passed = false
		if fatal {
			return
		}
	}
	common.Infof("Deleted repo not found by API query, as expected")

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
		common.Infof("No output from command %s", cmdOut)
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
