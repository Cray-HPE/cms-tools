#
# MIT License
#
# (C) Copyright 2021-2022, 2024 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
# If you wish to perform a local build, you will need to clone or copy the contents of the
# cms-meta-tools repo to ./cms_meta_tools

SHELL=/bin/bash
BUILD_METADATA ?= "1~development~$(shell git rev-parse --short HEAD)"
PY_VERSION ?= "3.10"

RPM_BUILD_DIR ?= $(PWD)/dist/rpmbuild

CMSDEV_LOGDIR := $(shell ./cmsdev_logdir.sh)
BBIT_LOGDIR := $(shell ./barebones_image_test_logdir.sh)

runbuildprep:
		./cms_tools_run_buildprep.sh

lint:
		./cms_meta_tools/scripts/runLint.sh

build_cmsdev:
		# Record the go version in the build output, just in case it is helpful
		go version
		mkdir -pv cmsdev/bin
		cd cmsdev && CGO_ENABLED=0 GO111MODULE=on GOARCH=amd64 GOOS=linux go build -o ./bin/cmsdev -mod vendor .

build_python_venv:
		PY_VERSION=$(PY_VERSION) ./build_python_venv.sh

rpm:
		RPM_BUILD_DIR=$(RPM_BUILD_DIR) \
		PY_VERSION=$(PY_VERSION) \
		BBIT_LOGDIR=$(BBIT_LOGDIR) \
		CMSDEV_LOGDIR=$(CMSDEV_LOGDIR) \
		BUILD_METADATA=$(BUILD_METADATA) \
		./build_rpm.sh
