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

# Run the barebones boot test from the appropriate virtual environment
# based on the installed system Python version

# The value for this variable is set by the Makefile during the build
BB_BASE_DIR=@BB_BASE_DIR@

function err_exit
{
    echo "ERROR: $*" 1>&2
    exit 1
}

[[ -n ${BB_BASE_DIR} ]] || err_exit "Build-time error with cmstools RPM"
[[ -e ${BB_BASE_DIR} ]] || err_exit "Directory '${BB_BASE_DIR}' should exist but it does not"
[[ -d ${BB_BASE_DIR} ]] || err_exit "'${BB_BASE_DIR}' exists; it should be a directory, but it is not"

PYTHON3_BASE_VERSION=$(rpm -q --queryformat '%{VERSION}' python3-base | cut -d. -f1-2) || PYTHON3_BASE_VERSION=""

for PYVER in $(ls "${BB_BASE_DIR}" | grep -E '^[1-9][0-9]*[.](0|[1-9][0-9]*)$' | sort -t. -n -k1,1 -k2,2 -r); do
    if [[ ${PYVER} == "${PYTHON3_BASE_VERSION}" ]]; then
        "${BB_BASE_DIR}/${PYVER}/cmstools-venv/bin/barebones_image_test" "$@"
        exit $?
    # Alternatively, see if pythonxxx-base is installed
    elif rpm -q "python${PYVER//.}-base" >/dev/null 2>&1; then
        "${BB_BASE_DIR}/${PYVER}/cmstools-venv/bin/barebones_image_test" "$@"
        exit $?
    fi
done

err_exit "No installed Python version found matching installed barebones boot test"
