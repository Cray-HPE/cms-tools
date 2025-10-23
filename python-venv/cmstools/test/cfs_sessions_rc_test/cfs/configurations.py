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
from http import HTTPStatus

from cmstools.lib.cfs import (CFS_CONFIGS_URL, create_cfs_config)
from cmstools.lib.api import request
from cmstools.test.cfs_sessions_rc_test.defs import CFSRCException
from cmstools.test.cfs_sessions_rc_test.log import logger

DEFAULT_PLAYBOOK = "compute_nodes.yml"


def delete_cfs_configuration(cfs_configuration_name: str) -> None:
    """
    Deletes the specified CFS configuration
    """
    url = f"{CFS_CONFIGS_URL}/{cfs_configuration_name}"
    resp = request("delete", url)

    if resp.status_code != HTTPStatus.NO_CONTENT:
        logger.error("Failed to delete CFS configuration %s: %s", cfs_configuration_name, resp.text)
        raise CFSRCException()


def find_or_create_cfs_config(name_prefix: str) -> tuple[str, bool]:
    url = CFS_CONFIGS_URL
    resp = request("get", url)

    if resp.status_code != HTTPStatus.OK:
        logger.error("Failed to list CFS configs: %s", resp.text)
        raise CFSRCException()

    configs = resp.json()
    configurations_data = configs["configurations"]

    if configurations_data:
        config_name = configurations_data[0]["name"]
        logger.info("Using existing CFS config: %s", config_name)
        return config_name, False

    # Create a new config
    config_name = f"{name_prefix}config"

    # Using dummy values for clone_url and commit
    cfs_config_list = [
        {
            "clone_url": "https://dummy-server-nmn.local/vcs/cray/example-repo.git",
            "commit": "43ecfa8236bed625b54325ebb70916f599999999",
            "playbook": DEFAULT_PLAYBOOK,
            "name": "compute"
        }
    ]

    resp_data = create_cfs_config(config_name=config_name, layers=cfs_config_list)
    logger.debug("Created %s: %s", config_name, resp_data)
    return config_name, True
