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

import urllib3
import requests
import threading
from typing import List, Any

from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.lib.api.api import API_REQUEST_TIMEOUT, add_api_auth
from cmstools.lib.cfs.defs import CFS_SESSIONS_URL_TEMPLATE
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.log import logger

from .cfs_session_gd_base import CFSSessionGDBase
from .response_handler import ResponseHandler
from .concurrent_requests import ConcurrentRequestManager, RequestBatch

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)



class CfsSessionMultiDeleteTest(CFSSessionGDBase):
    """
    Class to handle CFS session multi-delete test
    """
    def __init__(self, script_args: ScriptArgs) -> None:
        super().__init__(script_args=script_args)
        self.request_manager = ConcurrentRequestManager()
        self.response_handler = ResponseHandler(script_args=script_args, session_names=self._session_names)
        self._tlist: List[threading.Thread] = []
        self.delete_result_list: List[Any] = []

    @property
    def expected_http_status(self) -> int:
        if self.script_args.cfs_version == "v2":
            return 204
        return 200

    def _execute_test_logic(self) -> None:
        """Execute parallel delete requests."""
        logger.info(f"Starting multi-delete test with {len(self._session_names)} sessions")

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
            resp = requests.delete(url=url, params=params, timeout=API_REQUEST_TIMEOUT, headers=headers, verify=False)

            if resp.status_code == self.expected_http_status and self.script_args.cfs_version == "v3":
                deleted = resp.json() if resp.content else []
                logger.info(f"Deleted sessions: {deleted}")
                with self.lock:
                    self.delete_result_list.append(deleted)
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

    def _validate_results(self) -> None:
        """Validate test results."""
        self.response_handler.verify_sessions_after_multi_delete(self.delete_result_list)

def execute(script_args: ScriptArgs) -> None:
    """
    Create <max-multi-delete-reqs> parallel multi-delete requests, each of which is deleting all pending
    sessions that have the text prefix string in their names.
    (v3 only) If successful, each delete request will return a list of the session names that it deleted.
    These lists need to be saved. (This is only true for v3 – for v2 a successful response returns nothing)
    Wait for all parallel jobs to complete.
    """
    test = CfsSessionMultiDeleteTest(script_args=script_args)
    test.execute()

