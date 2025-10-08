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
CFS race condition test related definitions
"""

from typing import NamedTuple


class ScriptArgs(NamedTuple):
    """
    Encapsulates the command line arguments
    """
    cfs_session_name: str # prefix for cfs session names to "cfs-race-condition-test-"
    max_cfs_sessions: int # default to 20
    max_multi_cfs_sessions_delete_requests: int # default to 4
    delete_preexisting_cfs_sessions: bool
    cfs_version: str # default to v3
    page_size: int
    run_subtests: list[str] | None = None
    skip_subtests: list[str] | None = None


class TestSetupResponse:
    """
    Encapsulates the response from cfs_sessions_rc_test_setup()
    """
    def __init__(self, original_page_size: int | None, original_replicas: int | None, new_page_size: int | None):
        self.original_page_size = original_page_size
        self.original_replicas_count = original_replicas
        self.new_page_size = new_page_size
