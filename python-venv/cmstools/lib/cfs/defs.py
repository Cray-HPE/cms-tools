#
# MIT License
#
# (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP
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

from typing import Literal, Optional
from dataclasses import dataclass

from cmstools.lib.api import API_BASE_URL


# CFS URLs
CFS_URL = f"{API_BASE_URL}/cfs/v3"
CFS_CONFIGS_URL = f"{CFS_URL}/configurations"
CFS_SESSIONS_URL = f"{CFS_URL}/sessions"
CFS_COMPONENTS_URL = f"{CFS_URL}/components"
CFS_OPTIONS_URL = f"{CFS_URL}/options"

# CFS formatted URLs for API version as placeholder
CFS_SESSIONS_URL_TEMPLATE = f"{API_BASE_URL}/cfs/{{api_version}}/sessions"

# Type hinting

# There was a CFS v1, but it hasn't been in CSM since CSM 1.0
CFS_VERSION_INT = Literal[ 2, 3 ]

# CFS session operation HTTP return codes
CFS_V2_SESSION_DELETE_CODES = Literal[ 204, 400 ]
CFS_V3_SESSION_DELETE_CODES = Literal[ 200, 400 ]

# HTTP Status codes
HTTP_OK = 200
HTTP_NO_CONTENT = 204
HTTP_BAD_REQUEST = 400
HTTP_NOT_FOUND = 404


# Deployments
CFS_OPERATOR_DEPLOYMENT = "cray-cfs-operator"

# CFS options
CFS_DEFAULT_PAGE_SIZE = 1000

# Data classes
@dataclass
class SessionDeleteResult:
    """Data class to hold session delete result information."""
    status_code: int
    session_data: Optional[dict] = None # Filled for v3 only
    timed_out: bool = False
    error_message: Optional[str] = None

@dataclass
class SessionGetWithNameContainsResult:
    """Data class to hold session GET result information."""
    status_code: int
    session_data: Optional[list[dict]] = None
    timed_out: bool = False
    error_message: Optional[str] = None

