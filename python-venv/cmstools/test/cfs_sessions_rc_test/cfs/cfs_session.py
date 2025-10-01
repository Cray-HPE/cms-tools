#
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
CFS Session related functions
"""
import urllib.parse

from cmstools.lib.api import request
from cmstools.lib.cfs.defs import CFS_SESSIONS_URL_TEMPLATE
from cmstools.lib.defs import CmstoolsException as CFSRCException

from cmstools.test.cfs_sessions_rc_test.log import logger

def get_session_data(cfs_session_data: dict|list) -> list|None:
    """
    Get sessions list from the data as get cfs session data is different for V2 and V3.
    """
    if isinstance(cfs_session_data, list):
        return cfs_session_data
    if isinstance(cfs_session_data, dict) and "sessions" in cfs_session_data:
        return cfs_session_data["sessions"]

def cfs_session_exists(cfs_session_name_contains: str, cfs_version: str) -> bool:
    """
    Returns True if any CFS sessions exist with the specified name prefix and pending status.
    """
    sessions_list = get_cfs_sessions_list(cfs_session_name_contains, cfs_version)
    return bool(sessions_list)

def get_cfs_sessions_list(cfs_session_name_contains: str, cfs_version: str) -> list|None:
    """
    Returns a list of CFS sessions with the specified name prefix and pending status.
    """
    params = {
        "status": "pending",
        "name_contains": cfs_session_name_contains
    }
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    cfs_sessions_url_with_params = f"{url}?{urllib.parse.urlencode(params)}"
    logger.info(f"Getting CFS sessions with URL: {url}")
    resp = request("get", cfs_sessions_url_with_params)
    logger.info(f"GET {url} returned status code {resp.status_code}")

    if resp.status_code != 200:
        logger.error(f"Unexpected return code {resp.status_code} from GET query to {url}: {resp.text}")

    sessions = resp.json()
    logger.info(f"Found {sessions} CFS sessions with name prefix {cfs_session_name_contains} and status pending")
    session_data = get_session_data(sessions)
    logger.info(f"Session data found: {session_data}")
    return session_data

def delete_cfs_sessions(cfs_session_name_contains: str, cfs_version: str) -> None:
    """
    Delete all CFS sessions with the specified name prefix and pending status.
    """
    params = {
        "status": "pending",
        "name_contains": cfs_session_name_contains
    }
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    cfs_sessions_url_with_params = f"{url}?{urllib.parse.urlencode(params)}"
    resp = request("delete", cfs_sessions_url_with_params)
    if resp.status_code == 200:
        logger.info(f"Deleted CFS sessions with name prefix {cfs_session_name_contains} and status pending")
        return

    if resp.status_code == 400:
        return

    logger.error(f"Unexpected return code {resp.status_code} from Delete to {url}: {resp.text}")
    raise CFSRCException()

def delete_cfs_session_by_name(cfs_session_name: str, cfs_version: str) -> None:
    """
    Delete a CFS session by name.
    """
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    cfs_sessions_url_with_params = f"{url}/{cfs_session_name}"
    resp = request("delete", cfs_sessions_url_with_params)
    if resp.status_code == 204:
        logger.info(f"Deleted CFS session {cfs_session_name}")
        return

    if resp.status_code == 404:
        logger.info(f"CFS session {cfs_session_name} not found to delete")
        return

    logger.error(f"Unexpected return code {resp.status_code} from Delete to {url}: {resp.text}")
    raise CFSRCException()




