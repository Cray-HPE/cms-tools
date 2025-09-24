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
CfsConfig and CfsConfigLayerData classes
"""

# To support forward references in type hinting
from __future__ import annotations
from dataclasses import dataclass

from typing import ClassVar

from common.api import request, request_and_check_status
from common.defs import BBException
from common.log import logger
from barebones_image_test.test_resource import TestResource

from .defs import CFS_CONFIGS_URL

@dataclass(frozen=True)
class CfsConfigLayerData:
    """
    Data to define a CFS configuration layer
    """
    vcs_url: str
    git_commit: str
    playbook: str


class CfsConfig(TestResource):
    """
    CFS configuration
    """
    base_url: ClassVar[str] = CFS_CONFIGS_URL
    label: ClassVar[str] = "CFS configuration"

    @classmethod
    def create_in_cfs(cls, layer_data: CfsConfigLayerData, config_name: str|None=None) -> CfsConfig:
        """
        Create the configuration in CFS with the specified layer data
        """
        if config_name is None:
            new_config = cls()
        else:
            new_config = cls(name=config_name)
        logger.debug("Creating %s with layer data: %s", new_config.label_and_name, layer_data)
        logger.info("Creating %s", new_config.label_and_name)

        cfs_config_json = {
            "layers": [
                {
                    "clone_url": layer_data.vcs_url,
                    "commit": layer_data.git_commit,
                    "playbook": layer_data.playbook,
                    "name": "compute"
                }
            ]
        }
        resp_data = request_and_check_status("put", new_config.url, expected_status=200,
                                             parse_json=True, json=cfs_config_json)
        logger.debug("Created %s: %s", new_config.label_and_name, resp_data)
        return new_config

    @property
    def exists(self) -> bool:
        """
        Returns True if the config exists in CFS. Returns False if it does not.
        """
        resp = request("get", self.url)
        if resp.status_code == 200:
            logger.debug("%s exists: %s", self.label_and_name, resp.text)
            return True
        if resp.status_code == 404:
            logger.debug("%s does not exist", self.label_and_name)
            return False
        logger.error("Unexpected return code %d from GET query to %s: %s", self.url,
                     resp.status_code, resp.text)
        raise BBException()
