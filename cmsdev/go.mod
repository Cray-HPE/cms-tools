// Copyright 2021 Hewlett Packard Enterprise Development LP
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
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//
// (MIT License)

module stash.us.cray.com/SCMS/cms-tools/cmsdev

go 1.13

require (
	github.com/fatih/color v1.12.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pin/tftp v2.1.0+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.14
	k8s.io/apimachinery v0.21.14
	k8s.io/client-go v0.21.14
)

// Pinned to kubernetes v0.21.14
replace (
	k8s.io/api => k8s.io/api v0.21.14
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.14
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.14
	k8s.io/apiserver => k8s.io/apiserver v0.21.14
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.14
	k8s.io/client-go => k8s.io/client-go v0.21.14
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.14
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.14
	k8s.io/code-generator => k8s.io/code-generator v0.21.14
	k8s.io/component-base => k8s.io/component-base v0.21.14
	k8s.io/cri-api => k8s.io/cri-api v0.21.14
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.14
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.14
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.14
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.14
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.14
	k8s.io/kubectl => k8s.io/kubectl v0.21.14
	k8s.io/kubelet => k8s.io/kubelet v0.21.14
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.14
	k8s.io/metrics => k8s.io/metrics v0.21.14
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.14
)
