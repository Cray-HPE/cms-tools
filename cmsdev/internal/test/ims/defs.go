package ims

/*
 * defs.go
 *
 * ims commons file
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
 */

import (
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
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

type IMSRecipeRecord struct {
	Id, Created, Recipe_type, Linux_distribution, Name string
	Link                                               map[string]string
}

type IMSVersionRecord struct {
	Version string
}

var pvcNames = []string{
	"cray-ims-data-claim",
}

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()
