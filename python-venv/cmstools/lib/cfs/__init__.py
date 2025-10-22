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
CFS module for cmstools tests
"""

from .defs import (CFS_SESSIONS_URL_TEMPLATE, CFS_OPERATOR_DEPLOYMENT, CFS_OPTIONS_URL,
                   CFS_DEFAULT_PAGE_SIZE, CFS_CONFIGS_URL)
from .types import (SessionDeleteResult, MultiSessionsGetResult, CFS_V2_SESSIONS_DELETE_CODES,
                    CFS_V3_SESSIONS_DELETE_CODES, HTTP_OK, HTTP_NO_CONTENT, HTTP_BAD_REQUEST, HTTP_NOT_FOUND,
                    HTTP_CREATED)
from .config import create_cfs_config
