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

REQUIRE_STRING=""

function add_requirement {
    [[ -z ${REQUIRE_STRING} ]] || REQUIRE_STRING+=" or "
    REQUIRE_STRING+="$*"
}

function dbg {
    # Just echo to stderr
    echo "$*" >&2
}

LAST_MAJOR=""
FIRST_MINOR=""
LAST_MINOR=""

for PY_VER in ${PY_VERSIONS}; do
    PY_VER_NODOTS=${PY_VER//.}
    dbg "PY_VER_NODOTS='${PY_VER_NODOTS}', REQUIRE_STRING='${REQUIRE_STRING}'"
    add_requirement "python${PY_VER_NODOTS}-base"
    dbg "REQUIRE_STRING='${REQUIRE_STRING}'"
    PY_VER_MAJOR=$(echo ${PY_VER} | cut -d. -f1)
    PY_VER_MINOR=$(echo ${PY_VER} | cut -d. -f2)
    dbg "PY_VER_MAJOR='${PY_VER_MAJOR}', PY_VER_MINOR='${PY_VER_MINOR}'"
    if [[ -n ${LAST_MAJOR} ]]; then
        dbg "LAST_MAJOR='${LAST_MAJOR}'"
        if [[ ${PY_VER_MAJOR} -eq ${LAST_MAJOR} && $((LAST_MINOR + 1)) -eq ${PY_VER_MINOR} ]]; then
            dbg "LAST_MINOR='${LAST_MINOR}'"
            LAST_MINOR=${PY_VER_MINOR}
            dbg "LAST_MINOR='${LAST_MINOR}'"
            continue
        fi
        dbg "REQUIRE_STRING='${REQUIRE_STRING}'"
        add_requirement "(python${LAST_MAJOR}-base >= ${LAST_MAJOR}.${FIRST_MINOR} and python${LAST_MAJOR}-base < ${LAST_MAJOR}.$((LAST_MINOR + 1)))"
        dbg "REQUIRE_STRING='${REQUIRE_STRING}'"
    fi
    LAST_MAJOR=${PY_VER_MAJOR}
    FIRST_MINOR=${PY_VER_MINOR}
    LAST_MINOR=${PY_VER_MINOR}
    dbg "LAST_MAJOR='${LAST_MAJOR}', FIRST_MINOR='${FIRST_MINOR}', LAST_MINOR='${LAST_MINOR}'"
done
dbg "REQUIRE_STRING='${REQUIRE_STRING}'"
add_requirement "(python${LAST_MAJOR}-base >= ${LAST_MAJOR}.${FIRST_MINOR} and python${LAST_MAJOR}-base < ${LAST_MAJOR}.$((LAST_MINOR + 1)))"
dbg "REQUIRE_STRING='${REQUIRE_STRING}'"

echo "(${REQUIRE_STRING})"
exit 0
