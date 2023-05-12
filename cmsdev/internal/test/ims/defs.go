// MIT License
//
// (C) Copyright 2021-2023 Hewlett Packard Enterprise Development LP
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
 * defs.go
 *
 * ims commons file
 *
 */

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

const RECIPE_DISTRO_DEFAULT string = "sles15"

type Recipe struct {
	Name, Distro string
}

type IMSImageRecord struct {
	Created, Id, Name string
	Link              map[string]string
}

type IMSConnectionInfoRecord struct {
	Host string
	Port int
}

type IMSSSHContainerRecord struct {
	Connection_info map[string]IMSConnectionInfoRecord
	Jail            bool
	Name, Status    string
}

type IMSJobRecord struct {
	Artifact_id, Created, Id, Image_root_archive_name, Initrd_file_name,
	Job_type, Kernel_file_name, Kubernetes_configmap, Kubernetes_job,
	Kubernetes_namespace, Kubernetes_service, Public_key_id,
	Resultant_image_id, Status string
	Build_env_size int
	Enable_debug   bool
	Ssh_containers []IMSSSHContainerRecord
}

type IMSPublicKeyRecord struct {
	Created, Id, Name, Public_key string
}

type IMSRecipeKeyValuePair struct {
	Key, Value string
}

type IMSRecipeRecord struct {
	Id, Created, Recipe_type, Linux_distribution, Name, Arch string
	Link                                                     map[string]string
	Require_dkms                                             bool
	Template_dictionary                                      []IMSRecipeKeyValuePair
}

type IMSVersionRecord struct {
	Version string
}

var pvcNames = []string{
	"cray-ims-data-claim",
}

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()
