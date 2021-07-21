# Copyright 2021 Hewlett Packard Enterprise Development LP
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
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
# (MIT License)

NAME ?= cray-cmstools
VERSION ?= $(shell cat .version)-local

BUILD_METADATA ?= "1~development~$(shell git rev-parse --short HEAD)"
BUILD_DIR ?= $(PWD)/dist/rpmbuild

CMSDEV_SPEC_NAME ?= ${NAME}-crayctldeploy
CMSDEV_SPEC_FILE ?= ${CMSDEV_SPEC_NAME}.spec
CMSDEV_SOURCE_NAME ?= ${CMSDEV_SPEC_NAME}-${VERSION}
CMSDEV_SOURCE_PATH := ${BUILD_DIR}/SOURCES/${CMSDEV_SOURCE_NAME}.tar.bz2

TESTS_SPEC_NAME ?= ${NAME}-crayctldeploy-test
TESTS_SPEC_FILE ?= ${TESTS_SPEC_NAME}.spec
TESTS_SOURCE_NAME ?= ${TESTS_SPEC_NAME}-${VERSION}
TESTS_SOURCE_PATH := ${BUILD_DIR}/SOURCES/${TESTS_SOURCE_NAME}.tar.bz2

all : build_prep lint prepare rpm rpm_test
rpm: rpm_package_source rpm_build_source rpm_build
rpm_test: rpm_package_test_source rpm_build_test_source rpm_build_test

build_prep:
		./runBuildPrep.sh

lint:
		./runLint.sh

prepare:
		rm -rf $(BUILD_DIR)
		mkdir -p $(BUILD_DIR)/SPECS $(BUILD_DIR)/SOURCES
		cp $(CMSDEV_SPEC_FILE) $(BUILD_DIR)/SPECS/
		cp $(TESTS_SPEC_FILE) $(BUILD_DIR)/SPECS/

rpm_package_source:
		tar --transform 'flags=r;s,^,/$(CMSDEV_SOURCE_NAME)/,' --exclude .git --exclude dist --exclude ct-tests --exclude $(TESTS_SPEC_FILE) -cvjf $(CMSDEV_SOURCE_PATH) .

rpm_build_source:
		BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ts $(CMSDEV_SOURCE_PATH) --define "_topdir $(BUILD_DIR)"

rpm_build:
		BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ba $(CMSDEV_SPEC_FILE) --nodeps --define "_topdir $(BUILD_DIR)"

rpm_package_test_source:
		tar --transform 'flags=r;s,^,/$(TESTS_SOURCE_NAME)/,' --exclude .git --exclude dist --exclude cmsdev --exclude cmslogs --exclude cms-tftp --exclude $(CMSDEV_SPEC_FILE) -cvjf $(TESTS_SOURCE_PATH) .

rpm_build_test_source:
		BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ts $(TESTS_SOURCE_PATH) --define "_topdir $(BUILD_DIR)"

rpm_build_test:
		BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ba $(TESTS_SPEC_FILE) --nodeps --define "_topdir $(BUILD_DIR)"
