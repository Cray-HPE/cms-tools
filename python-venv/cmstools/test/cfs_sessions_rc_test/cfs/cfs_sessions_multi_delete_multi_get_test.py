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
CFS race condition multi delete multi get test related functions
"""

from cmstools.test.cfs_sessions_rc_test.defs import ScriptArgs
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session_deleter import CfsSessionDeleter
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session_creator import CfsSessionCreator

def cfs_sessions_multi_delete_multi_get_test(script_args: ScriptArgs) -> None:
    """
    Create <max-multi-delete-reqs> parallel mutli-delete requests, each of which is deleting all pending
    sessions that have the text prefix string in their names.
    Issue multiple parallel get requests to get all sessions that have the text prefix string in their names.
    For the session lists that were returned by the multi-get requests, validate that every entry in each list
    is a dict object that corresponds to one of the sessions we created.
    (It is fine if some sessions are not listed in any of the responses. It is fine if no sessions are listed
    in any of the responses.)
    """
    cfs_session_creator = CfsSessionCreator(
        script_args=script_args
    )
    cfs_sessions_list = cfs_session_creator.create_sessions()
    # Now issue the specified number of parallel multi-delete requests to delete all sessions
    cfs_session_deleter = CfsSessionDeleter(
        script_args=script_args,
        cfs_session_name_list=cfs_sessions_list
    )
    cfs_session_deleter.multi_delete_multi_get_sessions()
    logger.info("All CFS sessions successfully deleted")
