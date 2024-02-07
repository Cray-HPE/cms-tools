#!/usr/bin/bash
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

function check_build {
    if [[ -d build ]]; fi
        find build -type d -print
    else
        echo "no build"
    fi
}

set -exuo pipefail

check_build

source ./vars.sh

mkdir -pv "${BBIT_INSTALL_VENV_DIR}"

check_build

# Create our virtualenv
"${PYTHON_BIN}" -m venv ${BBIT_INSTALL_VENV_DIR}

check_build

which "${PYTHON_BIN}"

# Activate virtual env
source ${BBIT_INSTALL_VENV_BIN_DIR}/activate

check_build

which "${PYTHON_BIN}"

# For the purposes of the build log, we list the installed Python packages before and after each pip call
"${PYTHON_BIN}" -m pip list --format freeze --disable-pip-version-check

check_build

# Upgrade install/build tools
"${PYTHON_BIN}" -m pip install pip setuptools wheel -c barebones_image_test-constraints.txt --disable-pip-version-check --no-cache
"${PYTHON_BIN}" -m pip list --format freeze --disable-pip-version-check

check_build

# Install test preqrequisites
"${PYTHON_BIN}" -m pip install -r barebones_image_test-requirements.txt --disable-pip-version-check --no-cache
"${PYTHON_BIN}" -m pip list --format freeze --disable-pip-version-check

# Install the test itself
"${PYTHON_BIN}" -m pip install . -c barebones_image_test-constraints.txt --disable-pip-version-check --no-cache
"${PYTHON_BIN}" -m pip list --format freeze --disable-pip-version-check

# Remove build tools to decrease the virtualenv size.
"${PYTHON_BIN}" -m pip uninstall -y pip setuptools wheel
# Cannot list packages a final time, since we uninstalled pip
