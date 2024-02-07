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

# Looks at the Python versions listed in the local virtual environment directory.
# Sets the Requires: field in the spec file accordingly.

source ./vars
if [[ -z ${LOCAL_VENV_PYTHON_SUBDIR_NAME} ]]; then
    echo "ERROR: $0: LOCAL_VENV_PYTHON_SUBDIR_NAME should be set" 1>&2
    exit 1
fi

REQUIRE_STRING=""

function add_requirement {
    [[ -z ${REQUIRE_STRING} ]] || REQUIRE_STRING+=" or "
    REQUIRE_STRING+="$*"
}

LAST_MAJOR=""
FIRST_MINOR=""
LAST_MINOR=""

for PY_VER in $(ls "./${LOCAL_VENV_PYTHON_SUBDIR_NAME}" | sort -t. -n -k1,1 -k2,2); do
    PY_VER_NODOTS=${PY_VER//.}
    add_requirement "python${PY_VER_NODOTS}-base"
    PY_VER_MAJOR=$(echo ${PY_VER} | cut -d. -f1)
    PY_VER_MINOR=$(echo ${PY_VER} | cut -d. -f2)
    if [[ -n ${LAST_MAJOR} ]]; then
        if [[ ${PY_VER_MAJOR} -eq ${LAST_MAJOR} && $((LAST_MINOR + 1)) -eq ${PY_VER_MINOR} ]]; then
            LAST_MINOR=${PY_VER_MINOR}
            continue
        fi
        add_requirement "(python${LAST_MAJOR}-base >= ${LAST_MAJOR}.${FIRST_MINOR} and python${LAST_MAJOR}-base < ${LAST_MAJOR}.$((LAST_MINOR + 1)))"
    fi
    LAST_MAJOR=${PY_VER_MAJOR}
    FIRST_MINOR=${PY_VER_MINOR}
    LAST_MINOR=${PY_VER_MINOR}
done
add_requirement "(python${LAST_MAJOR}-base >= ${LAST_MAJOR}.${FIRST_MINOR} and python${LAST_MAJOR}-base < ${LAST_MAJOR}.$((LAST_MINOR + 1)))"

echo "(${REQUIRE_STRING})"
exit 0
