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
cfs sessions race condition single delete subtest
"""

from typing import ClassVar
from http import HTTPStatus

from cmstools.lib.cfs.types import SessionDeleteResult
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.cfs.session import delete_cfs_session_by_name
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.helpers.concurrent_requests import ConcurrentRequestManager, RequestBatch
from cmstools.test.cfs_sessions_rc_test.helpers.response_handler import ResponseHandler
from cmstools.test.cfs_sessions_rc_test.subtests.cfs_session_base import CFSSessionBase
from cmstools.test.cfs_sessions_rc_test.cfs.session_creator import CfsSessionCreator


class CFSSessionSingleDeleteTest(CFSSessionBase):

    """
    Class to handle CFS session single-delete test
    """
    subtest_name: ClassVar[str] = "single_delete"

    def __init__(self, script_args: ScriptArgs) -> None:
        super().__init__(script_args=script_args)
        self.request_manager = ConcurrentRequestManager()
        self.response_handler = ResponseHandler(script_args=script_args, session_names=self._session_names)
        self.delete_result_list: list[SessionDeleteResult] = []

    def _setup(self) -> list[str]:
        """Create a single CFS session for single-delete testing."""
        logger.info("Creating 1 CFS session for single-delete test")
        # Temporarily override to create only 1 session
        script_args = self.script_args._replace(max_cfs_sessions=1)
        cfs_session_creator = CfsSessionCreator(script_args=script_args)
        session_names = cfs_session_creator.create_sessions()
        return session_names

    def _execute_test_logic(self) -> None:
        """
            Execute parallel delete requests for individual sessions.
            <max-single-delete-reqs> parallel delete requests, each of which is deleting the specific session we just created.
            404 responses are okay. We don't need to preserve the contents of the response, other than the status codes.
            Wait for all parallel jobs to complete.
            """
        logger.info("Starting single-delete test with %d parallel delete requests",
                    self.script_args.max_single_cfs_sessions_delete_requests)

        # Create a batch request for executing parallel delete requests
        batch = RequestBatch(
            max_parallel=self.script_args.max_single_cfs_sessions_delete_requests,
            request_func=self._delete_session_thread
        )

        threads = self.request_manager.create_batch(batch=batch)
        self.request_manager.execute_batch(threads=threads)

    def _delete_session_thread(self) -> None:
        """
        Thread target function to delete a specific CFS session.
        Handles the actual deletion and stores the result.
        """
        session_name = self._session_names[0]
        logger.info("Deleting session: %s", session_name)

        result = delete_cfs_session_by_name(
            cfs_session_name=session_name,
            cfs_version=self.script_args.cfs_version,
            retry=False
        )

        with self.lock:
            self.delete_result_list.append(result)

    def _validate_results(self) -> None:
        """
        Perform all the single-delete subtest validations.
        Verify that all single-delete requests returned expected status codes and did not time out.
        """
        self.response_handler.validate_single_delete_response(
            delete_result=self.delete_result_list,
            expected_status=HTTPStatus.NO_CONTENT
        )
