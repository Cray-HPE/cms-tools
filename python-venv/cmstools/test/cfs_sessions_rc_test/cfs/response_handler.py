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

from typing import Any, List

from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session import get_cfs_sessions_list
from cmstools.test.cfs_sessions_rc_test.log import logger


class ResponseHandler:
    """Class to handle and validate API responses."""

    def __init__(self, script_args: ScriptArgs, session_names: List[str]) -> None:
        self.script_args = script_args
        self.session_names = session_names
        self.deleted_sessions_lists = []

    def validate_multi_get_sessions_response(self, multi_get_results: List[Any]) -> None:
        """
        Validate that every entry in each list returned by the multi-get requests is a dict
        and corresponds to one of the sessions we created.
        """
        for idx, session_list in enumerate(multi_get_results):
            for session in session_list:
                if not isinstance(session, dict):
                    logger.error(f"Entry in multi-get result at index {idx} is not a dict: {session}")
                    raise CFSRCException()
                if session.get("name") not in self.session_names:
                    logger.error(f"Session name {session.get('name')} not in created session names")
                    raise CFSRCException()
        logger.info("All multi-get session entries are valid and correspond to created sessions")

    def verify_sessions_after_multi_delete(self, deleted_sessions_list: List[Any]) -> None:
        """
        Verify that all CFS sessions with the specified name prefix are deleted after multi-delete.
        """
        logger.info(f"Verifying all CFS sessions with name prefix {self.script_args.cfs_session_name} are deleted")
        self.verify_all_sessions_deleted()

        if self.script_args.cfs_version == "v3":
            # When deletion is performed via the v3 API, verify the response
            # to make sure that all sessions were deleted with no duplicates
            self.verify_v3_api_deleted_cfs_sessions_response(deleted_sessions_list=deleted_sessions_list)

    def verify_all_sessions_deleted(self) -> None:
        """
        Verify that all CFS sessions with the specified name prefix are deleted.
        If any sessions still exist, attempt to delete them one by one.
        """
        sessions = get_cfs_sessions_list(cfs_session_name_contains=self.script_args.cfs_session_name,
                                         cfs_version=self.script_args.cfs_version,
                                         limit=self.script_args.page_size)

        if sessions:
            cfs_session_list = [s["name"] for s in sessions]
            logger.error(f"CFS sessions still exist with name prefix {self.script_args.cfs_session_name}: {cfs_session_list}")

    def verify_v3_api_deleted_cfs_sessions_response(self, deleted_sessions_list: List[Any]) -> None:
        """
        Verify that all sessions were deleted with no duplicates based on the lists of deleted session names
        """
        logger.debug(f"Deleted sessions lists from threads: {deleted_sessions_list}")
        # Flatten the list of lists into a single list of session names
        all_deleted_sessions = [session_name for sublist in deleted_sessions_list for session_name in
                                sublist.get('session_ids', [])]
        logger.info(f"All deleted sessions combined: {all_deleted_sessions}")
        unique_deleted_sessions = set(all_deleted_sessions)

        if len(all_deleted_sessions) != len(unique_deleted_sessions):
            logger.error(f"Duplicate session names found in deleted sessions: {all_deleted_sessions}")
            raise CFSRCException()

        if len(unique_deleted_sessions) != self.script_args.max_cfs_sessions:
            logger.error(
                f"Number of unique deleted sessions {len(unique_deleted_sessions)} does not match expected {self.script_args.max_cfs_sessions}")
            raise CFSRCException()

        logger.info(f"All {self.script_args.max_cfs_sessions} sessions successfully deleted with no duplicates")
