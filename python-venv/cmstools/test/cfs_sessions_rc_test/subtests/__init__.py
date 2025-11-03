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

from .cfs_session_base import all_subtests, CFSSessionBase
# Import all subtests below, so that their classes are defined and added to the all_subtests list
from .cfs_sessions_multi_delete_multi_get_test import CFSSessionMultiDeleteMultiGetTest
from .cfs_sessions_multi_delete_test import CfsSessionMultiDeleteTest
from .cfs_sessions_multi_delete_single_get_test import CFSSessionMultiDeleteSingleGetTest
from .cfs_sessions_single_delete_test import CFSSessionSingleDeleteTest
from .cfs_sessions_single_delete_single_get_test import CFSSessionSingleDeleteSingleGetTest
from .cfs_sessions_single_delete_multi_get_test import CFSSessionSingleDeleteMultiGetTest

