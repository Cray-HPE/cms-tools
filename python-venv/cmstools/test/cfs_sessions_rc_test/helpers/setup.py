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
CFS race condition setup related functions
"""
from typing import Optional

from cmstools.lib.k8s import get_deployment_replicas, set_deployment_replicas, check_replicas_and_pods_scaled
from cmstools.lib.cfs import CFS_OPERATOR_DEPLOYMENT
from cmstools.test.cfs_sessions_rc_test.cfs.options import get_cfs_page_size, set_cfs_page_size
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs, CFSRCException
from cmstools.test.cfs_sessions_rc_test.cfs.session import cfs_session_exists, delete_cfs_sessions
from cmstools.test.cfs_sessions_rc_test.defs import TestSetupResponse


def _calculate_v2_page_size(page_size: Optional[int], max_sessions: int, current: int) -> int:
    """
    Calculate and return the appropriate page size value for CFS v2.
    When using CFS v2, GET requests will fail if they would return more values than
    the global page-size. These failures would prevent the test from working.
    In the case of CFS v3, it is possible to specify a page-size override along with the
    GET request, so the global value does not cause us issues in that scenario. This is not
    supported in CFS v2.

    There are two cases:

    Case 1: <page_size> is None:
    This means that no page size value was specified by the user as an argument to the test.
    In this case, we just want to return a value that will definitely be higher than we need.
    10 * <max_sessions> should be higher than we need, but if the current global value is already
    higher than that, that's okay too.
    So the appropriate value is either <current> or 10 * <max_sessions>, whichever is higher.

    Case 2: <page_size> is an integer.
    This means that the user has called the test with a specific page size value.
    In this case, we are not really calculating the value, but instead we are validating to make sure
    that the value that they specified will work. The minimum possible page size that will work
    is <max_sessions>, because we are potentially going to be creating that many sessions, and
    listing them.
    So if <page_size> is greater than or equal to <max_sessions>, then <page_size> is the appropriate
    value.
    Otherwise (if <page_size> is less than <max_sessions>) raise an error. In this case, the user
    can address this by calling the test and not specifying a page_size, or specifying a higher
    page_size, or specifying a lower max_sessions value.
    """
    if page_size is None:
        return max(10 * max_sessions, current)

    if page_size < max_sessions:
        error_msg = (
            f"Specified page_size ({page_size}) is less than max_sessions ({max_sessions}). "
            f"For API version v2, Page size must be at least equal to max_sessions for the test to work correctly. "
            f"Either increase page_size, decrease max_sessions, or omit page_size to use auto-calculated value."
        )
        logger.error(error_msg)
        raise CFSRCException()

    return page_size


def _calculate_v3_page_size(page_size: Optional[int], max_sessions: int) -> int:
    """
    Calculate and return the appropriate page size value for CFS v3.
    If <page_size> is None, the appropriate page size is 10 * <max_sessions>.
    Otherwise, the appropriate page size is <page_size>
    """
    return page_size if page_size is not None else 10 * max_sessions


def set_page_size_v2(page_size: Optional[int], max_sessions: int) -> tuple[int, Optional[int]]:
    """
    Retrieves the current value of the CFS global page-size option.
    Calculates the desired global page-size value, based on the input arguments to this function.
    If the current value matches the desired value, then: return current_value, None
    If the current value does not match the desired value, then set the CFS global page-size
    option to the desired value, and return: desired_value, previous value (before we changed it).

    Explanation of return tuple:
    The first part of the tuple is always the value of CFS global page-size option after this function has run.
    The second part of the tuple is the value of the CFS global page-size option before this function was run,
    if this function changed it, otherwise the second value is None.
    """
    current_page_size = get_cfs_page_size()
    new_page_size = _calculate_v2_page_size(page_size, max_sessions, current_page_size)
    if new_page_size != current_page_size:
        set_cfs_page_size(new_page_size)
        logger.info("For CFS v2 API setting the page size %d", new_page_size)
        return new_page_size, current_page_size
    return new_page_size, None


def cfs_sessions_rc_test_setup(script_args: ScriptArgs) -> TestSetupResponse:
    """
    Perform any setup needed for the CFS sessions race condition test.
    Check the number of cfs-operator instances and, if it is non-0, remember the current value and scale it down to 0.
    This will ensure the sessions we create remain in pending state. This must be scaled back to its original value
    when the test exits.
    Check for previous test sessions
    If --delete-previous-sessions is true, then delete all pending state sessions with the text prefix in their names
    (if any), and verify that they are gone.
    If --delete-previous-sessions is not true, then list all pending state sessions with the text prefix in their names.
     If there are any, exit with an error, telling the user to delete these sessions, run with a different name string,
     or run with --delete-previous-sessions
    """
    current_page_size: Optional[int] = None
    new_page_size: int = script_args.page_size
    # If running CFS v2, set the V3 global CFS page-size option to <page-size>
    if script_args.cfs_version == "v2":
        new_page_size, current_page_size = set_page_size_v2(page_size=script_args.page_size, max_sessions=script_args.max_cfs_sessions)

    #  If running CFS v3, calculate the appropriate page size
    elif script_args.cfs_version == "v3":
        current_page_size = None
        new_page_size = _calculate_v3_page_size(page_size=script_args.page_size, max_sessions=script_args.max_cfs_sessions)

    # First check for deleting pre-existing sessions if requested
    logger.info("Checking for pre-existing CFS sessions with name prefix %s", script_args.cfs_session_name)

    if not script_args.delete_preexisting_cfs_sessions:
        # check for existing sessions and fail if any found
        if cfs_session_exists(script_args.cfs_session_name, script_args.cfs_version, script_args.page_size):
            logger.error(
                "Pre-existing CFS sessions found with specified name prefix. Use --delete-previous-sessions "
                "to delete them before proceeding.")
            raise CFSRCException()

    # If requested, delete any pre-existing sessions
    delete_cfs_sessions(script_args.cfs_session_name, script_args.cfs_version, script_args.page_size)

    # Get the current number of replicas for the cray-cfs-operator deployment
    current_replicas = get_deployment_replicas(deployment_name=CFS_OPERATOR_DEPLOYMENT)
    logger.info("Current number of replicas for cray-cfs-operator deployment: %d", current_replicas)

    if current_replicas > 0:
        # Scale the cray-cfs-operator deployment to 0 replicas to stop it from processing cfs sessions
        logger.info("Scaling cray-cfs-operator deployment to 0 replicas to stop it from processing CFS sessions")
        set_deployment_replicas(deployment_name=CFS_OPERATOR_DEPLOYMENT, replicas=0)
        check_replicas_and_pods_scaled(deployment_name=CFS_OPERATOR_DEPLOYMENT, expected_replicas=0)

    return TestSetupResponse(
        original_page_size=current_page_size,
        original_replicas=current_replicas,
        new_page_size=new_page_size
    )
