#
# MIT License
#
# (C) Copyright 2021-2022, 2024-2025 Hewlett Packard Enterprise Development LP
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
ComputeNode class and related functions
"""

from .api import API_BASE_URL
from .defs import ARM_ARCH, X86_ARCH

# HSM URLs
HSM_URL = f"{API_BASE_URL}/smd/hsm/v2"
HSM_COMP_STATE_URL = f"{HSM_URL}/State/Components"

# The strings HSM uses to identify node arch
HSM_ARM_ARCH = "ARM"
HSM_X86_ARCH = "X86"
HSM_UNKNOWN_ARCH = "UNKNOWN"
# For backwards compatability reasons, nodes with Unknown architecture in HSM are considered to be
# X86_64 architecture
HSM_ARCH_STRINGS = {
    ARM_ARCH: [HSM_ARM_ARCH],
    X86_ARCH: [HSM_X86_ARCH, HSM_UNKNOWN_ARCH]}
