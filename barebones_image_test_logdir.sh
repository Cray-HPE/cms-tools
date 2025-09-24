#!/usr/bin/bash
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

# Extracts the default log directory from the cmsdev source files and
# prints it. Prints an error to stderr and exits non-0 if there are
# problems.

SOURCEFILE="python-venv/cmstools/lib/common/log.py"

function err_exit
{
    [[ $# -ne 0 ]] && echo "ERROR: $*" >&2
    exit 1
}

function run_cmd
{
    local rc
    "$@"
    rc=$?
    [[ $rc -eq 0 ]] && return 0
    err_exit "Command failed with rc $rc: $*"
}

# First quick sanity checks
[[ -e "${SOURCEFILE}" ]] || err_exit "File does not exist: ${SOURCEFILE}"
[[ -f "${SOURCEFILE}" ]] || err_exit "Exists but is not a regular file: ${SOURCEFILE}"
[[ -s "${SOURCEFILE}" ]] || err_exit "File exists but is empty"

# Get a temporary file
TEMPFILE=$(run_cmd mktemp) || err_exit

# In theory the path could contain other characters, but it is not likely we would choose that.
run_cmd grep -E '^DEFAULT_LOG_DIR = ["]/[-./_a-zA-Z0-9]*["][[:space:]]*$' "${SOURCEFILE}" > "${TEMPFILE}" || err_exit

[[ -s "${TEMPFILE}" ]] || err_exit "Expected line not found in ${SOURCEFILE}"
[[ $(wc -l "${TEMPFILE}" | awk '{ print $1 }') == 1 ]] || err_exit "Multiple matching lines found in ${SOURCEFILE}"

run_cmd sed 's#^DEFAULT_LOG_DIR = ["]\(/[-./_a-zA-Z0-9]*\)["][[:space:]]*$#\1#' "${TEMPFILE}" || err_exit
rm "${TEMPFILE}" >/dev/null 2>&1
exit 0
