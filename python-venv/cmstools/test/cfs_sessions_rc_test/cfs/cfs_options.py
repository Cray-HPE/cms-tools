
# MIT License
#
# (C) Copyright 2025 Hewlett Packard Enterprise Development LP
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

"""
CFS options related functions
"""

from cmstools.lib.api import request
from cmstools.lib.cfs.defs import CFS_OPTIONS_URL
from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.test.cfs_sessions_rc_test.log import logger

def get_cfs_options() -> dict:
    """
    Returns the current CFS options values
    """
    resp = request("get", CFS_OPTIONS_URL)

    if resp.status_code != 200:
        logger.error(f"Failed to get CFS options: {resp.status_code} {resp.text}")
        raise CFSRCException()
    return resp.json()

def get_cfs_page_size() -> int:
    """
    Returns the current CFS page size option value
    """
    options = get_cfs_options()
    return options.get("default_page_size", 1000)  # Default to 1000 if not set

def set_cfs_page_size(page_size: int) -> None:
    """
    Sets the CFS page size option value
    """
    data = {
        "default_page_size": page_size
    }
    resp = request("patch", CFS_OPTIONS_URL, json=data)

    if resp.status_code != 200:
        logger.error(f"Failed to set CFS page size to {page_size}: {resp.text}")
        raise CFSRCException()
