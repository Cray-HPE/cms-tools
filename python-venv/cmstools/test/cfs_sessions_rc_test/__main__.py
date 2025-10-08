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
import re

from cmstools.test.cfs_sessions_rc_test.cfs.cleanup import cleanup_and_restore
from cmstools.test.cfs_sessions_rc_test.cfs.setup import get_cfs_config_name
from cmstools.test.cfs_sessions_rc_test.cfs.setup import cfs_sessions_rc_test_setup
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_sessions_multi_delete_test import cfs_sessions_multi_delete_test
from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.lib.log import LOG_FILE_PATH

from .log import logger
from .defs import ScriptArgs

# Subtests
CFS_SESSIONS_RC_MULTI_DELETE = "multi_delete"
CFS_SESSIONS_RC_MULTI_CREATE_MULTI_DELETE = "multi_create_multi_delete"

subtest_functions_dict = {
    CFS_SESSIONS_RC_MULTI_DELETE: cfs_sessions_multi_delete_test
}

def get_test_names(script_args: ScriptArgs) -> list[str] | None:
    """
    Get the list of subtests to run or skip based on command line arguments.
    If neither is specified, return None to indicate all tests should be run.
    """
    all_tests = list(subtest_functions_dict.keys())
    if script_args.run_subtests:
        return script_args.run_subtests

    if script_args.skip_subtests:
        return [name for name in all_tests if name not in script_args.skip_subtests]

    return all_tests  # Run all tests

def run(script_args: ScriptArgs):
    """
    CFS sessions race condition test main processing
    """
    orig_page_size = None
    orig_replica_count = 0

    try:
        # Setting up the test environment
        test_setup_response = cfs_sessions_rc_test_setup(script_args)
        script_args = script_args._replace(page_size=test_setup_response.new_page_size)
        logger.info(f"Using page size of {script_args.page_size} for CFS API calls")
        orig_page_size = test_setup_response.original_page_size
        orig_replica_count = test_setup_response.original_replicas_count

        # Run the tests
        test_names = get_test_names(script_args)
        if test_names:
            for test in test_names:
                logger.info(f"Starting subtest {test}")
                subtest_function = subtest_functions_dict.get(test)
                if subtest_function:
                    subtest_function(script_args)
                else:
                    logger.error(f"Subtest function for {test} not found")
                    raise CFSRCException()
                logger.info(f"Completed subtest {test}")
        else:
            logger.warning("No subtests to run")

    except CFSRCException as _:
        raise
    except Exception as _:
        raise
    finally:
        cleanup_and_restore(orig_replicas_count=orig_replica_count,
                            orig_page_size=orig_page_size,
                            config_name=get_cfs_config_name(),
                            name_prefix=script_args.cfs_session_name,
                            cfs_version=script_args.cfs_version,
                            page_size=script_args.page_size)

# Validations for command line arguments
def check_min_page_size(value) -> int:
    int_value = int(value)
    if int_value < 1:
        raise argparse.ArgumentTypeError("--page-size must be at least 1")
    return int_value

def validate_name(value) -> str:
    pattern = r'^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'
    # Script appends a number to the name, so allow up to 40 characters here
    if not (1 <= len(value) <= 40):
        raise argparse.ArgumentTypeError("--name must be between 1 and 40 characters")
    if not re.match(pattern, value):
        raise argparse.ArgumentTypeError("--name must match pattern: " + pattern)
    return value

def check_minimum_max_sessions(value) -> int:
    int_value = int(value)
    if int_value < 1:
        raise argparse.ArgumentTypeError("--max-sessions must be at least 1")
    return int_value

def check_minimum_max_delete_reqs(value) -> int:
    int_value = int(value)
    if int_value < 1:
        raise argparse.ArgumentTypeError("--max-multi-delete-reqs must be at least 1")
    return int_value

def check_subtest_names(value) -> list[str]:
    names = [name.strip() for name in value.split(",")]
    invalid_names = [name for name in names if name not in subtest_functions_dict.keys()]
    if invalid_names:
        raise argparse.ArgumentTypeError(f"Invalid subtest names: {', '.join(invalid_names)}")
    return names

def parse_command_line() -> ScriptArgs:
    """
    Parse the command line arguments
    --name	All sessions used for this test will have names with this prefix.
    It should default to one that is not likely to be used by a real customer. Something like cfs-race-condition-test-.
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
    subtest_group = parser.add_mutually_exclusive_group()
    subtest_group.add_argument(
        "--run-subtests",
        type=check_subtest_names,
        help="Comma-separated list of subtests to run. Mutually exclusive with --skip-subtests. If neither is specified, all subtests are run."
    )
    subtest_group.add_argument(
        "--skip-subtests",
        type=check_subtest_names,
        help="Comma-separated list of subtests to skip. Mutually exclusive with --run-subtests. If neither is specified, all subtests are run."
    )
    parser.add_argument(
        "--name",
        type=validate_name,
        default="cfs-race-condition-test-",
        help="Prefix for all session names (default: %(default)s)"
    )
    parser.add_argument(
        "--max-sessions",
        type=check_minimum_max_sessions,
        default=20,
        help="Maximum number of CFS sessions to create (default: %(default)s)"
    )
    parser.add_argument(
        "--max-multi-delete-reqs",
        type=check_minimum_max_delete_reqs,
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
        type=check_min_page_size,
        default=None,
        help="Page size for multi-get requests (default: 10 * max-sessions for v3, min=max-sessions for v2, min=1 for v3)"
    )

    args = parser.parse_args()

    return ScriptArgs(
        cfs_session_name=args.name,
        max_cfs_sessions=args.max_sessions,
        max_multi_cfs_sessions_delete_requests=args.max_multi_delete_reqs,
        delete_preexisting_cfs_sessions=args.delete_previous_sessions,
        cfs_version=args.cfs_version,
        page_size=args.page_size,
        run_subtests=args.run_subtests,
        skip_subtests=args.skip_subtests
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
