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

/*
 * cfs_defs.go
 *
 * cfs definitions
 *
 */
package cfs

import "stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"

var DEFAULT_PLAYBOOK = "compute_nodes.yml"

type CfsLayer struct {
	Clone_url string `json:"clone_url,omitempty"`
	CloneURL  string `json:"cloneUrl,omitempty"`
	Commit    string `json:"commit"`
	Name      string `json:"name"`
	Playbook  string `json:"playbook"`
}

type CFSConfiguration struct {
	Layers []CfsLayer `json:"layers"`
	Name   string     `json:"name"`
}

type CFSConfigurationsList struct {
	Configurations []CFSConfiguration `json:"configurations"`
}

type CsmProductCatalogConfiguration struct {
	Clone_url     string `json:"clone_url"`
	Commit        string `json:"commit"`
	Import_branch string `json:"import_branch"`
	Import_Date   string `json:"import_date"`
	Ssh_url       string `json:"ssh_url"`
}

type CFSSourceCredentials struct {
	Authentication_method string `json:"authentication_method"`
	Secret_name           string `json:"secret_name"`
}

type CFSSources struct {
	Clone_url   string               `json:"clone_url"`
	Name        string               `json:"name"`
	Credentials CFSSourceCredentials `json:"credentials"`
}

type CFSSourcesList struct {
	Sources []CFSSources `json:"sources"`
}

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()
