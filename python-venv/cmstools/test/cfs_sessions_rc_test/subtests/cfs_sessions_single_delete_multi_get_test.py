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
CFS race condition single delete and multi get test related functions
"""
import threading
from http import HTTPStatus
from typing import ClassVar

from cmstools.lib.cfs import MultiSessionsGetResult
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.cfs.session import get_cfs_sessions_list
from cmstools.test.cfs_sessions_rc_test.subtests.cfs_sessions_single_delete_test import CFSSessionSingleDeleteTest
from cmstools.test.cfs_sessions_rc_test.helpers.concurrent_requests import RequestBatch


class CFSSessionSingleDeleteMultiGetTest(CFSSessionSingleDeleteTest):
    """
    Class to handle CFS session single-delete multi-get test
    """
    subtest_name: ClassVar[str] = "single_delete_multi_get"

    def __init__(self, script_args: ScriptArgs) -> None:
        super().__init__(script_args=script_args)
        self.get_result_list: list[MultiSessionsGetResult] = []
        self._tlist: list[threading.Thread] = []

    def _setup(self) -> list[str]:
        return super()._setup()

    def _execute_test_logic(self) -> None:
        """
        Execute single delete request.
        Create <max-multi-get-reqs> parallel multi-get requests to get all sessions that have the text prefix string in
        their names.
        """
        logger.info("Starting single-delete multi-get test with session '%s'", self._session_names[0])

        # Create a batch request for executing parallel delete requests
        batch = RequestBatch(
            max_parallel=self.script_args.max_single_cfs_sessions_delete_requests,
            request_func=self._delete_session_thread
        )
        self._tlist.extend(self.request_manager.create_batch(batch=batch))

        # Create a batch request for executing parallel get requests
        get_batch = RequestBatch(
            max_parallel=self.script_args.max_multi_cfs_sessions_get_requests,
            request_func=self._get_sessions_thread,
        )
        self._tlist.extend(self.request_manager.create_batch(batch=get_batch))

        self.request_manager.execute_batch(threads=self._tlist, shuffle=True)

    def _get_sessions_thread(self) -> None:
        """
        Thread target function to get CFS sessions with the specified name prefix and pending status.
        """
        logger.info("Starting get sessions thread")
        result = get_cfs_sessions_list(
            cfs_session_name_contains=self.script_args.cfs_session_name,
            cfs_version=self.script_args.cfs_version,
            limit=self.script_args.page_size, retry=False)

        if result.session_data is not None and result.status_code == HTTPStatus.OK:
            logger.debug("Got %s sessions in thread with name prefix %s", result.session_data,
                         self.script_args.cfs_session_name)
            # Lock is not needed for appending to a list, but added here for future safety
            with self.lock:
                self.get_result_list.append(result)

    def _validate_results(self) -> None:
        """
        Verify that exactly 1 of the delete requests returned 204.
        Verify that the session no longer exists.
        Verify that all multi-get requests returned successful status and did not time out.
        For the session lists that were returned by the multi-get requests, validate that every list
        is either empty, or only contains a dict for the test session we created for this subtest.
        (It is fine if every list is empty)
        """
        exceptions = []

        validations = [
            (self.response_handler.validate_single_delete_response, self.delete_result_list, HTTPStatus.NO_CONTENT),
            (self.response_handler.validate_multi_get_sessions_response, self.get_result_list)
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
