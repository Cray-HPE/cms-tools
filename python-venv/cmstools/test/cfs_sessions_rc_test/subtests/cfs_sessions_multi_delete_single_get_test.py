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
cfs sessions race condition multi delete single get test
"""

from typing import ClassVar

from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.cfs.session import get_cfs_session_by_name
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.helpers.concurrent_requests import RequestBatch
from cmstools.lib.cfs import SessionsGetResponse

from .cfs_sessions_multi_delete_test import CfsSessionMultiDeleteTest


class CFSSessionMultiDeleteSingleGetTest(CfsSessionMultiDeleteTest):
    """
    Tests CFS session API behavior under concurrent DELETE and GET operations.

    This test verifies that:
    1. GET requests during concurrent DELETEs return either 200 OK (session still exists)
       or 404 NOT_FOUND (already deleted) - never partial/corrupted data
    2. All DELETE operations complete successfully
    3. No timeouts or unexpected errors occur

    The test simulates real-world scenarios where multiple clients read and delete
    sessions simultaneously.
    """

    subtest_name: ClassVar[str] = "multi_delete_single_get"

    def __init__(self, script_args: ScriptArgs) -> None:
        super().__init__(script_args=script_args)
        self.get_result_list: list[SessionsGetResponse] = []

    def _execute_test_logic(self) -> None:
        """Execute parallel delete requests."""
        logger.info("Starting %s test with %d sessions", self.subtest_name, len(self._session_names))

        # Create a batch request for executing parallel delete requests
        batch = RequestBatch(
            max_parallel=self.script_args.max_multi_cfs_sessions_delete_requests,
            request_func=self.delete_sessions_thread
        )

        self._tlist.extend(self.request_manager.create_batch(batch=batch))

        # Create a batch request for executing parallel get requests
        max_get_requests_count = self.script_args.max_single_cfs_sessions_get_requests
        session_count = len(self._session_names)

        if max_get_requests_count >= session_count:
            # One thread per session - adjust max_parallel to session_count
            batch = RequestBatch(
                max_parallel=session_count,
                request_func=self.get_sessions_thread
            )
            self._tlist.extend(self.request_manager.create_batch_with_items(batch=batch, items=self._session_names))
            logger.info("Creating %d GET threads (one per session) for %d sessions",
                        session_count, session_count)
        else:
            # Use thread pool to process all sessions with limited parallelism
            logger.info("Creating thread pool with %d threads to process %d sessions",
                        max_get_requests_count, session_count)
            self._tlist.extend(
                self.request_manager.create_batch_with_pool(
                    batch=RequestBatch(
                        max_parallel=max_get_requests_count,
                        request_func=self.get_sessions_thread
                    ),
                    items=self._session_names
                )
            )

        self.request_manager.execute_batch(self._tlist, shuffle=True)

    def get_sessions_thread(self, session_name: str) -> None:
        """
        Thread function to get sessions by names.
        """
        logger.debug("Starting get session thread for session name: %s", session_name)
        response = get_cfs_session_by_name(
            cfs_session_name=session_name,
            cfs_version=self.script_args.cfs_version,
            retry=False
        )
        with self.lock:
            # Lock is not needed for appending to a list, but added here for future safety
            self.get_result_list.append(response)

    def _validate_results(self) -> None:
        """
        Perform all the multi-delete subtest validations.
        Verify that all single-get responses during multi-delete are valid.
        """
        exceptions = []

        validations = [
            (self.response_handler.verify_sessions_after_multi_delete, self.delete_result_list),
            (self.response_handler.validate_single_get_response, self.get_result_list)
        ]

        for validation_func, *args in validations:
            try:
                validation_func(*args)
            except Exception as e:
                exceptions.append(e)

        if exceptions:
            if len(exceptions) == 1:
                raise exceptions[0]
            raise ExceptionGroup("Validation errors occurred", exceptions)
