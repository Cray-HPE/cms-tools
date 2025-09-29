#
# MIT License
#
# (C) Copyright 2024-2025 Hewlett Packard Enterprise Development LP
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

# This file is the "source of truth" for the repo
export NAME=cray-cmstools-crayctldeploy
export RPM_NAME=${NAME}
export GO_IMAGE='artifactory.algol60.net/csm-docker/stable/csm-docker-sle-go'
export PY_IMAGE='artifactory.algol60.net/csm-docker/stable/csm-docker-sle-python'
export RPM_ARCH='x86_64'
export RPM_OS='noos'
export RPM_SPEC_FILE=${RPM_NAME}.spec
export DESCRIPTION="Cray Management System: Tests and Tools"
export LOCAL_VENV_PYTHON_SUBDIR_NAME=venv-python
export LOCAL_VENV_PYTHON_BASE_DIR=$(pwd)/${LOCAL_VENV_PYTHON_SUBDIR_NAME}
export INSTALL_VENV_BASE_DIR=/usr/lib/cray-cmstools-crayctldeploy
export INSTALL_VENV_PYTHON_BASE_DIR=${INSTALL_VENV_BASE_DIR}/python
export CMSTOOLS_VENV_NAME=cmstools-venv
export RPM_BUILD_SUBDIR=dist/rpmbuild
export RPM_BUILD_DIR=$(pwd)/${RPM_BUILD_SUBDIR}

if [[ -v PY_VERSION && -n ${PY_VERSION} ]]; then
    export CMSTOOLS_INSTALL_VENV_DIR=${INSTALL_VENV_PYTHON_BASE_DIR}/${PY_VERSION}/${CMSTOOLS_VENV_NAME}
    export CMSTOOLS_INSTALL_VENV_BIN_DIR=${CMSTOOLS_INSTALL_VENV_DIR}/bin
    export PYTHON_BIN=python${PY_VERSION}
    export CMSTOOLS_VENV_PYTHON_BIN=${CMSTOOLS_INSTALL_VENV_BIN_DIR}/python${PYTHON_BIN}
fi

if [[ -f .version && -s .version ]]; then
    RPM_VERSION=$(head -1 .version)
    export RPM_VERSION
    export RPM_SOURCE_NAME=${RPM_NAME}-${RPM_VERSION}
    export RPM_SOURCE_BASENAME=${RPM_SOURCE_NAME}.tar.bz2
    if [[ -v RPM_BUILD_DIR && -n ${RPM_BUILD_DIR} ]]; then
        export RPM_SOURCE_PATH=${RPM_BUILD_DIR}/SOURCES/${RPM_SOURCE_BASENAME}
    fi
fi

if [[ -f .rpm_release && -s .rpm_release ]]; then
    RPM_RELEASE=$(head -1 .rpm_release)
    export RPM_RELEASE
fi

if [[ -f ./cmsdev/go.mod && -s ./cmsdev/go.mod ]]; then
    numpat='(0|[1-9][0-9]*)'
    GO_VERSION=$(grep -E "^go [1-9][0-9]*[.]${numpat}([[:space:]]*|[.]${numpat}[[:space:]]*)$" ./cmsdev/go.mod | awk '{ print $2 }' | cut -d. -f1-2)
    export GO_VERSION
fi

if [[ -d ./${LOCAL_VENV_PYTHON_SUBDIR_NAME} ]]; then
    PY_VERSIONS=$(ls "./${LOCAL_VENV_PYTHON_SUBDIR_NAME}" | sort -t. -n -k1,1 -k2,2 )
    export PY_VERSIONS
    NUM_PY_VERSIONS=$(echo "${PY_VERSIONS}" | wc -w)
    export NUM_PY_VERSIONS

    if [[ -f ./generate_rpm_python_requirements.sh ]]; then
        RPM_PYTHON_REQUIREMENTS=$(./generate_rpm_python_requirements.sh)
        export RPM_PYTHON_REQUIREMENTS
    fi
fi

if [[ -f ./cmsdev_logdir.sh ]]; then
    CMSDEV_LOGDIR=$(./cmsdev_logdir.sh)
    export CMSDEV_LOGDIR
fi

if [[ -f ./barebones_image_test_logdir.sh ]]; then
    BBIT_LOGDIR=$(./barebones_image_test_logdir.sh)
    export BBIT_LOGDIR
fi
