#
# MIT License
#
# (C) Copyright 2021-2026 Hewlett Packard Enterprise Development LP
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
CFS URL definitions
"""

from typing import get_args

from cmstools.lib.api import API_BASE_URL

from .types import CfsV2SessionsDeleteCode, CfsV3SessionsDeleteCode, CfsVersionInt

# CFS v3 URLs
CFS_URL = f"{API_BASE_URL}/cfs/v3"
CFS_CONFIGS_URL = f"{CFS_URL}/configurations"
CFS_SESSIONS_URL = f"{CFS_URL}/sessions"
CFS_COMPONENTS_URL = f"{CFS_URL}/components"
CFS_OPTIONS_URL = f"{CFS_URL}/options"

# CFS formatted URLs for API version as placeholder
CFS_URL_TEMPLATE = f"{API_BASE_URL}/cfs/{{api_version}}"
CFS_SESSIONS_URL_TEMPLATE = f"{CFS_URL_TEMPLATE}/sessions"

# Deployments
CFS_OPERATOR_DEPLOYMENT = "cray-cfs-operator"

# CFS options
CFS_DEFAULT_PAGE_SIZE = 1000

# CFS versions
# (using get_args allows us to avoid hard-coding the list in the types file and here)
CFS_VERSIONS_INT: frozenset[CfsVersionInt] = frozenset(get_args(CfsVersionInt))

# CFS sessions delete expected status codes
# (using get_args allows us to avoid hard-coding the list in the types file and here)
CFS_V2_SESSIONS_DELETE_CODES: frozenset[CfsV2SessionsDeleteCode] = frozenset(get_args(CfsV2SessionsDeleteCode))
CFS_V3_SESSIONS_DELETE_CODES: frozenset[CfsV3SessionsDeleteCode] = frozenset(get_args(CfsV3SessionsDeleteCode))
