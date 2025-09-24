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
HSM query functions
"""

from common.api import request_and_check_status
from common.defs import BBException
from common.log import logger

from .defs import HSM_ARCH_STRINGS, HSM_COMP_STATE_URL, HSM_UNKNOWN_ARCH
from .compute_node import ComputeNode


TARGET_FILTER = {
    "Type": "Node",
    "Role": "Compute",
    "Enabled": True }


def get_compute_node(node_xname: str) -> ComputeNode:
    """
    Queries HSM for the specified compute node.
    Raises an exception if:
    1) It is not found
    2) It is not type==Node
    3) It is not role==Compute
    4) It is not enabled==true
    5) Its arch is not Arm, X86, or Unknown
    """
    logger.info("Querying HSM for status of '%s' node", node_xname)
    url = f"{HSM_COMP_STATE_URL}/{node_xname}"
    resp_json = request_and_check_status("get", url, expected_status=200, parse_json=True)
    logger.debug("HSM API response: %s", resp_json)
    errors = False
    for field_name, field_value in TARGET_FILTER.items():
        if resp_json[field_name] != field_value:
            logger.error("Target node should have field '%s'='%s' in HSM, but '%s' has '%s'='%s'",
                         field_name, field_value, node_xname, field_name, resp_json[field_name])
            errors = True
    node = ComputeNode.from_hsm_node_component(resp_json)
    if errors:
        raise BBException()
    logger.debug("Found node '%s' in HSM with HSM arch '%s'", node_xname, node.hsm_arch)
    logger.info("Compute node '%s' has architecture '%s'", node.xname, node.arch)
    return node


def find_compute_node(requested_arch: str) -> ComputeNode:
    """
    Find a compute node to use for the boot test. Query HSM to return a list of all enabled compute
    nodes with the specified architecture. Return the first one that is found (subject to
    the caveat in the next paragraph).

    For backwards compatability reasons, nodes with Unknown architecture are considered to be X86_64
    architecture. However, such nodes are only chosen if no other suitable nodes are available.
    """
    logger.info("Querying HSM to find an enabled compute node with '%s' arch", requested_arch)
    # Specify parameters as lists to allow us to specify multiple values (for architecture)
    params = { field_name: [field_value] for field_name, field_value in TARGET_FILTER.items() }

    logger.debug("Filtering for arch: %s", requested_arch)
    params["Arch"] = HSM_ARCH_STRINGS[requested_arch]

    resp_json = request_and_check_status("get", HSM_COMP_STATE_URL, params=params,
                                         expected_status=200, parse_json=True)
    logger.debug("HSM API response: %s", resp_json)

    # sort through to find a compute node to use
    first_node_unknown_arch = None
    for node_comp in resp_json['Components']:
        compute_node=ComputeNode.from_hsm_node_component(node_comp)
        if compute_node.hsm_arch != HSM_UNKNOWN_ARCH:
            # Use this one
            logger.debug("Found node '%s' in HSM with HSM arch '%s'", compute_node.xname,
                         compute_node.hsm_arch)
            logger.info("Found compute node '%s' with architecture '%s'", compute_node.xname,
                        compute_node.arch)
            return compute_node
        if first_node_unknown_arch is None:
            # Remember this Unknown arch node, to use if we cannot find any others
            first_node_unknown_arch=compute_node

    # bail with a failing error if there are no suitable enabled compute nodes present
    if first_node_unknown_arch is None:
        logger.error("No enabled compute nodes found in HSM with %s architecture", requested_arch)
        raise BBException()

    logger.warning("Only suitable enabled compute node found ('%s') has Unknown architecture in "
                   "HSM", first_node_unknown_arch.xname)
    logger.debug("Found node '%s' in HSM with HSM arch '%s'", first_node_unknown_arch.xname,
                 first_node_unknown_arch.hsm_arch)
    return first_node_unknown_arch
