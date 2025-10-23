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
Used to define type hints for CFS related data structures
"""
from dataclasses import dataclass
from typing import Literal, Optional

# There was a CFS v1, but it hasn't been in CSM since CSM 1.0
CFS_VERSION_INT = Literal[ 2, 3 ]

# CFS session operation HTTP return codes
CFS_V2_SESSIONS_DELETE_CODES = Literal[ 204, 400]
CFS_V3_SESSIONS_DELETE_CODES = Literal[ 200, 400]


# Data classes
@dataclass
class BaseRequestResult:
    """Base data class for request result information."""
    status_code: int
    timed_out: bool = False
    error_message: Optional[str] = None


@dataclass
class SessionDeleteResult(BaseRequestResult):
    """Data class to hold session delete result information."""
    session_data: Optional[dict] = None  # Filled for v3 only


@dataclass
class MultiSessionsGetResult(BaseRequestResult):
    """Data class to hold session GET result information."""
    session_data: Optional[list[dict]] = None
