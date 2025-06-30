// MIT License
//
// (C) Copyright 2021-2023, 2025 Hewlett Packard Enterprise Development LP
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

/*
 * signingkeys.go
 *
 * IMS signing keys test file
 *
 */

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/k8s"
)

const configMapName = "cray-configmap-ims-v2-image-create-kiwi-ng"
const imageFieldName = "image_job_create.yaml.template"

// Retrieve the cray-configmap-ims-v2-image-create-kiwi-ng configmap in kubernetes
// Look at the data field for "image_job_create.yaml.template"
// Parse that field to find the image which has kiwi-ng in its name, and return that name
func getOpensuseImage() (imagename string, err error) {
	imagename = ""

	// Retrieve image_job_create.yaml.template data field from the config map
	imageFieldBytes, err := k8s.GetConfigMapDataField(common.NAMESPACE, configMapName, imageFieldName)
	if err != nil {
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
	err = yaml.Unmarshal(imageFieldBytes, &in)
	if err != nil {
		common.Errorf("Error parsing config map data as YAML")
		return
	}
	common.Debugf("SPEC: %s", in)
	common.Debugf("Find image with name containing kiwi-ng")
	for _, cont := range in.Spec.Template.Spec.InitContainers {
		common.Debugf("NAME: %s", cont.Image)
		if strings.Contains(cont.Image, "kiwi-ng") {
			imagename = cont.Image
			return
		}
	}
	err = fmt.Errorf("No image found with name containing kiwi-ng")
	return
}

// Run the command podman run --entrypoint "" registry.local/imagename ls /signing-keys
// return the signing keys found
func getSigningkeys(imagename string) (keys []string, err error) {
	var cmd *exec.Cmd
	var cmdOut []byte
	// Run the command podman run --entrypoint "" registry.local/imagename ls /signing-keys
	cmd = exec.Command("podman", "run", "--entrypoint", "", "registry.local/"+imagename, "ls", "/signing-keys")
	common.Infof("Running command: %s", cmd)
	cmdOut, err = cmd.CombinedOutput()
	common.Debugf("OUT: %s", cmdOut)
	if err != nil {
		common.Errorf("Error running podman ls /signing-keys command")
		return
	}
	// Find signing keys by finding strings that end with .asc
	re := regexp.MustCompile(".*.asc")
	keys = re.FindAllString(string(cmdOut), -1)
	if len(keys) == 0 {
		err = fmt.Errorf("podman ls /signing-keys command returned no keys")
		return
	}
	for _, key := range keys {
		common.Debugf("Found Key: %s", key)
	}
	return
}

// Run the signing keys test
func signingkeysTest() bool {
	common.Infof("Performing RPM signing keys test")

	// Get opensuse image
	imagename, err := getOpensuseImage()
	if err != nil {
		common.Error(err)
		return false
	}

	// Get signing keys
	keys, err := getSigningkeys(imagename)
	if err != nil {
		common.Error(err)
		return false
	}

	// Verify there are signing keys baked in the image
	if len(keys) == 0 {
		common.Errorf("No signing keys found in image %s", imagename)
		return false
	}

	common.Infof("RPM signing keys test passed")
	return true
}
