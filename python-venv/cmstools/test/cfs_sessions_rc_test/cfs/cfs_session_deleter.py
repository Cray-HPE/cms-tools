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

from cmstools.lib.api.api import API_REQUEST_TIMEOUT, add_api_auth
from cmstools.lib.cfs.defs import CFS_SESSIONS_URL_TEMPLATE
from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session import delete_cfs_session_by_name, get_cfs_sessions_list

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class CfsSessionDeleter:
    def __init__(self, name_prefix: str, max_sessions: int, max_multi_delete_reqs: int, cfs_session_name_list: list[str],
                 page_size: int, cfs_version: str = "v3"):
        self.name_prefix = name_prefix
        self.max_sessions = max_sessions
        self.max_multi_delete_reqs = max_multi_delete_reqs
        self.cfs_version = cfs_version
        self.deleted_sessions_lists = []  # Only used for v3
        self.cfs_session_names_list = cfs_session_name_list
        self.page_size = page_size # only used for V3 GET requests

    @property
    def expected_http_status(self) -> int:
        if self.cfs_version == "v2":
            return 204
        return 200

    def _delete_sessions_thread(self, results_list) -> None:
        """
        Thread target function to delete CFS sessions with the specified name prefix and pending status.
        Verify that all multi-delete requests returned successful status and did not time out
        """
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

    def delete_sessions(self):
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

        # Verify all sessions are deleted
        logger.info(f"Verifying all CFS sessions with name prefix {self.name_prefix} are deleted")
        self.verify_all_sessions_deleted()

        if self.cfs_version == "v3":
            # When deletion is performed via the v3 API, verify the response
            # to make sure that all sessions were deleted with no duplicates
            self.deleted_sessions_lists = results
            self.verify_v3_api_deleted_cfs_sessions_response()
