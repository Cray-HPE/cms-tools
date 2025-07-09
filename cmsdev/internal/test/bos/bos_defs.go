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
package bos

import (
	"sync"

	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

/*
 * bos_defs.go
 *
 * BOS commons file
 *
 */

type CsmProductCatalogData struct {
	Configuration CsmProductCatalogConfiguration `json:"configuration"`
	Images        map[string]interface{}         `json:"images"`
	Recipes       map[string]interface{}         `json:"recipes"`
}

type CsmProductCatalogConfiguration struct {
	Clone_url     string `json:"clone_url"`
	Commit        string `json:"commit"`
	Import_branch string `json:"import_branch"`
	Import_Date   string `json:"import_date"`
	Ssh_url       string `json:"ssh_url"`
}

type ImageLink struct {
	S3_Etag string `json:"etag"`
	S3_Path string `json:"path"`
	Type    string `json:"type"`
}

type ImsImage struct {
	ImageID string    `json:"id"`
	Arch    string    `json:"arch"`
	Name    string    `json:"name"`
	Link    ImageLink `json:"link"`
}

type BootSet struct {
	Arch              string   `json:"arch"`
	Etag              string   `json:"etag"`
	Kernel_parameters string   `json:"kernel_parameters"`
	Node_roles_groups []string `json:"node_roles_groups"`
	Path              string   `json:"path"`
	Type              string   `json:"type"`
}

type BOSTemplateBootSet struct {
	Compute BootSet `json:"compute"`
}

type BOSCfs struct {
	Configuration string `json:"configuration"`
}

type BOSSessionTemplate struct {
	Enable_cfs bool               `json:"enable_cfs"`
	Name       string             `json:"name"`
	Tenant     string             `json:"tenant"`
	Boot_sets  BOSTemplateBootSet `json:"boot_sets"`
	Cfs        BOSCfs             `json:"cfs"`
}

type BOSSession struct {
	Components       string `json:"components"`
	Include_disabled bool   `json:"include_disabled"`
	Limit            string `json:"limit"`
	Name             string `json:"name"`
	Operation        string `json:"operation"`
	Stage            bool   `json:"stage"`
	Template_name    string `json:"template_name"`
	Tenant           string `json:"tenant"`
}

type BOSSessionTemplateInventory struct {
	TemplateNameList []string `json:"template_name_list"`
}

var (
	instance *BOSSessionTemplateInventory
	once     sync.Once
)

// GetBOSSessionTemplateInventoryInstance returns the singleton instance of BOSSessionTemplateInventory
func GetBOSSessionTemplateInventoryInstance() *BOSSessionTemplateInventory {
	once.Do(func() {
		instance = &BOSSessionTemplateInventory{
			TemplateNameList: []string{},
		}
	})
	return instance
}

// CMS service endpoints
var endpoints map[string]map[string]*common.Endpoint = common.GetEndpoints()
