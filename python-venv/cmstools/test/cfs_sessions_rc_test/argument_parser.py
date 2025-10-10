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
Function for parsing command line arguments
"""

import argparse
from typing import get_args

from .validators import (check_subtest_names, check_minimum_max_sessions, check_minimum_max_parallel_reqs
                            , check_min_page_size, validate_name)
from .defs import (DEFAULT_MAX_SESSIONS, DEFAULT_MAX_PARALLEL_REQUESTS, DEFAULT_CFS_VERSION,
                   DEFAULT_SESSION_NAME_PREFIX, CFS_VERSIONS_STR)

def _add_mutually_exclusive_arguments(parser: argparse.ArgumentParser) -> None:
    """
    --run-subtests	A comma-separated list of subtests. Only these subtests will be run. Mutually exclusive with
    --skip-subtests. If neither is specified, all subtests are run.
    --skip-subtests	A comma-separated list of subtests. All subtests will be run EXCEPT for these. Mutually exclusive
     with --run-subtests. If neither is specified, all subtests are run.
    """
    group = parser.add_mutually_exclusive_group()
    group.add_argument(
        "--run-subtests",
        type=check_subtest_names,
        help="Comma-separated list of subtests to run. Mutually exclusive with --skip-subtests. If neither is specified, all subtests are run."
    )
    group.add_argument(
        "--skip-subtests",
        type=check_subtest_names,
        help="Comma-separated list of subtests to skip. Mutually exclusive with --run-subtests. If neither is specified, all subtests are run."
    )

def _add_cfs_session_arguments(parser: argparse.ArgumentParser) -> None:
    """
    Add arguments related to CFS sessions
    --name	All sessions used for this test will have names with this prefix.
    It should default to one that is not likely to be used by a real customer. Something like cfs-race-condition-test-.
    """
    parser.add_argument(
        "--name",
        type=validate_name,
        default=DEFAULT_SESSION_NAME_PREFIX,
        help="Prefix for all session names (default: %(default)s)"
    )

def _add_setup_arguments(parser: argparse.ArgumentParser) -> None:
    """
    Add arguments related to set up
    --delete-previous-sessions	If true, the test will automatically delete any sessions that exist at the
    start of the test that are in pending state and contain the specified name string. If false, if such sessions exist,
    the test will exit in error.
    --cfs-version	Which version of the CFS API to use. Defaults to 3.
    --page-size	The page size to use for multi-get requests (default to 10*<max-sessions>).
    If using CFS v2, the minimum value is <max-sessions>. Otherwise, minimum value of 1
    If running CFS v2, set the V3 global CFS page-size option to <page-size>
    (if it does not already have that value), but the original value should be restored when the test exits
    """
    parser.add_argument(
        "--delete-previous-sessions",
        action='store_true',
        help="Delete any existing pending sessions with the specified name prefix"
    )
    parser.add_argument(
        "--cfs-version",
        type=str,
        default=DEFAULT_CFS_VERSION,
        choices=get_args(CFS_VERSIONS_STR),
        help="CFS API version to use (default: %(default)s)"
    )
    parser.add_argument(
        "--page-size",
        type=check_min_page_size,
        default=None,
        help="Page size for multi-get requests (default: 10 * max-sessions for v3, min=max-sessions for v2, min=1 for v3)"
    )

def _add_subtests_arguments(parser: argparse.ArgumentParser) -> None:
    """
    Add arguments related to subtests
     --max-sessions	Maximum number of CFS sessions to create (maybe start with a default of 20)
    --max-multi-delete-reqs	Maximum number of parallel multi-delete requests (maybe start with a default of 4)
    --max-multi-get-reqs	Maximum number of parallel multi-get requests (maybe start with a default of 4)
    """
    parser.add_argument(
        "--max-sessions",
        type=check_minimum_max_sessions,
        default=DEFAULT_MAX_SESSIONS,
        help="Maximum number of CFS sessions to create (default: %(default)s)"
    )
    parser.add_argument(
        "--max-multi-get-reqs",
        type=check_minimum_max_parallel_reqs,
        default=DEFAULT_MAX_PARALLEL_REQUESTS,
        help="Maximum number of parallel multi-get requests (default: %(default)s)"
    )
    parser.add_argument(
        "--max-multi-delete-reqs",
        type=check_minimum_max_parallel_reqs,
        default=DEFAULT_MAX_PARALLEL_REQUESTS,
        help="Maximum number of parallel multi-delete requests (default: %(default)s)"
    )