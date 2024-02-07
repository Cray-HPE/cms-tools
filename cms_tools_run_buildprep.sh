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

set -exuo pipefail

source ./vars.sh
sed -i "s#@BB_BASE_DIR@#${INSTALL_VENV_PYTHON_BASE_DIR}#" run_barebones_image_test.sh
[[ -n ${LOCAL_VENV_PYTHON_SUBDIR_NAME} ]]
if [[ -d ./${LOCAL_VENV_PYTHON_SUBDIR_NAME} ]]; then
    rm -rvf "./${LOCAL_VENV_PYTHON_SUBDIR_NAME}"
fi
[[ ! -e ${LOCAL_VENV_PYTHON_SUBDIR_NAME} ]]
mkdir -pv "${LOCAL_VENV_PYTHON_SUBDIR_NAME}"
./cms_meta_tools/scripts/runBuildPrep.sh

# If the `build` directory exists, delete it
if [[ -e ./build ]]; then
    [[ -d ./build ]]
    rm -rvf ./build
fi
