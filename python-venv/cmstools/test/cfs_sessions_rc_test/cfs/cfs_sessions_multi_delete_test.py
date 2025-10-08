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
CFS race condition multi delete related functions
"""

from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session_deleter import CfsSessionDeleter
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session_creator import CfsSessionCreator

def cfs_sessions_multi_delete_test(script_args: ScriptArgs) -> None:
    """
    Create <max-multi-delete-reqs> parallel mutli-delete requests, each of which is deleting all pending
    sessions that have the text prefix string in their names.
    (v3 only) If successful, each delete request will return a list of the session names that it deleted.
    These lists need to be saved. (This is only true for v3 – for v2 a successful response returns nothing)
    Wait for all parallel jobs to complete.
    """
    cfs_session_creator = CfsSessionCreator(
        name_prefix=script_args.cfs_session_name,
        max_sessions=script_args.max_cfs_sessions,
        cfs_version=script_args.cfs_version,
        page_size=script_args.page_size
    )
    cfs_sessions_list = cfs_session_creator.create_sessions()
    # Now issue the specified number of parallel multi-delete requests to delete all sessions
    cfs_session_deleter = CfsSessionDeleter(
        name_prefix=script_args.cfs_session_name,
        max_sessions=script_args.max_cfs_sessions,
        max_multi_delete_reqs=script_args.max_multi_cfs_sessions_delete_requests,
        cfs_session_name_list=cfs_sessions_list,
        page_size=script_args.page_size,
        cfs_version=script_args.cfs_version
    )
    cfs_session_deleter.delete_sessions()
    logger.info("All CFS sessions successfully deleted")

