// MIT License
//
// (C) Copyright 2020-2025 Hewlett Packard Enterprise Development LP
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
package vcs

/*
 * vcs.go
 *
 * vcs repo test
 *
 */

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	resty "gopkg.in/resty.v1"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
)

const VCSURL = common.BASEURL + "/vcs/api/v1"

var useInsecure bool = false

type GitRepo struct {
	RepoName   string
	Host       string
	VcsOrgName string
	Username   string
	Password   string
}

func (gitRepo *GitRepo) Url() string {
	return fmt.Sprintf("https://%s/vcs/%s/%s.git", gitRepo.Host, gitRepo.VcsOrgName, gitRepo.RepoName)
}

func (gitRepo *GitRepo) GitCmdEnvVars() map[string]string {
	return map[string]string{
		"GIT_CONFIG_COUNT":   "2",
		"GIT_CONFIG_KEY_0":   fmt.Sprintf("credential.https://%s.username", gitRepo.Host),
		"GIT_CONFIG_VALUE_0": gitRepo.Username,
		"GIT_CONFIG_KEY_1":   fmt.Sprintf("credential.https://%s.helper", gitRepo.Host),
		"GIT_CONFIG_VALUE_1": fmt.Sprintf("!f() { test \"$1\" = get && echo \"password=%s\"; }; f", gitRepo.Password),
	}
}

// When running git commands, just like with the vcs requests we want to retry them insecurely
// if needed
func (gitRepo *GitRepo) runGitCmd(shouldPass bool, cmdArgs ...string) bool {
	var cmdResult *common.CommandResult
	var err error
	var tryInsecure bool
	var newCmdArgs []string

	tryInsecure = false

	if !shouldPass {
		// Because this command should not work anyway, we will run it insecurely,
		// to avoid failures that are because of that

		// Do some lovely golang gymnastics to prepend two strings to the
		// variadic cmdArgs argument
		newCmdArgs = make([]string, 0, 2+len(cmdArgs))
		newCmdArgs = append(append(newCmdArgs, "-c", "http.sslVerify=false"), cmdArgs...)
		return runCmdWithEnv(shouldPass, gitRepo.GitCmdEnvVars(), "git", newCmdArgs...)
	} else if !useInsecure {
		cmdResult, err = common.RunNameWithEnv(gitRepo.GitCmdEnvVars(), "git", cmdArgs...)
		if err != nil {
			common.Error(err)
			common.Errorf("Error attempting to run command")
			return false
		} else if cmdResult.Rc == 0 {
			return true
		}
		// Check to see if this may be an SSL error
		if strings.Contains(cmdResult.ErrString(), "SSL") {
			common.Infof("Command failure may be an SSL issue. Will retry insecurely.")
			common.Infof("Important: This means the overall test will fail even if this command works on retry!")
			tryInsecure = true
		} else {
			common.Errorf("Command failed")
			return false
		}
	}
	// Run the command insecurely
	newCmdArgs = make([]string, 0, 2+len(cmdArgs))
	newCmdArgs = append(append(newCmdArgs, "-c", "http.sslVerify=false"), cmdArgs...)
	ok := runCmdWithEnv(shouldPass, gitRepo.GitCmdEnvVars(), "git", newCmdArgs...)
	if ok && tryInsecure {
		// Let's remember that we had to run this command insecurely in order for it to work, so that
		// we'll be insecure for our future git commands in this test
		useInsecure = true
	}
	return ok
}

