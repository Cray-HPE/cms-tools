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

# Arguments <target directory> <rpm1> [<rpm2>] ...

MYNAME=$0
ALL_ARGS="$*"

function usage
{
    echo "ERROR: ${MYNAME}: $*. Invalid arguments: ${ALL_ARGS}"
    exit 1
}

set -euo pipefail

# First quick sanity checks
[[ $# -ge 2 ]] || usage "Script requires at least 2 arguments but received $#."
TARGET_DIR=$1
[[ -n ${TARGET_DIR} ]] || usage "Target directory cannot be blank."
[[ -e ${TARGET_DIR} ]] || usage "Target directory (${TARGET_DIR}) does not exist."
[[ -d ${TARGET_DIR} ]] || usage "Target (${TARGET_DIR}) exists but is not a directory."
shift

TEMPFILE=$(mktemp)

while [[ $# -gt 0 ]]; do
    RPM=$1
    [[ -n ${RPM} ]] || usage "RPM argument may not be blank"
    [[ -e ${RPM} ]] || usage "RPM (${RPM}) does not exist."
    [[ -f ${RPM} ]] || usage "RPM (${RPM}) exists but is not a regular file."
    echo "$0: Extracting RPM '${RPM}' into target directory '${TARGET_DIR}'"
    rpm2cpio "${RPM}" | cpio -idmv --quiet -D "${TARGET_DIR}" 2>&1 | tee -a "${TEMPFILE}"
    echo "$0: Done extracting RPM '${RPM}' into target directory '${TARGET_DIR}'; updating INSTALLED_FILES list"
    grep -Ev "^cpio: " "${TEMPFILE}" | sed 's/^[.]//' | tee -a INSTALLED_FILES
    echo "$0: INSTALLED_FILES list updated"
    shift
done

# Don't fail the script if we happen to hit an error deleting the temporary file
rm "${TEMPFILE}" || true

echo "$0: All done!"
