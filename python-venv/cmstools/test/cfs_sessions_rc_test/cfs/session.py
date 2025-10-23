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

import requests
from http import HTTPStatus

from cmstools.lib.api import (request, request_and_check_status,
                              add_api_auth, SYSTEM_CA_CERTS)
from cmstools.lib.cfs import CFS_SESSIONS_URL_TEMPLATE, MultiSessionsGetResult
from cmstools.lib.defs import JsonDict
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.defs import CFSRCException, API_REQUEST_TIMEOUT, CfsVersionsStrLiteral

DELETE_SESSIONS_SUCCESS_STATUS_CODES: dict[CfsVersionsStrLiteral, int] = {
    "v2": HTTPStatus.NO_CONTENT,
    "v3": HTTPStatus.OK
}


def get_next_id(data: JsonDict) -> str | None:
    """
    Get the next_id from the data if it exists.
    """
    next_obj = data.get("next")
    if isinstance(next_obj, dict):
        return next_obj.get("after_id")
    return None


def get_session_data(cfs_session_data: JsonDict | list[JsonDict]) -> list[JsonDict] | None:
    """
    Get sessions list from the data as get cfs session data is different for V2 and V3.
    """
    if isinstance(cfs_session_data, list):
        return cfs_session_data
    if isinstance(cfs_session_data, dict) and "sessions" in cfs_session_data:
        return cfs_session_data["sessions"]
    return None


def cfs_session_exists(cfs_session_name_contains: str, cfs_version: CfsVersionsStrLiteral, limit: int) -> bool:
    """
    Returns True if any CFS sessions exist with the specified name prefix and pending status.
    """
    result = get_cfs_sessions_list(cfs_session_name_contains, cfs_version, limit)
    return result.status_code == HTTPStatus.OK and bool(result.session_data)


def get_cfs_sessions_list_params(cfs_session_name_contains: str, cfs_version: CfsVersionsStrLiteral, limit: int | None) -> dict[str, str | int | None]:
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


def make_request(url: str, params: dict, retry: bool) -> requests.Response:
    """Make a GET request with optional retry logic."""
    if not retry:
        logger.debug("No retry for requests")
        headers = {}
        add_api_auth(headers)
        return requests.get(url=url, params=params, timeout=API_REQUEST_TIMEOUT,
                            headers=headers, verify=SYSTEM_CA_CERTS)
    return request("get", url=url, params=params)


def get_all_cfs_sessions_v2(cfs_session_name_contains: str, cfs_version: CfsVersionsStrLiteral, retry: bool) -> MultiSessionsGetResult:
    """
    Return all CFS sessions using v2 API.No pagination support in v2.
    """
    params = get_cfs_sessions_list_params(
        cfs_session_name_contains=cfs_session_name_contains,
        cfs_version=cfs_version,
        limit=None
    )

    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    try:
        resp = make_request(url=url, params=params, retry=retry)

        if resp.status_code != HTTPStatus.OK:
            logger.error("Unexpected return code %d from GET query to %s: %s", resp.status_code, url, resp.text)
            return MultiSessionsGetResult(status_code=resp.status_code, error_message=resp.text)

        sessions = resp.json()
        session_data = get_session_data(sessions)
        if not session_data:
            return MultiSessionsGetResult(status_code=HTTPStatus.OK, session_data=[], error_message="No sessions found")
        logger.debug("Session data found: %s", session_data)
        return MultiSessionsGetResult(status_code=HTTPStatus.OK, session_data=session_data)
    except requests.exceptions.Timeout:
        logger.exception("Request timed out for GET query to %s", url)
        return MultiSessionsGetResult(status_code=0, timed_out=True, error_message="Request timed out")
    except Exception as exc:
        logger.exception("Exception during CFS session get: %s", str(exc))
        return MultiSessionsGetResult(status_code=0, error_message=str(exc))


