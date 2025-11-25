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
CFS configuration related functions
"""

from cmstools.lib.api import request_and_check_status
from cmstools.lib.cfs import CFS_CONFIGS_URL
from cmstools.lib.common_logger import logger
from cmstools.lib.defs import JsonDict

def create_cfs_config(config_name: str, layers: list[JsonDict]) -> JsonDict:
    """
    Create a CFS configuration using V3 API with the specified name and layers.

    Args:
        config_name: Name of the configuration to create
        layers: List of layer dictionaries containing clone_url, commit, playbook, name

    Returns:
        Dictionary containing the created configuration data
    """
    url = f"{CFS_CONFIGS_URL}/{config_name}"
    cfs_config_json = {"layers": layers}

    resp_data = request_and_check_status("put", url, expected_status=200,
                                         parse_json=True, json=cfs_config_json)
    logger.debug("Created CFS config %s: %s", config_name, resp_data)
    return resp_data
