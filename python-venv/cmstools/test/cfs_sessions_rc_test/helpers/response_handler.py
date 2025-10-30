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
Class to validate the response from CFS API calls.
"""

from http import HTTPStatus

from cmstools.lib.cfs import CFS_V2_SESSIONS_DELETE_CODES, CFS_V3_SESSIONS_DELETE_CODES, \
    SessionDeleteResult, MultiSessionsGetResult, SessionsGetResponse
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs, CFSRCException
from cmstools.test.cfs_sessions_rc_test.cfs.session import get_cfs_sessions_list
from cmstools.test.cfs_sessions_rc_test.log import logger


class ResponseHandler:
    """Class to handle and validate API responses."""

    def __init__(self, script_args: ScriptArgs, session_names: list[str]) -> None:
        self.script_args = script_args
        self.session_names = session_names

    def validate_multi_get_sessions_response(self, multi_get_results: list[MultiSessionsGetResult]) -> None:
        """
        Validate that every entry in each list returned by the multi-get requests is a dict
        and corresponds to one of the sessions we created.
        """

        errors = False
        # First check if none of the requests timed out
        timeouts = [r for r in multi_get_results if r.timed_out]
        if timeouts:
            logger.error("%d multi-get operations timed out", len(timeouts))
            errors = True

        # Filter out timed-out responses for further validation
        remaining_responses = [r for r in multi_get_results if not r.timed_out]

        # Then check if all requests returned a 200 status code
        invalid_responses = [r for r in remaining_responses if r.status_code != HTTPStatus.OK]
        if invalid_responses:
            logger.error("%d multi-get operations returned unexpected status codes", len(invalid_responses))
            errors = True

        # Only validate the content of successful responses (OK status)
        successful_gets = [r for r in remaining_responses if r.status_code == HTTPStatus.OK]

        # Now validate the content of each response
        for idx, session_list in enumerate(successful_gets):
            for session in session_list.session_data:
                if not isinstance(session, dict):
                    logger.error("Entry in multi-get result at index %d is not a dict: %s", idx, session)
                    errors = True

                if session.get("name") not in self.session_names:
                    logger.error("Session name %s not in created session names", session.get('name'))
                    errors = True
        if errors:
            raise CFSRCException()

        logger.info("All multi-get session entries are valid and correspond to created sessions")

    def verify_sessions_after_multi_delete(self, deleted_sessions_list: list[SessionDeleteResult]) -> None:
        """
        Verify that all CFS sessions with the specified name prefix are deleted after multi-delete.
        """

        errors = False
        # Verify if any of the delete operation timeout
        timeouts = [r for r in deleted_sessions_list if r.timed_out]
        if timeouts:
            logger.error("%d delete operations timed out", len(timeouts))
            errors = True

        # Filter out timed-out responses for further validation
        remaining_responses = [r for r in deleted_sessions_list if not r.timed_out]

        expected_codes = CFS_V3_SESSIONS_DELETE_CODES if self.script_args.cfs_version == "v3" else CFS_V2_SESSIONS_DELETE_CODES
        invalid_responses = [r for r in remaining_responses if r.status_code not in expected_codes]
        if invalid_responses:
            logger.error("%d delete operation returned unexpected status codes", len(invalid_responses))
            errors = True

        logger.info("Verifying all CFS sessions with name prefix %s are deleted", self.script_args.cfs_session_name)
        self.verify_all_sessions_deleted()

        if self.script_args.cfs_version == "v3":
            # # Only validate the content of successful responses (OK status)
            successful_responses = [r for r in remaining_responses if r.status_code == HTTPStatus.OK]
            # When deletion is performed via the v3 API, verify the response
            # to make sure that all sessions were deleted with no duplicates
            self.verify_v3_api_deleted_cfs_sessions_response(deleted_sessions_list=successful_responses)

        if errors:
            raise CFSRCException()

    def verify_all_sessions_deleted(self) -> None:
        """
        Verify that all CFS sessions with the specified name prefix are deleted.
        If any sessions still exist, attempt to delete them one by one.
        """
        result = get_cfs_sessions_list(cfs_session_name_contains=self.script_args.cfs_session_name,
                                       cfs_version=self.script_args.cfs_version,
                                       limit=self.script_args.page_size)

        if result.status_code == HTTPStatus.OK and result.session_data:
            cfs_session_list = [s["name"] for s in result.session_data]
            logger.error("CFS sessions still exist with name prefix %s: %s", self.script_args.cfs_session_name,
                         cfs_session_list)
            raise CFSRCException()

    def _is_valid_v3_delete_response(self, result: SessionDeleteResult) -> bool:
        return (result.status_code == HTTPStatus.OK and isinstance(result.session_data, dict) and
                'session_ids' in result.session_data)

    def verify_v3_api_deleted_cfs_sessions_response(self, deleted_sessions_list: list[SessionDeleteResult]) -> None:
        """
        Verify that all sessions were deleted with no duplicates based on the lists of deleted session names
        """

        for _, result in enumerate(deleted_sessions_list):
            if self._is_valid_v3_delete_response(result=result):
                logger.info("Deleted sessions: %s", result.session_data.get('session_ids', []))

        # Flatten the list of lists into a single list of session names
        all_deleted_sessions = [
            session_name
            for sublist in deleted_sessions_list
            if self._is_valid_v3_delete_response(result=sublist)
            for session_name in sublist.session_data.get('session_ids', [])
        ]

        logger.info("All deleted sessions combined: %s", all_deleted_sessions)
        unique_deleted_sessions = set(all_deleted_sessions)

        if len(all_deleted_sessions) != len(unique_deleted_sessions):
            logger.error("Duplicate session names found in deleted sessions: %s", all_deleted_sessions)
            raise CFSRCException()

        if len(unique_deleted_sessions) != self.script_args.max_cfs_sessions:
            logger.error("Number of unique deleted sessions %d does not match expected %d",
                         len(unique_deleted_sessions), self.script_args.max_cfs_sessions)
            raise CFSRCException()

        logger.info("All %d sessions successfully deleted with no duplicates", self.script_args.max_cfs_sessions)

    def validate_single_get_response(self, result: list[SessionsGetResponse]) -> None:
        """
        Validate the single-get responses during multi-delete.
        Verify that all single-get requests returned successful status OR 404, and did not time out (it is fine if they
        all returned 404 or all were successful, or any mix)
        ️For all the single-get requests which succeeded with non-404 status (if any),
        validate that the expected session data was returned.
        ️(It is fine if some sessions are not listed in any of the responses.
        It is fine if no sessions are listed in any of the responses.)
        """

        errors = False

        # First, check for timeouts and remove them from further validation
        timeouts = [r for r in result if r.timed_out]
        if timeouts:
            logger.error("%d GET requests timed out", len(timeouts))
            errors = True

        # Filter out timed-out responses for further validation
        remaining_responses = [r for r in result if r not in timeouts]

        # Check for unexpected status codes in the remaining responses
        invalid_responses = [r for r in remaining_responses if r.status_code not in [HTTPStatus.OK, HTTPStatus.NOT_FOUND]]
        if invalid_responses:
            status_codes = [r.status_code for r in invalid_responses]
            logger.error("%d GET requests returned unexpected status codes: %s",
                         len(invalid_responses), status_codes)
            errors = True

        # Only validate the content of successful responses (OK status)
        successful_gets = [r for r in remaining_responses if r.status_code == HTTPStatus.OK]
        for resp in successful_gets:
            if not isinstance(resp.session_data, dict):
                logger.error("GET response is not a dict: %s", resp.session_data)
                errors = True
                continue

            if resp.session_data.get("name") not in self.session_names:
                logger.error("Session name %s not in expected session names", resp.session_data.get('name'))
                errors = True

        if errors:
            raise CFSRCException()

        logger.info("Validated %d GET responses: %d successful, %d not found, %d timed out",
                    len(result), len(successful_gets),
                    len([r for r in result if r.status_code == HTTPStatus.NOT_FOUND]),
                    len(timeouts))
