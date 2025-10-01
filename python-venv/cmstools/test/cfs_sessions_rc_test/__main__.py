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

import argparse
import sys

from typing import NamedTuple

from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session import cfs_session_exists, delete_cfs_sessions
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_options import get_cfs_page_size, set_cfs_page_size
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_configurations import delete_cfs_configuration
from cmstools.lib.defs import CmstoolsException as CFSRCException, CmstoolsException
from cmstools.lib.k8s import get_deployment_replicas, set_deployment_replicas
from cmstools.lib.cfs.defs import CFS_OPERATOR_DEPLOYMENT
from cmstools.lib.log import LOG_FILE_PATH

from .log import logger
from .cfs.cfs_session_creator import CfsSessionCreator
from .cfs.cfs_session_deleter import CfsSessionDeleter


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

def cleanup_and_restore(current_replicas: int, current_page_size: int | None, config_name: str| None):
    """
    Cleanup function to restore the cray-cfs-operator deployment and CFS page size
    """
    logger.info(f"Restoring cray-cfs-operator deployment to its original number of replicas: {current_replicas}")
    set_deployment_replicas(deployment_name="cray-cfs-operator", replicas=current_replicas)

    if current_page_size is not None:
        logger.info(f"Restoring CFS v3 global page-size option to its original value: {current_page_size}")
        set_cfs_page_size(current_page_size)

    if config_name is not None:
        logger.info(f"Deleting CFS configuration {config_name}")
        delete_cfs_configuration(config_name)