def get_all_cfs_sessions_v3(cfs_session_name_contains: str, cfs_version: CfsVersionsStrLiteral, limit: int, retry: bool) -> MultiSessionsGetResult:
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    params = get_cfs_sessions_list_params(
        cfs_session_name_contains=cfs_session_name_contains,
        cfs_version=cfs_version,
        limit=limit
    )
    all_sessions: list[JsonDict] = []
    after_id: str | None = None

    try:
        while True:
            if after_id:
                params["after_id"] = after_id
            resp = make_request(url=url, params=params, retry=retry)
            if resp.status_code != HTTPStatus.OK:
                logger.error("Unexpected return code %d from GET query to %s: %s", resp.status_code, url, resp.text)
                return MultiSessionsGetResult(status_code=resp.status_code, error_message=resp.text)
            data = resp.json()
            sessions = get_session_data(data)
            if not sessions:
                break
            all_sessions.extend(sessions)
            next_id = get_next_id(data)
            if not next_id:
                break
            after_id = next_id
        logger.debug("Session data found: %s", all_sessions)
        return MultiSessionsGetResult(status_code=HTTPStatus.OK, session_data=all_sessions)
    except requests.exceptions.Timeout:
        logger.exception("Request timed out for GET query to %s", url)
        return MultiSessionsGetResult(status_code=0, timed_out=True, error_message="Request timed out")
    except Exception as exc:
        logger.exception("Exception during CFS session get: %s", str(exc))
        return MultiSessionsGetResult(status_code=0, error_message=str(exc))


def get_cfs_sessions_list(cfs_session_name_contains: str, cfs_version: CfsVersionsStrLiteral, limit: int, retry: bool = True) -> MultiSessionsGetResult:
    """
    Returns a list of CFS sessions with the specified name prefix and pending status.
    """
    if cfs_version == "v2":
        return get_all_cfs_sessions_v2(
            cfs_session_name_contains=cfs_session_name_contains,
            cfs_version=cfs_version, retry=retry)

    return get_all_cfs_sessions_v3(
        cfs_session_name_contains=cfs_session_name_contains,
        cfs_version=cfs_version,
        limit=limit, retry=retry)


def delete_cfs_sessions(cfs_session_name_contains: str, cfs_version: CfsVersionsStrLiteral, limit: int) -> None:
    """
    Delete all CFS sessions with the specified name prefix and pending status.
    """
    result = get_cfs_sessions_list(cfs_session_name_contains, cfs_version, limit)

    if result.status_code != HTTPStatus.OK:
        logger.error("Get CFS sessions list failed: expected %d or %d, got status_code=%d, error=%s",
                     HTTPStatus.OK, HTTPStatus.BAD_REQUEST, result.status_code, result.error_message)
        raise CFSRCException()

    if not result.session_data:
        logger.info("No CFS sessions found with name prefix %s and status pending to delete",
                    cfs_session_name_contains)
        return

    params = {
        "status": "pending",
        "name_contains": cfs_session_name_contains
    }
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    resp = request("delete", url=url, params=params)

    if resp.status_code == DELETE_SESSIONS_SUCCESS_STATUS_CODES[cfs_version]:
        logger.info("Deleted CFS sessions with name prefix %s and status pending", cfs_session_name_contains)
        return

    if resp.status_code == HTTPStatus.BAD_REQUEST:
        return

    logger.error("Unexpected return code %d from Delete to %s: %s", resp.status_code, url, resp.text)
    raise CFSRCException()


def delete_cfs_session_by_name(cfs_session_name: str, cfs_version: CfsVersionsStrLiteral) -> None:
    """
    Delete a CFS session by name.
    """
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)
    cfs_sessions_url = f"{url}/{cfs_session_name}"
    resp = request("delete", cfs_sessions_url)
    if resp.status_code == HTTPStatus.NO_CONTENT:
        logger.info("Deleted CFS session %s", cfs_session_name)
        return

    if resp.status_code == HTTPStatus.NOT_FOUND:
        logger.info("CFS session %s not found to delete", cfs_session_name)
        return

    logger.error("Unexpected return code %d from Delete to %s: %s", resp.status_code, url, resp.text)
    raise CFSRCException()


def create_cfs_session(session_name: str, cfs_version: CfsVersionsStrLiteral,
                       session_payload: dict, expected_http_status: int = HTTPStatus.OK) -> JsonDict:
    """
    Create a CFS session with the specified name and configuration.

    Args:
        session_name: Name of the session to create
        cfs_version: CFS API version to use ("v2" or "v3")
        session_payload: Payload for the session creation
        expected_http_status: Expected HTTP status code for the creation request

    Returns:
        Dictionary containing the created session data

    """
    url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=cfs_version)

    resp_data = request_and_check_status("post", url,
                                         json=session_payload,
                                         expected_status=expected_http_status,
                                         parse_json=True)
    logger.debug("Created CFS session %s: %s", session_name, resp_data)
    return resp_data