// Perform a VCS request of the specified type to the specified uri, with the specified JSON data (if any)
// If the request fails, it is retried with authentication disabled.
// Verify that the request succeeded and that the response had the expected status code
// Log any errors found
// Returns true if the request worked as expected (either originally or on unauthenticated retry)
// Otherwise returns false
func vcsRequest(requestType, requestUri, jsonString string, expectedStatusCode int) (ok bool) {
	var resp *resty.Response
	var dataArray []byte
	var tryInsecure bool

	ok = false
	tryInsecure = false

	// Get vcs user and password
	common.Debugf("Getting vcs user and password")
	vcsUser, vcsPass, err := k8s.GetVcsUsernamePassword()
	if err != nil {
		common.Error(err)
		return
	}

	client := resty.New()
	client.SetTimeout(common.API_TIMEOUT_SECONDS)
	client.SetRetryCount(common.API_RETRY_COUNT)
	client.SetRetryWaitTime(common.API_RETRY_WAIT_SECONDS * time.Second)
	client.SetHeader("Content-Type", "application/json")
	client.SetBasicAuth(vcsUser, vcsPass)

	// Add retry condition for HTTP 503 status code
	client.AddRetryCondition(func(r *resty.Response) (bool, error) {
		return r.StatusCode() == 503, errors.New("Received HTTP 503 from server, retrying...")
	})
	requestUrl := VCSURL + requestUri

	for true {
		if useInsecure || tryInsecure {
			common.Infof("InsecureSkipVerify=true %s %s", requestType, requestUrl)
			client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		} else {
			common.Infof("%s %s", requestType, requestUrl)
		}
		if len(jsonString) > 0 {
			dataArray = []byte(jsonString)
			common.Debugf("data: %s", jsonString)
		}
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
		common.PrettyPrintJSON(resp)

		if err != nil {
			common.Error(err)
			if useInsecure || tryInsecure {
				// Failed even with verify set to false
				return
			}
			common.Infof("This is a failure, but will retry request with InsecureSkipVerify set to true")
			common.Infof("Important: This means the overall test will fail even if this request works on retry!")
			client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
			tryInsecure = true
			continue
		}
		break
	}

	if resp.StatusCode() != expectedStatusCode {
		common.Errorf("%s %s: expected status code %d, got %d", requestType, requestUrl, expectedStatusCode, resp.StatusCode())
		return
	}
	common.Infof("%s %s: expected status code %d and got it", requestType, requestUrl, expectedStatusCode)
	if tryInsecure {
		// Let's remember that we had to use insecure in order for this to work, so we won't bother
		// trying securely for future requests in this test
		useInsecure = true
	}
	ok = true
	return
}

// Wrapper for vcs delete calls. In this test, these calls never include JSON data and
// always expect StatusNoContent, so those are set appropriately.
func vcsDelete(requestUri string) bool {
	return vcsRequest("DELETE", requestUri, "", http.StatusNoContent)
}

// Wrapper for vcs get calls. In this test, these calls never include JSON data, so
// that is set to an empty string
func vcsGet(requestUri string, expectedStatusCode int) bool {
	return vcsRequest("GET", requestUri, "", expectedStatusCode)
}

