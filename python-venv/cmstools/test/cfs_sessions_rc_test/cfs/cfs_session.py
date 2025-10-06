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

from cmstools.lib.api import request
from cmstools.lib.cfs.defs import CFS_SESSIONS_URL_TEMPLATE
from cmstools.lib.defs import CmstoolsException as CFSRCException

from cmstools.test.cfs_sessions_rc_test.log import logger

def get_next_id(data: dict) -> str|None:
    """
    Get the next_id from the data if it exists.
    """
    next_obj = data.get("next")
    if isinstance(next_obj, dict):
        return next_obj.get("after_id")
    return None

def get_session_data(cfs_session_data: dict|list) -> list|None:
    """
    Get sessions list from the data as get cfs session data is different for V2 and V3.
    """
    if isinstance(cfs_session_data, list):
        return cfs_session_data
    if isinstance(cfs_session_data, dict) and "sessions" in cfs_session_data:
        return cfs_session_data["sessions"]

def cfs_session_exists(cfs_session_name_contains: str, cfs_version: str, limit: int) -> bool:
    """
    Returns True if any CFS sessions exist with the specified name prefix and pending status.
    """
    sessions_list = get_cfs_sessions_list(cfs_session_name_contains, cfs_version, limit)
    return bool(sessions_list)

def get_cfs_sessions_list_params(cfs_session_name_contains: str, cfs_version: str, limit: int|None) -> dict:
    """
    Returns the URL with parameters to list CFS sessions with the specified name prefix and pending status.
    """
    if cfs_version == "v2":
        return {
            "status": "pending",
            "name_contains": cfs_session_name_contains
        }

    return {
            "status": "pending",
            "name_contains": cfs_session_name_contains,
            "limit": limit
        }

def get_all_cfs_sessions_v2 (cfs_session_name_contains: str, cfs_version: str) -> list|None:
    """
    Return all CFS sessions using v2 API.No pagination support in v2.
    """
    params = get_cfs_sessions_list_params(
        cfs_session_name_contains=cfs_session_name_contains,
        cfs_version=cfs_version,
        limit=None
    )

    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    resp = request("get", url=url, params=params)

    if resp.status_code != 200:
        logger.error(f"Unexpected return code {resp.status_code} from GET query to {url}: {resp.text}")
        raise CFSRCException()

    sessions = resp.json()
    session_data = get_session_data(sessions)
    logger.debug(f"Session data found: {session_data}")
    return session_data

def get_all_cfs_sessions_v3(cfs_session_name_contains: str, cfs_version: str, limit: int) -> list:
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    params = get_cfs_sessions_list_params(
        cfs_session_name_contains=cfs_session_name_contains,
        cfs_version=cfs_version,
        limit=limit
    )
    all_sessions = []
    after_id = None

    while True:
        if after_id:
            params["after_id"] = after_id
        resp = request("get", url=url, params=params)
        if resp.status_code != 200:
            logger.error(f"Unexpected return code {resp.status_code} from GET query to {url}: {resp.text}")
            raise CFSRCException()
        data = resp.json()
        sessions = get_session_data(data)
        if not sessions:
            break
        all_sessions.extend(sessions)
        next_id = get_next_id(data)
        if not next_id:
            break
        after_id = next_id
    logger.debug(f"Session data found: {all_sessions}")
    return all_sessions

def get_cfs_sessions_list(cfs_session_name_contains: str, cfs_version: str, limit: int) -> list|None:
    """
    Returns a list of CFS sessions with the specified name prefix and pending status.
    """
    if cfs_version == "v2":
        return get_all_cfs_sessions_v2(
            cfs_session_name_contains=cfs_session_name_contains,
        cfs_version=cfs_version)

    return get_all_cfs_sessions_v3(
        cfs_session_name_contains=cfs_session_name_contains,
        cfs_version=cfs_version,
        limit=limit)

def delete_cfs_sessions(cfs_session_name_contains: str, cfs_version: str, limit: int) -> None:
    """
    Delete all CFS sessions with the specified name prefix and pending status.
    """
    sessions = get_cfs_sessions_list(cfs_session_name_contains, cfs_version, limit)
    if not sessions:
        logger.info(f"No CFS sessions found with name prefix {cfs_session_name_contains} and status pending to delete")
        return

    params = {
        "status": "pending",
        "name_contains": cfs_session_name_contains
    }
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    resp = request("delete", url=url, params=params)
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
    cfs_sessions_url = f"{url}/{cfs_session_name}"
    resp = request("delete", cfs_sessions_url)
    if resp.status_code == 204:
        logger.info(f"Deleted CFS session {cfs_session_name}")
        return

    if resp.status_code == 404:
        logger.info(f"CFS session {cfs_session_name} not found to delete")
        return

    logger.error(f"Unexpected return code {resp.status_code} from Delete to {url}: {resp.text}")
    raise CFSRCException()




