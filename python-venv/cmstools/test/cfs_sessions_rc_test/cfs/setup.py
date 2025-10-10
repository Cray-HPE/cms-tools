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

from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.lib.k8s import get_deployment_replicas, set_deployment_replicas, check_replicas_and_pods_scaled
from cmstools.lib.cfs.defs import CFS_OPERATOR_DEPLOYMENT
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_options import get_cfs_page_size, set_cfs_page_size
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session import cfs_session_exists, delete_cfs_sessions
from cmstools.test.cfs_sessions_rc_test.defs import TestSetupResponse

cfs_config_name = None
def set_cfs_config_name(config_name: str) -> None:
    """
    Set the CFS configuration to the specified name.
    """
    global cfs_config_name
    cfs_config_name = config_name
    logger.info(f"Using CFS configuration {config_name}")

def get_cfs_config_name() -> str | None:
    """
    Get the CFS configuration name.
    """
    global cfs_config_name
    return cfs_config_name

def set_page_size_if_needed(page_size: int|None, max_sessions: int, cfs_version: str) -> tuple[int, int | None]:
    """
    Set the CFS global page-size option to the desired value if it is not already set to that value for v2.
    Return the current value if it was changed, otherwise return None.
    For v3, if page_size is None, set it to 10 * max_sessions.
    For v2, if page_size is None, set it to 10 * max_sessions, but if that is less than the current value,
    leave it unchanged. If page_size is specified and is less than max_sessions, set it to max_sessions.
    For v2, return the current value if it was changed, otherwise return None.
    """
    current_page_size = None

    if page_size is None:
        page_size = 10 * max_sessions
        if cfs_version == "v2":
            current_page_size = get_cfs_page_size()
            if current_page_size < page_size:
                set_cfs_page_size(page_size)
                logger.info(f"Using CFS v2 API with page size {page_size}")
    else:
        if cfs_version == "v2":
            if page_size < max_sessions:
                logger.info(f"For CFS v2, setting --page-size to {max_sessions} to match --max-sessions")
                page_size = max_sessions
            current_page_size = get_cfs_page_size()
            set_cfs_page_size(page_size)
            logger.info(f"Using CFS v2 API with page size {page_size}")
    return page_size, current_page_size


def cfs_sessions_rc_test_setup(script_args: ScriptArgs ) -> TestSetupResponse:
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
    # If running CFS v2, set the V3 global CFS page-size option to <page-size>
    # Update page_size in script_args by creating a new instance
    new_page_size, current_page_size = set_page_size_if_needed(
        page_size=script_args.page_size,
        max_sessions=script_args.max_cfs_sessions,
        cfs_version=script_args.cfs_version
    )

    # First check for deleting pre-existing sessions if requested
    logger.info(f"Checking for pre-existing CFS sessions with name prefix {script_args.cfs_session_name}")

    if not script_args.delete_preexisting_cfs_sessions:
        # check for existing sessions and fail if any found
        if cfs_session_exists(script_args.cfs_session_name, script_args.cfs_version, script_args.page_size):
            logger.error(
                "Pre-existing CFS sessions found with specified name prefix. Use --delete-previous-sessions to delete them before proceeding.")
            raise CFSRCException()

    # If requested, delete any pre-existing sessions
    delete_cfs_sessions(script_args.cfs_session_name, script_args.cfs_version, script_args.page_size)

    # Get the current number of replicas for the cray-cfs-operator deployment
    current_replicas = get_deployment_replicas(deployment_name=CFS_OPERATOR_DEPLOYMENT)
    logger.info(f"Current number of replicas for cray-cfs-operator deployment: {current_replicas}")

    # Scale the cray-cfs-operator deployment to 0 replicas to stop it from processing cfs sessions
    logger.info("Scaling cray-cfs-operator deployment to 0 replicas to stop it from processing CFS sessions")
    set_deployment_replicas(deployment_name=CFS_OPERATOR_DEPLOYMENT, replicas=0)
    check_replicas_and_pods_scaled(deployment_name=CFS_OPERATOR_DEPLOYMENT, expected_replicas=0)

    return TestSetupResponse(
        original_page_size=current_page_size,
        original_replicas=current_replicas,
        new_page_size=new_page_size
    )