// Wrapper for vcs delete calls. In this test, these calls always expect StatusCreated,
// so that is set appropriately.
func vcsPost(requestUri, jsonString string) bool {
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
	var vcsUser, vcsPass string

	passed = true

	// Attempt to create a temporary organization
	orgName := "test-cmsdev-" + common.AlnumString(8)
	common.Infof("Create vcs organization named %s", orgName)
	orgDataJsonString := fmt.Sprintf(
		`{ "username": "%s", "visibility": "public", "description": "Test org created by cmsdev" }`,
		orgName)
	if ok := vcsPost("/orgs", orgDataJsonString); ok {
		common.Infof("Org created successfully")
	} else {
		passed = false
		common.Errorf("Failed to create vcs organization")
		return
	}

	// Query new org
	orgUri := "/orgs/" + orgName
	common.Infof("Query new vcs org")
	if ok := vcsGet(orgUri, http.StatusOK); ok {
		common.Infof("Successfully queried new vcs org")
	} else {
		passed = false
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
	if ok := vcsPost("/org/"+orgName+"/repos", repoDataJsonString); ok {
		common.Infof("Repo created successfully")
	} else {
		passed = false
		common.Errorf("Failed to create vcs repo")
		common.Infof("Try to delete org")
		vcsDelete(orgUri)
		return
	}

	// Verify we can query new repo via API
	repoUri := "/repos/" + orgName + "/" + repoName
	common.Infof("Query new vcs repo")
	if ok := vcsGet(repoUri, http.StatusOK); ok {
		common.Infof("Successfully queried new vcs repo")
	} else {
		passed = false
		common.Errorf("Failed to query vcs repo")
		// Despite the query failure, we will proceed with the test, since the
		// create request did appear to work
	}

	gitRepo := new(GitRepo)
	gitRepo.Host = common.BASEHOST
	gitRepo.VcsOrgName = orgName
	gitRepo.RepoName = repoName

	// Get vcs user and password
	common.Debugf("Getting vcs user and password")
	vcsUser, vcsPass, err := k8s.GetVcsUsernamePassword()
	if err != nil {
		common.Error(err)
		passed = false
		common.Infof("Unable to perform clone test without vcs user credentials")
	} else {
		gitRepo.Username = vcsUser
		gitRepo.Password = vcsPass
		if !gitRepo.cloneTest() {
			passed = false
		}
	}

	// Delete repo
	common.Infof("Delete repo %s", repoName)
	if ok := vcsDelete(repoUri); ok {
		common.Infof("Successfully deleted vcs repo")
	} else {
		passed = false
		common.Errorf("Failed to delete vcs repo")
		common.Infof("Try to delete org anyway")
		vcsDelete(orgUri)
		return
	}

	// Try API query to verify that it fails
	common.Infof("Verify that API query of repo now returns not found status")
	if ok := vcsGet(repoUri, http.StatusNotFound); ok {
		common.Infof("Deleted repo not found by API query, as expected")
	} else {
		passed = false
		common.Errorf("Deleted repo was found by API query")
		common.Infof("Try to delete org anyway")
		vcsDelete(orgUri)
		return
	}

	// Try a clone to verify that it now fails
	if len(gitRepo.Username) == 0 {
		common.Infof("Unable to perform bad path clone test without vcs user credentials")
	} else if !gitRepo.cloneShouldFail() {
		passed = false
	}

	// Delete vcs org
	common.Infof("Delete org %s", orgName)
	if ok := vcsDelete(orgUri); ok {
		common.Infof("Successfully deleted vcs org")
	} else {
		passed = false
		common.Errorf("Failed to delete vcs org")
		return
	}

	// Try API query to verify that it fails
	common.Infof("Verify that API query of org now returns not found status")
	if ok := vcsGet(orgUri, http.StatusNotFound); ok {
		common.Infof("Deleted org not found by API query, as expected")
	} else {
		passed = false
		common.Errorf("Deleted org was found by API query")
	}

	// We return failure if we had to use insecure requests or git operations
	if passed && useInsecure {
		common.Errorf("Even though all operations succeeded, test failed because insecure operations were required")
		passed = false
	}

	return
}

// Wrapper to run a command and verify it passes or fails
// Logs errors if any
// Returns true if no errors, false otherwise
func runCmdWithEnv(shouldPass bool, cmdEnv map[string]string, cmdName string, cmdArgs ...string) bool {
	var cmdResult *common.CommandResult
	var err error
	if !shouldPass {
		common.Debugf("WE EXPECT THE FOLLOWING %s COMMAND TO FAIL", cmdName)
	}
	cmdResult, err = common.RunNameWithEnv(cmdEnv, cmdName, cmdArgs...)
	if err != nil {
		common.Error(err)
		common.Errorf("Error attempting to run command")
		return false
	} else if cmdResult.Rc != 0 {
		if !cmdResult.Ran || shouldPass {
			common.Errorf("Command failed")
			return false
		} else {
			return true
		}
	} else if shouldPass {
		return true
	} else {
		common.Errorf("Command passed but we expected it to fail")
		return false
	}
}

// Wrapper to runCmdWithEnv with no env variables set
func runCmd(shouldPass bool, cmdName string, cmdArgs ...string) bool {
	return runCmdWithEnv(shouldPass, nil, cmdName, cmdArgs...)
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
func (gitRepo *GitRepo) cloneTest() bool {
	var repoDir string
	var err error

	repoDir = fmt.Sprintf("/tmp/%s", gitRepo.RepoName)

	if !gitRepo.runGitCmd(true, "clone", gitRepo.Url(), repoDir) {
		return false
	}

	// set user.email and user.name in cloned repo, to avoid errors/warnings
	if !gitRepo.runGitCmd(true, "-C", repoDir, "config", "user.email", "catfood@dogfood.gov") {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	if !gitRepo.runGitCmd(true, "-C", repoDir, "config", "user.name", "Joseph Catfood") {
		runCmd(true, "rm", "-r", repoDir)
		return false
	}

	// Generate text file
	textFileBase := fmt.Sprintf("vcs_repo_test_textfile.%d.%d.txt", common.Intn(1000000), common.Intn(1000000))
	textFile := "/tmp/" + textFileBase
	textFileSizeBytes := common.IntInRange(1, 20*1024)
	common.Debugf("Creating file to put in new repo: %s", textFile)
	f, err := os.Create(textFile)
	if err != nil {
		common.Error(err)
		runCmd(true, "rm", "-r", repoDir)
		return false
	}
	defer f.Close()

	common.Debugf("Writing %d characters to %s", textFileSizeBytes, textFile)
	_, err = f.WriteString(common.TextStringWithNewlines(textFileSizeBytes))
	if err != nil {
		common.Error(err)
		runCmd(true, "rm", "-r", repoDir)
		return false
	}
	f.Sync()

	// Copy text file into repo
	if !runCmd(true, "cp", textFile, repoDir) {
		runCmd(true, "rm", "-r", repoDir, textFile)
		return false
	}

	// git add
	if !gitRepo.runGitCmd(true, "-C", repoDir, "add", ".") {
		runCmd(true, "rm", "-r", repoDir, textFile)
		return false
	}

	// git commit
	if !gitRepo.runGitCmd(true, "-C", repoDir, "commit", "-m", "Test commit") {
		runCmd(true, "rm", "-r", repoDir, textFile)
		return false
	}

	// git push
	if !gitRepo.runGitCmd(true, "-C", repoDir, "push") {
		runCmd(true, "rm", "-r", repoDir, textFile)
		return false
	}

	// Remove the cloned directory
	if !runCmd(true, "rm", "-r", repoDir) {
		runCmd(true, "rm", textFile)
		return false
	}

	// Clone into new directory
	repoDir = fmt.Sprintf("/tmp/%s-take2", gitRepo.RepoName)
	if !gitRepo.runGitCmd(true, "clone", gitRepo.Url(), repoDir) {
		runCmd(true, "rm", textFile)
		return false
	}

	// Compare the existing text file to the one in the repo
	repoFilePath := fmt.Sprintf("%s/%s", repoDir, textFileBase)
	if !runCmd(true, "cmp", textFile, repoFilePath) {
		runCmd(true, "rm", "-r", repoDir, textFile)
		return false
	}

	// Remove the repo dir
	if !runCmd(true, "rm", "-r", repoDir, textFile) {
		return false
	}
	return true
}

func (gitRepo *GitRepo) cloneShouldFail() bool {
	var repoDir string

	repoDir = fmt.Sprintf("/tmp/%s-shouldfail", gitRepo.RepoName)
	if gitRepo.runGitCmd(false, "clone", gitRepo.Url(), repoDir) {
		return true
	}

	// Cleanup the repo dir, since the clone was unexpectedly successful
	runCmd(true, "rm", "-r", repoDir)
	return false
}
