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

NAME ?= cray-cmstools-crayctldeploy
SHELL=/bin/bash
RPM_VERSION ?= $(shell head -1 .version)
RPM_RELEASE ?= $(shell head -1 .rpm_release)
BUILD_METADATA ?= "1~development~$(shell git rev-parse --short HEAD)"
PY_VERSION ?= "3.10"
PYTHON_BIN := python$(PY_VERSION)

LOCAL_VENV_DIR ?= $(PWD)/venv

BBIT_BASE_INSTALL_DIR ?= /usr/lib/$(NAME)
BBIT_VENV_INSTALL_DIR ?= /usr/lib/$(NAME)/python/$(PY_VERSION)/barebones_image_test-venv
BBIT_VENV_PYTHON_BIN ?= $(BBIT_VENV_INSTALL_DIR)/bin/python$(PY_VERSION)

BUILD_DIR ?= $(PWD)/dist/rpmbuild

PYVER_NODOTS := $(shell echo ${PY_VERSION//.})
PYTHON_MAJOR := $(shell echo ${PY_VERSION} | cut -d. -f1)
PYTHON_MINOR := $(shell echo ${PY_VERSION} | cut -d. -f2)
NEXT_PY_VERSION := $(PYTHON_MAJOR).$(shell expr $(PYTHON_MINOR) + 1)


BBIT_NAME := $(NAME)-python$(PYVER_NODOTS)
BBIT_SPEC_FILE ?= $(BBIT_NAME).spec
BBIT_SOURCE_NAME ?= $(BBIT_NAME)-$(RPM_VERSION)
BBIT_SOURCE_BASENAME := $(BBIT_SOURCE_NAME).tar.bz2
BBIT_SOURCE_PATH := $(BUILD_DIR)/SOURCES/$(BBIT_SOURCE_BASENAME)

CMSDEV_LOGDIR := $(shell ./cmsdev_logdir.sh)
BBIT_LOGDIR := $(shell ./barebones_image_test_logdir.sh)


all : runbuildprep lint build_cmsdev prepare rpm
rpm: rpm_package_source rpm_build_source rpm_build

runbuildprep:
		sed -i 's#@BB_BASE_DIR@#$(BASE_INSTALL_DIR)#' run_barebones_image_test.sh
		mkdir -pv $(LOCAL_VENV_DIR)
		./cms_meta_tools/scripts/runBuildPrep.sh

lint:
		./cms_meta_tools/scripts/runLint.sh

build_cmsdev:
		# Record the go version in the build output, just in case it is helpful
		go version
		mkdir -p cmsdev/bin
		cd cmsdev && CGO_ENABLED=0 GO111MODULE=on GOARCH=amd64 GOOS=linux go build -o ./bin/cmsdev -mod vendor .

build_python_venv:
		mkdir -pv $(BBIT_VENV_INSTALL_DIR)
		# Create our virtualenv
		$(PYTHON_BIN) -m venv $(BBIT_VENV_INSTALL_DIR)
		# For the purposes of the build log, we list the installed Python packages before and after each pip call
		$(BBIT_VENV_PYTHON_BIN) -m pip list --format freeze --disable-pip-version-check
		# Upgrade install/build tools
		$(BBIT_VENV_PYTHON_BIN) -m pip install pip setuptools wheel -c barebones_image_test-constraints.txt --disable-pip-version-check --no-cache
		$(BBIT_VENV_PYTHON_BIN) -m pip list --format freeze --disable-pip-version-check
		# Install test preqrequisites
		$(BBIT_VENV_PYTHON_BIN) -m pip install -r barebones_image_test-requirements.txt --disable-pip-version-check --no-cache
		$(BBIT_VENV_PYTHON_BIN) -m pip list --format freeze --disable-pip-version-check
		# Install the test itself
		$(BBIT_VENV_PYTHON_BIN) -m pip install . -c barebones_image_test-constraints.txt --disable-pip-version-check --no-cache
		$(BBIT_VENV_PYTHON_BIN) -m pip list --format freeze --disable-pip-version-check
		# Remove build tools to decrease the virtualenv size.
		$(BBIT_VENV_PYTHON_BIN) -m pip uninstall -y pip setuptools wheel
		# Cannot list packages a final time, since we uninstalled pip

prepare:
		rm -rf $(BUILD_DIR)
		mkdir -p $(BUILD_DIR)/SPECS $(BUILD_DIR)/SOURCES
		cp $(CMSDEV_SPEC_FILE) $(BUILD_DIR)/SPECS/

rpm_package_source:
		touch $(CMSDEV_SOURCE_PATH)
		tar --transform 'flags=r;s,^,/$(CMSDEV_SOURCE_NAME)/,' \
			--exclude .git \
			--exclude ./cms_meta_tools \
			--exclude ./cmsdev/vendor \
			--exclude ./dist \
			--exclude $(CMSDEV_SOURCE_BASENAME) \
			-cvjf $(CMSDEV_SOURCE_PATH) .

rpm_build_source:
		BBIT_LOGDIR=$(BBIT_LOGDIR) \
		CMSDEV_LOGDIR=$(CMSDEV_LOGDIR) \
		CMSDEV_SOURCE_BASENAME=$(CMSDEV_SOURCE_BASENAME) \
		PYTHON_BIN=$(PYTHON_BIN) \
		NEXT_PY_VERSION=$(NEXT_PY_VERSION) \
		BUILD_METADATA=$(BUILD_METADATA) \
		BASE_INSTALL_DIR=$(BBIT_BASE_INSTALL_DIR) \
		rpmbuild -bs $(CMSDEV_SPEC_FILE) --target $(RPM_ARCH) --define "_topdir $(BUILD_DIR)"

rpm_build:
		BBIT_LOGDIR=$(BBIT_LOGDIR) \
		CMSDEV_LOGDIR=$(CMSDEV_LOGDIR) \
		CMSDEV_SOURCE_BASENAME=$(CMSDEV_SOURCE_BASENAME) \
		PYTHON_BIN=$(PYTHON_BIN) \
		NEXT_PY_VERSION=$(NEXT_PY_VERSION) \
		BUILD_METADATA=$(BUILD_METADATA) \
		BASE_INSTALL_DIR=$(BBIT_BASE_INSTALL_DIR) \
		rpmbuild -ba $(CMSDEV_SPEC_FILE) --target $(RPM_ARCH) --define "_topdir $(BUILD_DIR)"
