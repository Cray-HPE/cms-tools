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
cfs session deleter class for parallel deletion of cfs sessions
"""

import urllib3
import threading
import requests
import random

from cmstools.lib.api.api import API_REQUEST_TIMEOUT, add_api_auth
from cmstools.lib.cfs.defs import CFS_SESSIONS_URL_TEMPLATE
from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session import delete_cfs_session_by_name, get_cfs_sessions_list

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class CfsSessionDeleter:
    def __init__(self, script_args: ScriptArgs, cfs_session_name_list: list[str]):

        self.name_prefix = script_args.cfs_session_name
        self.max_sessions = script_args.max_cfs_sessions
        self.max_multi_delete_reqs = script_args.max_multi_cfs_sessions_delete_requests
        self.max_multi_get_reqs = script_args.max_multi_cfs_sessions_get_requests
        self.cfs_version = script_args.cfs_version
        self.deleted_sessions_lists = []  # Only used for v3
        self.cfs_session_names_list = cfs_session_name_list
        self.page_size = script_args.page_size # only used for V3 GET requests

    @property
    def expected_http_status(self) -> int:
        if self.cfs_version == "v2":
            return 204
        return 200

    def _get_sessions_thread(self, results_list) -> None:
        """
        Thread target function to get CFS sessions with the specified name prefix and pending status.
        """
        logger.info("Starting get sessions thread")
        sessions_list = get_cfs_sessions_list(
            cfs_session_name_contains=self.name_prefix,
            cfs_version=self.cfs_version,
            limit=self.page_size, retry=False)

        if sessions_list is not None:
            logger.debug(f"Got {sessions_list} sessions in thread with name prefix {self.name_prefix}")
            with threading.Lock():
                results_list.append(sessions_list)


    def _delete_sessions_thread(self, results_list) -> None:
        """
        Thread target function to delete CFS sessions with the specified name prefix and pending status.
        Verify that all multi-delete requests returned successful status and did not time out
        """
        logger.info("Starting delete sessions thread")
        params = {
            "status": "pending",
            "name_contains": self.name_prefix
        }
        url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=self.cfs_version)
        try:
            headers = {}
            add_api_auth(headers)
            resp = requests.delete(url=url, params=params, timeout=API_REQUEST_TIMEOUT, headers=headers, verify=False)

            if resp.status_code == self.expected_http_status and self.cfs_version == "v3":
                deleted = resp.json() if resp.content else []
                logger.info(f"Deleted sessions: {deleted}")
                with threading.Lock():
                    results_list.append(deleted)
                return

            if resp.status_code == 400:
                logger.info(f"Bad request response from Delete to {url}: {resp.text}")
                return

            if resp.status_code not in (self.expected_http_status, 400):
                logger.error(f"Unexpected return code {resp.status_code} from Delete to {url}: {resp.text}")
                raise CFSRCException()

        except Exception as exc:
            logger.exception(f"Exception during CFS session delete: {str(exc)}")
            raise

    def verify_v3_api_deleted_cfs_sessions_response(self) -> None:
        """
        Verify that all sessions were deleted with no duplicates based on the lists of deleted session names
        """
        logger.info(f"Deleted sessions lists from threads: {self.deleted_sessions_lists}")
        # Flatten the list of lists into a single list of session names
        all_deleted_sessions = [session_name for sublist in self.deleted_sessions_lists for session_name in sublist.get('session_ids', [])]
        logger.info(f"All deleted sessions combined: {all_deleted_sessions}")
        unique_deleted_sessions = set(all_deleted_sessions)

        if len(all_deleted_sessions) != len(unique_deleted_sessions):
            logger.error(f"Duplicate session names found in deleted sessions: {all_deleted_sessions}")
            raise CFSRCException()

        if len(unique_deleted_sessions) != self.max_sessions:
            logger.error(f"Number of unique deleted sessions {len(unique_deleted_sessions)} does not match expected {self.max_sessions}")
            raise CFSRCException()

        logger.info(f"All {self.max_sessions} sessions successfully deleted with no duplicates")

    def verify_all_sessions_deleted(self) -> None:
        """
        Verify that all CFS sessions with the specified name prefix are deleted.
        If any sessions still exist, attempt to delete them one by one.
        """
        sessions = get_cfs_sessions_list(self.name_prefix, self.cfs_version, self.page_size)

        if sessions:
            cfs_session_list = [s["name"] for s in sessions]
            logger.info(f"CFS sessions still exist with name prefix {self.name_prefix}: {cfs_session_list}")

            for session_name in cfs_session_list:
                try:
                    delete_cfs_session_by_name(session_name, self.cfs_version)
                except CFSRCException:
                    logger.error(f"Failed to delete CFS session: {session_name}")
            raise CFSRCException()

    def verify_sessions_after_delete(self, sessions: list) -> None:
        """
        Verify that all CFS sessions with the specified name prefix are deleted after multi-delete.
        """
        logger.info(f"Verifying all CFS sessions with name prefix {self.name_prefix} are deleted")
        self.verify_all_sessions_deleted()

        if self.cfs_version == "v3":
            # When deletion is performed via the v3 API, verify the response
            # to make sure that all sessions were deleted with no duplicates
            self.deleted_sessions_lists = sessions
            self.verify_v3_api_deleted_cfs_sessions_response()

    def validate_multi_get_sessions_response(self, multi_get_results: list, created_session_names: list[str]) -> None:
        """
        Validate that every entry in each list returned by the multi-get requests is a dict
        and corresponds to one of the sessions we created.
        """
        for idx, session_list in enumerate(multi_get_results):
            for session in session_list:
                if not isinstance(session, dict):
                    logger.error(f"Entry in multi-get result at index {idx} is not a dict: {session}")
                    raise CFSRCException()
                if session.get("name") not in created_session_names:
                    logger.error(f"Session name {session.get('name')} not in created session names")
                    raise CFSRCException()
        logger.info("All multi-get session entries are valid and correspond to created sessions")

    def multi_delete_multi_get_sessions(self):
        """
        Get CFS sessions in parallel using multiple threads.
        Each thread will attempt to get all sessions with the specified name prefix and pending status.
        Verify that all get requests returned successful status and did not time out.
        """
        threads = []
        get_results = []  # For collecting sessions from each get thread
        results = []  # For collecting sessions from each delete thread
        for _ in range(self.max_multi_get_reqs):
            t = threading.Thread(target=self._get_sessions_thread, args=(get_results,))
            threads.append(t)

        for _ in range(self.max_multi_delete_reqs):
            t = threading.Thread(target=self._delete_sessions_thread, args=(results,))
            threads.append(t)

        # start all threads randomly in threads list
        random.shuffle(threads)
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        self.verify_sessions_after_delete(sessions=results)
        self.validate_multi_get_sessions_response(multi_get_results=get_results,
                                         created_session_names=self.cfs_session_names_list)

    def multi_delete_sessions(self):
        """
        Delete CFS sessions in parallel using multiple threads.
        Each thread will attempt to delete all sessions with the specified name prefix and pending status.
        (v3 only) If successful, each delete request will return a list of the session names that it deleted.
        These lists need to be saved. (This is only true for v3 – for v2 a successful response returns nothing)
        """
        threads = []
        results = []  # For collecting deleted session names (v3)

        for _ in range(self.max_multi_delete_reqs):
            t = threading.Thread(target=self._delete_sessions_thread, args=(results,))
            threads.append(t)
            t.start()
        for t in threads:
            t.join()

        self.verify_sessions_after_delete(sessions=results)


