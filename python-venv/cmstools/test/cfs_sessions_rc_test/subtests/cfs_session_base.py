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
CFS sessions race condition base class and related functions
"""

import threading
from abc import ABC, abstractmethod
import inspect
from typing import ClassVar, Optional

from cmstools.test.cfs_sessions_rc_test.defs import CFSRCException
from cmstools.test.cfs_sessions_rc_test.helpers.cleanup import cleanup_cfs_sessions
from cmstools.test.cfs_sessions_rc_test.cfs.session import cfs_session_exists
from cmstools.test.cfs_sessions_rc_test.cfs.session_creator import CfsSessionCreator
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.log import logger


class CFSSessionBase(ABC):
    """Abstract base class for CFS session race condition tests."""

    _all_subtests: ClassVar[dict[str, type["CFSSessionBase"]]] = {}

    # Actual subtest subclasses must set this class variable to a
    # string value.
    subtest_name: ClassVar[Optional[str]] = None

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        # Add the class to our registry if it meets both of the following criteria:
        # * It is not abstract
        # * The subtest_name class variable has been over-ridden from its null value
        if inspect.isabstract(cls):
            return
        if cls.subtest_name is None:
            return
        # Raise an exception if attempting to re-use an existing subtest name
        assert cls.subtest_name not in cls._all_subtests, f"Re-defining existing subtest: {cls.subtest_name}"
        cls._all_subtests[cls.subtest_name] = cls

    @classmethod
    def get_all_subtests(cls) -> dict[str, type["CFSSessionBase"]]:
        """Return all registered subtest classes."""
        return cls._all_subtests.copy()

    def __init__(self, script_args: ScriptArgs):
        self.script_args = script_args
        self._session_names: list[str] = self._setup()
        self.lock = threading.Lock()

    @abstractmethod
    def _execute_test_logic(self) -> None:
        """Execute the specific test logic. Must be implemented by subclasses."""
        pass

    @abstractmethod
    def _validate_results(self) -> None:
        """Validate test results. Must be implemented by subclasses."""
        pass

    def run(self) -> None:
        """Run the test with setup, execution, validation, and cleanup."""
        try:
            self._execute_test_logic()
            self._validate_results()
        except Exception as err:
            logger.exception("Test execution failed: %s", err)
            raise CFSRCException() from err
        finally:
            self._cleanup()

    @staticmethod
    def execute(script_args: ScriptArgs) -> None:
        """Template method for test execution."""
        pass

    def _setup(self) -> list[str]:
        """Creating CFS sessions needed for tests."""
        return self._create_sessions()

    def _cleanup(self) -> None:
        """CLeanup any sessions with name prefix after testing."""
        if self._session_names:
            self._delete_sessions()

    def _create_sessions(self) -> list[str]:
        """Create CFS sessions for testing."""
        cfs_session_creator = CfsSessionCreator(script_args=self.script_args)
        return cfs_session_creator.create_sessions()

    def _delete_sessions(self) -> None:
        """Delete specified CFS sessions."""
        if cfs_session_exists(
                cfs_session_name_contains=self.script_args.cfs_session_name,
                cfs_version=self.script_args.cfs_version,
                limit=self.script_args.page_size):
            logger.info("Cleaning up any remaining CFS sessions with name prefix %s", self.script_args.cfs_session_name)
            cleanup_cfs_sessions(name_prefix=self.script_args.cfs_session_name,
                                 cfs_version=self.script_args.cfs_version,
                                 page_size=self.script_args.page_size)


def all_subtests() -> dict[str, type[CFSSessionBase]]:
    """Return all registered subtest classes."""
    return CFSSessionBase.get_all_subtests()
