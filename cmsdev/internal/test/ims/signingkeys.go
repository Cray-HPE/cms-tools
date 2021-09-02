package ims

/*
 * signingkeys.go
 *
 * ims signing keys test file
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
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

// Run the command kubectl -n services get cm cray-configmap-ims-v2-image-create-kiwi-ng -o go-template='{{index .data "image_job_create.yaml.template"}}'
// Get the image with kiwi-ng in the image name and return
func getOpensuseImage() (imagename string, err error) {
	imagename = ""
	var cmd *exec.Cmd
	var cmdOut []byte
	var kubectlPath string

	kubectlPath, err = k8s.GetKubectlPath()
	if err != nil {
		common.Errorf("Error getting kubectl path")
		common.Errorf("Error %s", err)
		return
	}
	// kubectl -n services get cm cray-configmap-ims-v2-image-create-kiwi-ng -o go-template='{{index .data "image_job_create.yaml.template"}}'
	cmd = exec.Command(kubectlPath, "-n", common.NAMESPACE, "get", "cm", "cray-configmap-ims-v2-image-create-kiwi-ng", "-o", "go-template='{{index .data \"image_job_create.yaml.template\"}}'")
	common.Infof("Running command: %s", cmd)
	cmdOut, err = cmd.CombinedOutput()
	common.Infof("OUT: %s", cmdOut)
	if err != nil || len(cmdOut) == 0 {
		common.Errorf("Error getting Open SuSE Image from cray-configmap-ims-v2-image-create-kiwi-ng")
		common.Errorf("Error %s", err)
		return
	}
	// Structure to parse yaml from kubectl and find images
	type ImageName struct {
		Spec struct {
			Template struct {
				Spec struct {
					InitContainers []struct {
						Image string `yaml:"image"`
					} `yaml:"initContainers"`
				} `yaml:"spec"`
			} `yaml:"template"`
		} `yaml:"spec"`
	}
	var in ImageName
	err = yaml.Unmarshal(cmdOut, &in)
	common.Infof("SPEC: %s", in)
	// Find image with name containing kiwi-ng
	for _, cont := range in.Spec.Template.Spec.InitContainers {
		common.Infof("NAME: %s", cont.Image)
		if strings.Contains(cont.Image, "kiwi-ng") {
			imagename = cont.Image
		}
	}
	return
}

// Run the command podman run --entrypoint "" registry.local/imagename ls /signing-keys
// return the signing keys found
func getSigningkeys(imagename string) (keys []string, err error) {
	var cmd *exec.Cmd
	var cmdOut []byte
	// Run the command podman run --entrypoint "" registry.local/imagename ls /signing-keys
	cmd = exec.Command("podman", "run", "--entrypoint", "\"\"", "registry.local/"+imagename, "ls", "/signing-keys")
	//cmd = exec.Command("docker", "run", "--entrypoint=", "--rm", "artifactory.algol60.net/"+imagename, "ls", "/signing-keys")
	common.Infof("Running command: %s", cmd)
	cmdOut, err = cmd.CombinedOutput()
	common.Infof("OUT: %s", cmdOut)
	if err != nil {
		common.Errorf("Error running podman ls /signing-keys command")
		common.Errorf("Error %s", err)
		return
	}
	// Find signing keys by finding strings that end with .asc
	re := regexp.MustCompile(".*.asc")
	keys = re.FindAllString(string(cmdOut), -1)
	if len(keys) == 0 {
		common.Errorf("Error podman ls /signing-keys command returned no keys")
		common.Errorf("Error %s", err)
		return
	}
	for _, key := range keys {
		common.Infof("Found Key: %s", key)
	}
	return
}

// Run the command podman run --entrypoint "" registry.local/imagename cat /scripts/entrypoint.sh
// Verify that keys are present in the script
func verifySigningkeys(imagename string, keys []string) (foundall bool, err error) {
	var cmd *exec.Cmd
	var cmdOut []byte
	foundall = false
	// Run the command podman run --entrypoint "" registry.local/imagename cat /scripts/entrypoint.sh
	cmd = exec.Command("podman", "run", "--entrypoint", "\"\"", "registry.local/"+imagename, "cat", "/scripts/entrypoint.sh")
	//cmd = exec.Command("docker", "run", "--entrypoint=", "--rm", "artifactory.algol60.net/"+imagename, "cat", "/scripts/entrypoint.sh")
	common.Infof("Running command: %s", cmd)
	cmdOut, err = cmd.CombinedOutput()
	common.Infof("OUT: %s", cmdOut)
	if err != nil || len(cmdOut) == 0 {
		common.Errorf("Error running podman cat /scripts/entrypoint.sh command")
		common.Errorf("Error %s", err)
		return
	}
	keysfound := 0
	// Look for lines with --signing-key /signing-keys/key
	for _, key := range keys {
		re := regexp.MustCompile(".*--signing-key /signing-keys/" + key)
		foundkey := re.FindAllString(string(cmdOut), -1)
		if len(foundkey) > 0 {
			common.Infof("Found Key: %s %s", key, foundkey[0])
			keysfound++
		}
	}
	if keysfound == len(keys) {
		foundall = true
	}
	return
}

// Run the signingkeys test
func signingkeysTest() bool {
	// Get opensuse image
	imagename, err := getOpensuseImage()
	if err != nil {
		return false
	}
	common.Infof(imagename)
	if len(imagename) == 0 {
		return false
	}
	// Get signingkeys
	keys, err := getSigningkeys(imagename)
	if err != nil {
		return false
	}
	if len(keys) == 0 {
		return false
	}
	// verify signingkeys
	found, err := verifySigningkeys(imagename, keys)
	if err != nil {
		return false
	}
	return found
}
