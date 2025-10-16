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
from typing import NoReturn, Callable

from cmstools.test.cfs_sessions_rc_test.helpers.cleanup import cleanup_and_restore
from cmstools.test.cfs_sessions_rc_test.helpers.setup import get_cfs_config_name
from cmstools.test.cfs_sessions_rc_test.helpers.setup import cfs_sessions_rc_test_setup
from cmstools.test.cfs_sessions_rc_test.subtests.cfs_session_gd_base import CFSSessionGDBase
from cmstools.test.cfs_sessions_rc_test.subtests import cfs_sessions_multi_delete_multi_get_test, \
    cfs_sessions_multi_delete_test
from cmstools.lib.log import LOG_FILE_PATH

from .log import logger
from .defs import ScriptArgs, CFSRCException
from .argument_parser import (_add_mutually_exclusive_arguments, _add_setup_arguments,
                              _add_subtests_arguments, _add_cfs_session_arguments)


# These imports are required to trigger subtest class registration via __init_subclass__
_ = (cfs_sessions_multi_delete_test, cfs_sessions_multi_delete_multi_get_test)

def get_subtest_functions() -> dict[str, Callable[[ScriptArgs], None]]:
    """Get all registered subtest execute functions."""
    return {
        name: cls.execute
        for name, cls in CFSSessionGDBase.get_all_subtests().items()
    }


def get_test_names(script_args: ScriptArgs) -> list[str] | None:
    """Get the list of subtests to run or skip based on command line arguments."""
    all_tests = list(get_subtest_functions().keys())
    logger.info("Available subtests: %s", all_tests)
    if script_args.run_subtests:
        return script_args.run_subtests

    if script_args.skip_subtests:
        return [name for name in all_tests if name not in script_args.skip_subtests]

    return all_tests

def _execute_subtests(test_names: list[str], script_args: ScriptArgs) -> None:
    """Execute the specified subtests."""
    subtest_functions = get_subtest_functions()

    for test_name in test_names:
        logger.info("Starting subtest %s", test_name)
        subtest_function = subtest_functions.get(test_name)

        if not subtest_function:
            logger.error("Subtest function for %s not found", test_name)
            raise CFSRCException()

        subtest_function(script_args)
        logger.info("Completed subtest %s", test_name)

def run(script_args: ScriptArgs) -> None:
    """
    CFS sessions race condition test main processing
    """
    orig_page_size = None
    orig_replica_count = 0

    try:
        # Run the tests
        test_names = get_test_names(script_args)
        if not test_names:
            logger.warning("No subtests to run after applying run/skip filters")
            return

        # Setting up the test environment
        test_setup_response = cfs_sessions_rc_test_setup(script_args)
        script_args = script_args._replace(page_size=test_setup_response.new_page_size)
        logger.info("Using page size of %d for CFS API calls", script_args.page_size)
        orig_page_size = test_setup_response.original_page_size
        orig_replica_count = test_setup_response.original_replicas_count

        _execute_subtests(test_names=test_names, script_args=script_args)

    finally:
        cleanup_and_restore(orig_replicas_count=orig_replica_count,
                            orig_page_size=orig_page_size,
                            config_name=get_cfs_config_name())



def parse_command_line() -> ScriptArgs:
    """
    Parse the command line arguments
    """
    parser = argparse.ArgumentParser(
         description="CFS Sessions Race Condition Test Script",)

    _add_mutually_exclusive_arguments(parser=parser)
    _add_cfs_session_arguments(parser=parser)
    _add_setup_arguments(parser=parser)
    _add_subtests_arguments(parser=parser)

    args = parser.parse_args()

    return ScriptArgs(
        cfs_session_name=args.name,
        max_cfs_sessions=args.max_sessions,
        max_multi_cfs_sessions_delete_requests=args.max_multi_delete_reqs,
        delete_preexisting_cfs_sessions=args.delete_previous_sessions,
        max_multi_cfs_sessions_get_requests=args.max_multi_get_reqs,
        cfs_version=args.cfs_version,
        page_size=args.page_size,
        run_subtests=args.run_subtests,
        skip_subtests=args.skip_subtests
    )


def main() -> NoReturn:
   logger.info("CFS Sessions Race Condition Test Starting")
   logger.info("For complete logs look in the file %s", LOG_FILE_PATH)

   try:
       # process any command line inputs
       script_args = parse_command_line()
       run(script_args)
   except CFSRCException:
       logger.error("Failure in cfs sessions race condition test.")
       sys.exit(1)
   except Exception as err:
       logger.exception("An unanticipated exception occurred during cfs sessions race condition test: %s;", err)
       sys.exit(1)

   logger.info("Successfully completed cfs sessions race condition test.")
   sys.exit(0)

if __name__ == "__main__":
   main()
