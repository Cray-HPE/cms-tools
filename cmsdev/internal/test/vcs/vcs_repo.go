package vcs

/*
 * vcs.go
 *
 * vcs repo test
 *
 * Copyright 2020 Hewlett Packard Enterprise Development LP
 */

import (
	"fmt"
	"github.com/go-resty/resty"
	"net/http"
	"os/exec"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

const VCSURL = common.BASEURL + "/vcs/api/v1"

func repoTest() (passed bool) {
	var resp *resty.Response

	passed = false

	// Get vcs user and password
	vcsUser, vcsPass, err := k8s.GetVcsUsernamePassword()
	if err != nil {
		common.Error(err)
		return
	}

	client := resty.New()
	client.SetHeader("Content-Type", "application/json")
	client.SetBasicAuth(vcsUser, vcsPass)

	// Attempt to create a repo
	repoName := "harf-" + common.AlnumString(8)
	common.Infof("Create vcs repo named %s", repoName)
	repoDataJsonString := fmt.Sprintf(
		`{ "name": "%s", "auto_init": null, "description": "Test repo named %s", "gitignores": null, "license": null, "private": false, "readme": null }`,
		repoName, repoName)
	repoDataArray := []byte(repoDataJsonString)
	createUrl := VCSURL + "/org/cray/repos"
	common.Infof("POST %s", createUrl)
	common.Infof("data: %s", repoDataJsonString)
	resp, err = client.R().
		SetBody(repoDataArray).
		Post(createUrl)
	if err != nil {
		common.Error(err)
		return
	}
	common.PrettyPrintJSON(resp)
	if resp.StatusCode() != http.StatusCreated {
		common.Errorf("POST %s: expected status code %d, got %d", createUrl, http.StatusCreated, resp.StatusCode())
		return
	}
	common.Infof("Repo created successfully")

	repoUrl := fmt.Sprintf("https://%s:%s@%s/vcs/cray/%s.git", vcsUser, vcsPass, common.BASEHOST, repoName)
	passed = true
	if !cloneTest(repoName, repoUrl) {
		passed = false
	}

	// Delete repo
	deleteUrl := VCSURL + "/repos/cray/" + repoName
	common.Infof("DELETE %s", deleteUrl)
	resp, err = client.R().
		Delete(deleteUrl)
	if err != nil {
		common.Error(err)
		passed = false
		return
	}
	common.PrettyPrintJSON(resp)
	if resp.StatusCode() != http.StatusNoContent {
		common.Errorf("DELETE %s: expected status code %d, got %d", createUrl, http.StatusNoContent, resp.StatusCode())
		passed = false
		return
	}
	common.Infof("Repo deleted successfully")

	// Try a clone to verify that it now fails
	if !cloneShouldFail(repoName, repoUrl) {
		passed = false
	}

	return
}

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

func cloneTest(repoName, repoUrl string) bool {
	var baseLsPath, repoDir, repoLsPath string
	var ok bool

	repoDir = fmt.Sprintf("/tmp/%s", repoName)

	if !runCmd(true, "git", "clone", repoUrl, repoDir) {
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