def run(script_args: ScriptArgs):
    """
    CFS sessions race condition test main processing
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
    cfs_config_name = None
    # First check for deleting pre-existing sessions if requested
    logger.info(f"Checking for pre-existing CFS sessions with name prefix {script_args.cfs_session_name}")

    if not script_args.delete_preexisting_cfs_sessions:
        # check for existing sessions and fail if any found
        if cfs_session_exists(script_args.cfs_session_name, script_args.cfs_version):
            logger.error("Pre-existing CFS sessions found with specified name prefix. Use --delete-previous-sessions to delete them before proceeding.")
            raise CFSRCException()

    # If requested, delete any pre-existing sessions
    delete_cfs_sessions(script_args.cfs_session_name, script_args.cfs_version)
    # Get the current number of replicas for the cray-cfs-operator deployment
    current_replicas = get_deployment_replicas(deployment_name=CFS_OPERATOR_DEPLOYMENT)
    logger.info(f"Current number of replicas for cray-cfs-operator deployment: {current_replicas}")
    current_page_size = None

    try:
        # Scale the cray-cfs-operator deployment to 0 replicas to stop it from processing cfs sessions
        logger.info("Scaling cray-cfs-operator deployment to 0 replicas to stop it from processing CFS sessions")
        set_deployment_replicas(deployment_name=CFS_OPERATOR_DEPLOYMENT, replicas=0)

        # If running CFS v2, set the V3 global CFS page-size option to <page-size>
        if script_args.cfs_version == "v2":
            current_page_size = get_cfs_page_size()
            set_cfs_page_size(script_args.page_size)
            logger.info(f"Using CFS v2 API with page size {script_args.page_size}")

        # Create the specified number of CFS sessions
        cfs_session_creator = CfsSessionCreator(
            name_prefix=script_args.cfs_session_name,
            max_sessions=script_args.max_cfs_sessions,
            cfs_version=script_args.cfs_version,
            page_size=script_args.page_size
        )
        cfs_sessions_list, cfs_config_name = cfs_session_creator.create_sessions()

        # Now issue the specified number of parallel multi-delete requests to delete all sessions
        cfs_session_deleter = CfsSessionDeleter(
            name_prefix=script_args.cfs_session_name,
            max_sessions=script_args.max_cfs_sessions,
            max_multi_delete_reqs=script_args.max_multi_cfs_sessions_delete_requests,
            cfs_session_name_list=cfs_sessions_list,
            cfs_version=script_args.cfs_version
        )
        cfs_session_deleter.delete_sessions()
        logger.info("All CFS sessions successfully deleted")
    except CFSRCException as _:
        raise
    except Exception as _:
        raise
    finally:
        # Check if cfs configuration was created, delete it
        # if cfs_config_name contains script_args.cfs_session_name.
        # This means it was created by this script.
        if not script_args.cfs_session_name in cfs_config_name:
            cfs_config_name = None
        cleanup_and_restore(current_replicas=current_replicas, current_page_size=current_page_size, config_name=cfs_config_name)

def parse_command_line() -> ScriptArgs:
    """
    Parse the command line arguments
    --name	All sessions used for this test will have names with this prefix.
    It should default to one that is not likely to be used by a real customer. Something like "cfs-race-condition-test-".
    --max-sessions	Maximum number of CFS sessions to create (maybe start with a default of 20)
    --max-multi-delete-reqs	Maximum number of parallel multi-delete requests (maybe start with a default of 4)
    --delete-previous-sessions	If true, the test will automatically delete any sessions that exist at the
    start of the test that are in pending state and contain the specified name string. If false, if such sessions exist,
    the test will exit in error.
    --cfs-version	Which version of the CFS API to use. Defaults to 3.
    --page-size	The page size to use for multi-get requests (default to 10*<max-sessions>).
    If using CFS v2, the minimum value is <max-sessions>. Otherwise, minimum value of 1
    If running CFS v2, set the V3 global CFS page-size option to <page-size>
    (if it does not already have that value), but the original value should be restored when the test exits

    """
    parser = argparse.ArgumentParser(
         description="CFS Sessions Race Condition Test Script",)
    parser.add_argument(
        "--name",
        type=str,
        default="cfs-race-condition-test-",
        help="Prefix for all session names (default: %(default)s)"
    )
    parser.add_argument(
        "--max-sessions",
        type=int,
        default=20,
        help="Maximum number of CFS sessions to create (default: %(default)s)"
    )
    parser.add_argument(
        "--max-multi-delete-reqs",
        type=int,
        default=4,
        help="Maximum number of parallel multi-delete requests (default: %(default)s)"
    )
    parser.add_argument(
        "--delete-previous-sessions",
        action='store_true',
        help="Delete any existing pending sessions with the specified name prefix"
    )
    parser.add_argument(
        "--cfs-version",
        type=str,
        default="v3",
        choices=["v2", "v3"],
        help="CFS API version to use (default: %(default)s)"
    )
    parser.add_argument(
        "--page-size",
        type=int,
        default=None,
        help="Page size for multi-get requests (default: 10 * max-sessions for v3, min=max-sessions for v2, min=1 for v3)"
    )

    args = parser.parse_args()

    # Set default page size if not provided
    if args.page_size is None:
        args.page_size = 10 * args.max_sessions

    # Enforce minimum page size rules
    if args.cfs_version == "v2":
        if args.page_size < args.max_sessions:
            logger.info(f"For CFS v2, setting --page-size to {args.max_sessions} to match --max-sessions")
            args.page_size = args.max_sessions
    else:
        if args.page_size < 1:
            logger.info("For CFS v3, setting --page-size to 1")
            args.page_size = 1
    return ScriptArgs(
        cfs_session_name=args.name,
        max_cfs_sessions=args.max_sessions,
        max_multi_cfs_sessions_delete_requests=args.max_multi_delete_reqs,
        delete_preexisting_cfs_sessions=args.delete_previous_sessions,
        cfs_version=args.cfs_version,
        page_size=args.page_size
    )


def main():
   logger.info("CFS Sessions Race Condition Test Starting")
   logger.info(f"For complete logs look in the file {LOG_FILE_PATH}")

   try:
       # process any command line inputs
       script_args = parse_command_line()
       run(script_args)
   except CFSRCException:
       logger.error("Failure in cfs sessions race condition test.")
       sys.exit(1)
   except Exception as err:
       logger.exception(f"An unanticipated exception occurred during cfs sessions race condition test: {str(err)};")
       sys.exit(1)

   logger.info("Successfully completed cfs sessions race condition test.")
   sys.exit(0)

if __name__ == "__main__":
   main()
