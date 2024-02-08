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

# Prepare to build RPMs
if [[ -d ${RPM_BUILD_DIR} ]]; then
    rm -rvf "${RPM_BUILD_DIR}"
fi

if [[ -e ${RPM_BUILD_DIR} ]]; then
    echo "ERROR: '${RPM_BUILD_DIR}' exists, but should not"
    exit 1
fi

mkdir -pv "${RPM_BUILD_DIR}/SPECS" "${RPM_BUILD_DIR}/SOURCES"
cp -v "${RPM_SPEC_FILE}" "${RPM_BUILD_DIR}/SPECS/"

# Package source

# This find commands only prints files and empty directories.
# Non-empty directories will be included automatically if we include
# any of their contents.
find . \
    -name .git\* -prune -o \
    -type d \( \
        -name __pycache__ -o \
        -name cms_meta_tools -o \
        -path ./build -o \
        -path ./dist -o \
        -path ./cmsdev/vendor \
    \) -prune -o \
    -type f \( \
        -name \*.pyc -o \
        -name "${RPM_SOURCE_BASENAME}" \
    \) -prune -o \
    -type d -empty -print0 -o \
    -type f -print0 |
tar --null --transform "flags=r;s,^[.]/,/${RPM_SOURCE_NAME}/," \
    -cvjf "${RPM_SOURCE_PATH}" -T -

# build source rpm
rpmbuild -bs "${RPM_SPEC_FILE}" --target "${RPM_ARCH}" --define "_topdir ${RPM_BUILD_DIR}"

# build main rpm
rpmbuild -ba "${RPM_SPEC_FILE}" --target "${RPM_ARCH}" --define "_topdir ${RPM_BUILD_DIR}"
