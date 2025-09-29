#
# MIT License
#
# (C) Copyright 2021-2022, 2024-2025 Hewlett Packard Enterprise Development LP
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
TestSession and SessionStatusFields classes
"""

# To support forward references in type hinting
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import asdict, dataclass
import time

from cmstools.lib.defs import CmstoolsException as BBException

from .log import logger
from .test_resource import TestResource


@dataclass(frozen=True)
class SessionStatusFields(ABC):
    """
    Select status fields for BOS and CFS sessions
    """
    session: TestSession
    status: str|None = None

    @property
    def started(self) -> bool:
        """
        Returns true if the session status is set and not pending
        """
        return self.status and self.status != "pending"

    @property
    def completed(self) -> bool:
        """
        Returns true if the status field is "complete"
        """
        return self.status == "complete"

    @abstractmethod
    def passed(self) -> bool:
        """
        Returns True if session is complete and succeeded.
        Logs an error and raises an exception if the session is complete and failed.
        Returns False otherwise.
        """

class TestSession(TestResource, ABC):
    """
    For BOS and CFS sessions
    """
    @property
    @abstractmethod
    def current_status_fields(self) -> SessionStatusFields:
        """
        Queries the API and returns the current session status fields
        """

    def __log_field_changes(self, field_name: str, new_value, old_value) -> None:
        """
        For the specified field, log if it has changed
        """
        if new_value == old_value:
            return
        if new_value is None:
            logger.info("%s '%s' field is now null", self.label_and_name, field_name)
        else:
            logger.info("%s '%s' field is now '%s'", self.label_and_name, field_name, new_value)

    def __log_status_changes(self, new_status: SessionStatusFields,
                             old_status: SessionStatusFields|None=None) -> None:
        """
        Compare the new status fields to the previous ones and report on changes values.
        If no previous values specified, then instead it will report any values which are
        set to non-null values.
        """
        # For the initial call to this, we just report the current status field value
        if old_status is None:
            if new_status.status is None:
                logger.info("%s 'status' field is null", self.label_and_name)
            else:
                logger.info("%s 'status' field is '%s'", self.label_and_name, new_status.status)
            return
        self.__log_field_changes("status", new_status.status, old_status.status)
        for field_name, new_value in asdict(new_status).items():
            if field_name in { "status", "session" }:
                # Already covered status, and we know it is the same session
                continue
            self.__log_field_changes(field_name, new_value, getattr(old_status, field_name))

    def wait_for_session_to_complete(self, wait_seconds_between_queries: int = 10,
                                     timeout_minutes_if_not_started: int = 10,
                                     overall_timeout_minutes: int = 30) -> None:
        """
        Check the status of the session periodically until the session is completed, or until
        we time out. Additionally, if a session is not in "running" state before that timeout
        period has elapsed, the test will timeout.
        """
        logger.info("Waiting for %s to complete", self.label_and_name)

        start_time = time.time()
        def overall_timeout_not_reached() -> bool:
            """ Returns True if we have not reached the overall timeout; False otherwise """
            return time.time() <= start_time + overall_timeout_minutes*60

        def pending_timeout_not_reached() -> bool:
            """ Returns True if we have not reached the pending timeout; False otherwise """
            return time.time() <= start_time + timeout_minutes_if_not_started*60

        started = False
        current_fields = self.current_status_fields
        self.__log_status_changes(current_fields)
        last_fields = current_fields
        if current_fields.passed():
            logger.info("%s completed successfully", self.label_and_name)
            return
        while overall_timeout_not_reached():
            time.sleep(wait_seconds_between_queries)
            current_fields = self.current_status_fields
            self.__log_status_changes(current_fields, last_fields)
            last_fields = current_fields
            if current_fields.passed():
                logger.info("%s completed successfully", self.label_and_name)
                return
            if started or current_fields.started:
                started = True
                continue
            # If we are not in running or completed state, time out if the pending time limit has
            # been exceeded
            if pending_timeout_not_reached():
                continue
            logger.error("%s not in 'running' or 'complete' state even after %d minutes",
                         self.label_and_name, timeout_minutes_if_not_started)
            raise BBException()

        logger.error("%s has not completed even after %d minutes", self.label_and_name,
                     overall_timeout_minutes)
        raise BBException()
