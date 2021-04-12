#!/bin/bash
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

# Version must be of the form #.#.# where unnecessary leading 0s are not allowed
VPATTERN="^(0|[1-9][0-9]*)[.](0|[1-9][0-9]*)[.](0|[1-9][0-9]*)$"

VERSION_TAG="@VERSION@"
VERSION_FILE=".version"
FILES_WITH_VERSION_TAGS=(
    "cray-cmstools-crayctldeploy.spec"
    "cray-cmstools-crayctldeploy-test.spec"
    "cmsdev/internal/cmd/version.go"
)

function error_exit
{
    echo "ERROR: $*"
    exit 1
}

function run_cmd
{
    "$@" || error_exit "Command failed with rc $?: $*"
    return 0
}

function process_file
{
    # $1 - file
    # $2 - tag
    # $3 - string
    [ $# -eq 3 ] || 
        error_exit "PROGRAMMING LOGIC ERROR: process_file should get exactly 3 arguments but it received $#: $*"
    F="$1"
    VTAG="$2"
    VSTRING="$3"
    grep -q "$VTAG" "$F" ||
        error_exit "Version tag ($VTAG) not found in file $F"
    BEFORE="${F}.before"
    run_cmd cp "$F" "$BEFORE"
    run_cmd sed -i s/${VTAG}/${VSTRING}/g "$F"
    echo "# diff \"$BEFORE\" \"$F\""
    diff "$BEFORE" "$F"
    rc=$?
    if [ $rc -eq 0 ]; then
        error_exit "diff reports no difference after tag replacement"
    elif [ $rc -ne 1 ]; then
        error_exit "diff encountered an error comparing the files"
    fi
    run_cmd rm "$BEFORE"
}

# Get CRUS version string from $VERSION_FILE
VERSION_STRING=$(cat "$VERSION_FILE") ||
    error_exit "Failed: cat $VERSION_FILE"
echo "Version string from $VERSION_FILE is \"$VERSION_STRING\""
[[ $VERSION_STRING =~ $VPATTERN ]] ||
    error_exit "Version string does not match expected format"

# Update files
for F in "${FILES_WITH_VERSION_TAGS[@]}" ; do
    process_file "$F" "$VERSION_TAG" "$VERSION_STRING"
done

exit 0
