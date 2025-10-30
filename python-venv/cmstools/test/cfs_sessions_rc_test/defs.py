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

from typing import NamedTuple, Literal, Optional, get_args

from cmstools.lib import CmstoolsException

CFSRCException = CmstoolsException

DEFAULT_SESSION_NAME_PREFIX = "cfs-race-condition-test-"
DEFAULT_MAX_SESSIONS = 20
DEFAULT_MAX_PARALLEL_REQUESTS = 4
DEFAULT_CFS_VERSION = "v3"

# Create the Literal from the constants
CfsVersionsStrLiteral = Literal["v2", "v3"]

# Create the frozenset for runtime validation
CFS_VERSIONS_STR = frozenset(get_args(CfsVersionsStrLiteral))
MAX_NAME_LENGTH = 40
MIN_NAME_LENGTH = 1


class ScriptArgs(NamedTuple):
    """
    Encapsulates the command line arguments
    """
    cfs_session_name: str  # prefix for cfs session names to "cfs-race-condition-test-"
    max_cfs_sessions: int  # default to 20
    max_multi_cfs_sessions_delete_requests: int  # default to 4
    max_multi_cfs_sessions_get_requests: int  # default to 4
    max_single_cfs_sessions_get_requests: int  # default to 4
    delete_preexisting_cfs_sessions: bool
    cfs_version: CfsVersionsStrLiteral  # default to v3
    page_size: Optional[int] = None
    run_subtests: Optional[list[str]] = None
    skip_subtests: Optional[list[str]] = None


class TestSetupResponse:
    """
    Encapsulates the response from cfs_sessions_rc_test_setup()
    """
    def __init__(self, original_page_size: Optional[int], original_replicas: Optional[int], new_page_size: Optional[int]):
        self.original_page_size = original_page_size
        self.original_replicas_count = original_replicas
        self.new_page_size = new_page_size


API_REQUEST_TIMEOUT = 30  # seconds
