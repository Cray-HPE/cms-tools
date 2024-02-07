#
# MIT License
#
# (C) Copyright 2024 Hewlett Packard Enterprise Development LP
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
# The format is just <variable_name>=<variable_value>
# It needs to work if sourced from a shell script
export NAME=cray-cmstools-crayctldeploy
export RPM_SPEC_FILE=${NAME}.spec
export LOCAL_VENV_PYTHON_SUBDIR_NAME=venv-python
export INSTALL_VENV_BASE_DIR=/usr/lib/cray-cmstools-crayctldeploy
export INSTALL_VENV_PYTHON_BASE_DIR=${INSTALL_VENV_BASE_DIR}/python
export BBIT_VENV_NAME=barebones_image_test-venv
if [[ -v PY_VERSION && -n ${PY_VERSION} ]]; then
    export BBIT_INSTALL_VENV_DIR=${INSTALL_VENV_PYTHON_BASE_DIR}/${PY_VERSION}/${BBIT_VENV_NAME}
    export BBIT_VENV_PYTHON_BIN=${BBIT_INSTALL_VENV_DIR}/bin/python${PY_VERSION}
fi
if [[ -v RPM_VERSION && -n ${RPM_VERSION} ]]; then
    export RPM_SOURCE_NAME=${NAME}-${RPM_VERSION}
	export RPM_SOURCE_BASENAME=${RPM_SOURCE_NAME}.tar.bz2
    if [[ -v RPM_BUILD_DIR && -n ${RPM_BUILD_DIR} ]]; then
        export RPM_SOURCE_PATH=${RPM_BUILD_DIR}/SOURCES/${RPM_SOURCE_BASENAME}
    fi
fi
