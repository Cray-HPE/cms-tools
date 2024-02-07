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
rm -rvf "${RPM_BUILD_DIR}"
mkdir -pv "${RPM_BUILD_DIR}/SPECS" "${RPM_BUILD_DIR}/SOURCES"

PYTHON_RPM_REQS=$(./generate_rpm_python_requirements.sh)
echo "${PYTHON_RPM_REQS}"

sed -i "s#@PYTHON_REQUIREMENTS@#${PYTHON_RPM_REQS}#" "${RPM_SPEC_FILE}"
cp -v "${RPM_SPEC_FILE}" "${RPM_BUILD_DIR}/SPECS/"

# Package source
touch "${RPM_SOURCE_PATH}"
tar --transform "flags=r;s,^,/${RPM_SOURCE_NAME}/," \
    --exclude .git \
    --exclude ./cms_meta_tools \
    --exclude ./cmsdev/vendor \
    --exclude ./dist \
    --exclude "${RPM_SOURCE_BASENAME}" \
    -cvjf "${RPM_SOURCE_PATH}" .

# build source rpm
rpmbuild -bs "${RPM_SPEC_FILE}" --target "${RPM_ARCH}" --define "_topdir ${RPM_BUILD_DIR}"

# build main rpm
rpmbuild -ba "${RPM_SPEC_FILE}" --target "${RPM_ARCH}" --define "_topdir ${RPM_BUILD_DIR}"