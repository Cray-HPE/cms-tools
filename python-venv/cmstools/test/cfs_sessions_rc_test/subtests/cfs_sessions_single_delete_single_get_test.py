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
CFS race condition single delete and single get test related functions
"""
from http import HTTPStatus
import threading
from typing import ClassVar

from cmstools.lib.cfs.types import SessionsGetResponse
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.cfs.session import get_cfs_session_by_name
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.helpers.concurrent_requests import RequestBatch
from cmstools.test.cfs_sessions_rc_test.subtests.cfs_sessions_single_delete_test import CFSSessionSingleDeleteTest


class CFSSessionSingleDeleteSingleGetTest(CFSSessionSingleDeleteTest):
    """
    Class to handle CFS session single-delete and single-get test
    """
    subtest_name: ClassVar[str] = "single_delete_single_get"

    def __init__(self, script_args: ScriptArgs) -> None:
        super().__init__(script_args=script_args)
        self.get_result_list: list[SessionsGetResponse] = []
        self._tlist: list[threading.Thread] = []

    def _setup(self) -> list[str]:
        # Use the parent setup to create a single session
        return super()._setup()

    def _execute_test_logic(self) -> None:
        """Execute single delete and single get requests."""
        logger.info("Starting single-delete single-get test with session '%s'", self._session_names[0])

        # Create a batch request for executing parallel delete requests
        batch = RequestBatch(
            max_parallel=self.script_args.max_single_cfs_sessions_delete_requests,
            request_func=self._delete_session_thread
        )
        self._tlist.extend(self.request_manager.create_batch(batch=batch))

        # Create a batch request for executing parallel single get requests
        get_batch = RequestBatch(
            max_parallel=self.script_args.max_single_cfs_sessions_get_requests,
            request_func=self._get_session_thread
        )

        # Creating multiple batches to reach the total number of get requests
        for _ in range(self.script_args.max_single_cfs_sessions_get_requests):
            self._tlist.extend(self.request_manager.create_batch_with_items(batch=get_batch, items=self._session_names))

        self.request_manager.execute_batch(threads=self._tlist, shuffle=True)

    def _get_session_thread(self, session_name: str) -> None:
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
        Verify that all single-delete requests returned expected status codes and did not time out.
        Verify that all single-get requests returned successful status OR 404, and did not time out
        (it is fine if they all returned 404 or all were successful, or any mix)
        For all the single-get requests which succeeded with non-404 status, validate that the
        expected session data was returned.
        """
        exceptions = []

        validations = [
            (self.response_handler.validate_single_delete_response, self.delete_result_list, HTTPStatus.NO_CONTENT),
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
