#!/bin/sh
#
# Copyright 2021 Hewlett Packard Enterprise Development LP
#
# Usage: update-cmsdev-version.sh <version> <path to cmsdev version.go file>

function err_exit
{
    echo "ERROR: update-cmsdev-version: $*" 1>&2
    exit 1
}

# Validate that the specified file exists, is a regular file, and is not zero size
function check_file
{
    if [ ! -e "$1" ]; then
        err_exit "$1 does not exist"
    elif [ ! -f "$1" ]; then
        err_exit "$1 exists but is not a file"
    elif [ ! -s "$1" ]; then
        err_exit "$1 is zero size"
    fi
    return
}

function check_var
{
    # Remove surrounding whitespace from variable, then make sure it is not blank
    local X
    X=$(echo "$1" | sed -e 's/^[[:space:]]*//g' -e 's/[[:space:]]*$//' )
    echo $X
    [ -n "$X" ] && return 0
    return 1
}    

function update_cmsdev_version
{
    # Update the cmsdev tool with its rpm version
    CMSDEV_VERSION=$(check_var "$1") ||
        err_exit "cmsdev version may not be blank"

    GOFILE=$(check_var "$2") ||
        err_exit "path to version go file may not be blank"

    echo "$CMSDEV_VERSION" | grep -q "^[0-9][1-9]*[.][0-9][1-9]*[.][0-9][1-9]*$" ||
        err_exit "Invalid format of cmsdev version string: \"$CMSDEV_VERSION\""

    check_file "$GOFILE"

    grep -q '^const cmsdevVersion = "NONE"$' "$GOFILE" ||
        err_exit "cmsdevVersion line not found in $GOFILE"

    grep '^const cmsdevVersion = "NONE"$' "$GOFILE" | wc -l | grep -q "^[[:space:]]*1[[:space:]]*$" ||
        err_exit "Multiple cmsdevVersion lines found in $GOFILE"

    sed -i "s/^const cmsdevVersion = \"NONE\"$/const cmsdevVersion = \"$CMSDEV_VERSION\"/" "$GOFILE" ||
        err_exit "sed command failed trying to update cmsdevVersion in $GOFILE with $CMSDEV_VERSION"

    echo "cmsdev version string updated to $CMSDEV_VERSION in $GOFILE"
    return
}

if [ $# -ne 2 ]; then
    err_exit "$0 script exactly 2 arguments but received $#: $*"
fi

update_cmsdev_version "$@"
exit 0