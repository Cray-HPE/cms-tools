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
CFS race condition multi delete test related functions
"""

import requests
import threading
from typing import Any

from cmstools.lib.api.api import API_REQUEST_TIMEOUT, add_api_auth, SYSTEM_CA_CERTS
from cmstools.lib.cfs.defs import CFS_SESSIONS_URL_TEMPLATE
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs, CFSRCException
from cmstools.test.cfs_sessions_rc_test.log import logger

from cmstools.test.cfs_sessions_rc_test.subtests.cfs_session_gd_base import CFSSessionGDBase
from cmstools.test.cfs_sessions_rc_test.helpers.response_handler import ResponseHandler
from cmstools.test.cfs_sessions_rc_test.helpers.concurrent_requests import ConcurrentRequestManager, RequestBatch



class CfsSessionMultiDeleteTest(CFSSessionGDBase):
    """
    Class to handle CFS session multi-delete test
    """
    def __init__(self, script_args: ScriptArgs) -> None:
        super().__init__(script_args=script_args)
        self.request_manager = ConcurrentRequestManager()
        self.response_handler = ResponseHandler(script_args=script_args, session_names=self._session_names)
        self._tlist: list[threading.Thread] = []
        self.delete_result_list: list[Any] = []

    @staticmethod
    def get_subtest_name() -> str:
        return "multi_delete"

    @property
    def expected_http_status(self) -> int:
        if self.script_args.cfs_version == "v2":
            return 204
        return 200

    def _execute_test_logic(self) -> None:
        """Execute parallel delete requests."""
        logger.info("Starting multi-delete test with %d sessions", len(self._session_names))

        # Create a batch request for executing parallel delete requests
        batch = RequestBatch(
            max_parallel=self.script_args.max_multi_cfs_sessions_delete_requests,
            request_func=self.delete_sessions_thread
        )

        self._tlist = self.request_manager.create_batch(batch=batch)
        self.request_manager.execute_batch(threads=self._tlist)

    def delete_sessions_thread(self) -> None:
        """
        Thread target function to delete CFS sessions with the specified name prefix and pending status.
        Verify that all multi-delete requests returned successful status and did not time out
        """
        logger.info("Starting delete sessions thread")
        params = {
            "status": "pending",
            "name_contains": self.script_args.cfs_session_name
        }
        url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=self.script_args.cfs_version)
        try:
            headers = {}
            add_api_auth(headers)
            resp = requests.delete(url=url, params=params, timeout=API_REQUEST_TIMEOUT, headers=headers, verify=SYSTEM_CA_CERTS)

            if resp.status_code == self.expected_http_status and self.script_args.cfs_version == "v3":
                deleted = resp.json() if resp.content else []
                logger.info("Deleted sessions: %s", deleted)
                with self.lock:
                    self.delete_result_list.append(deleted)
                return

            if resp.status_code == 400:
                logger.info("Bad request response from Delete to %s: %s", url, resp.text)
                return

            if resp.status_code not in (self.expected_http_status, 400):
                logger.error("Unexpected return code %d from Delete to %s: %s", resp.status_code, url, resp.text)
                raise CFSRCException()

        except Exception as exc:
            logger.exception("Exception during CFS session delete: %s", str(exc))
            raise

    def _validate_results(self) -> None:
        """
        Verify that all multi-delete requests returned successful status and did not time out
        List all sessions in pending state that have the text prefix string in their names,
        and verify that none exist (if this check fails, delete the sessions one by one, for test cleanup)
        (v3 only) For the session name lists that were returned for each multi-delete request:
        Validate that every test session appears in exactly 1 of the lists
        Validate that no other session names are included in the lists
        """
        self.response_handler.verify_sessions_after_multi_delete(self.delete_result_list)

    @staticmethod
    def execute(script_args: ScriptArgs) -> None:
        """
        Create <max-multi-delete-reqs> parallel multi-delete requests, each of which is deleting all pending
        sessions that have the text prefix string in their names.
        (v3 only) If successful, each delete request will return a list of the session names that it deleted.
        These lists need to be saved. (This is only true for v3 – for v2 a successful response returns nothing)
        Wait for all parallel jobs to complete.
        """
        test = CfsSessionMultiDeleteTest(script_args=script_args)
        test.run()

