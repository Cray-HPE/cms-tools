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
from dataclasses import dataclass
from .defs import CFS_COMPONENTS_URL
from barebones_image_test.log import logger
from barebones_image_test.api import request_and_check_status


@dataclass(frozen=True)
class CfsComponentUpdateData:
    """
    Data to update a CFS component
    """
    desired_config: str


class CfsComponents:
    """
    CFS Components
    """
    @classmethod
    def update_cfs_component(cls, cfs_component_name: str, data: CfsComponentUpdateData) -> None:
        """
        Update CFS components
        """
        url = f"{CFS_COMPONENTS_URL}/{cfs_component_name}"
        update_data_json = {
            "desiredConfig": data.desired_config
        }
        _ = request_and_check_status("patch", url, expected_status=200,
                                             parse_json=True, json=update_data_json)
        logger.info(f"Updated CFS component '{cfs_component_name}' with desired config '{data.desired_config}'")