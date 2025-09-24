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

# To support forward references in type hinting
from __future__ import annotations
from dataclasses import dataclass

from common.defs import BBException, JsonDict
from common.log import logger

from .defs import HSM_ARCH_STRINGS

@dataclass(frozen=True)
class ComputeNode:
    """
    A compute node and its hardware architecture
    """
    xname: str
    arch: str
    hsm_arch: str

    @classmethod
    def from_hsm_node_component(cls, hsm_node_component: JsonDict) -> ComputeNode:
        """
        Return a ComputeNode based on the specified HSM node component
        """
        try:
            xname = hsm_node_component["ID"]
            hsm_arch = hsm_node_component["Arch"]
        except KeyError as exc:
            logger.exception("HSM component has no '%s' field: %s", exc, hsm_node_component)
            raise BBException() from exc
        if not xname:
            logger.error("HSM component has empty or null 'ID' field: %s", hsm_node_component)
            raise BBException()
        if not hsm_arch:
            logger.error("HSM component has empty or null 'Arch' field: %s", hsm_node_component)
            raise BBException()

        for arch_string, hsm_arch_strings in HSM_ARCH_STRINGS.items():
            if hsm_arch in hsm_arch_strings:
                return cls(xname=xname, arch=arch_string, hsm_arch=hsm_arch)

        logger.error("Node '%s' has unsupported 'arch' in HSM: '%s'", xname, hsm_arch)
        raise BBException()
