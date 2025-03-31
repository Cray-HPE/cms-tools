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
TestResource class
"""

from abc import ABC
from dataclasses import dataclass
from typing import ClassVar

from barebones_image_test.api import request_and_check_status
from barebones_image_test.defs import BB_TEST_RESOURCE_NAME, JsonObject
from barebones_image_test.log import logger


@dataclass(frozen=True)
class TestResource(ABC):
    """
    An API resource, most likely created by the test
    """
    # Name/ID for an API resource
    name: str = BB_TEST_RESOURCE_NAME

    # A string describing what this resource is (e.g. "BOS session template", "IMS image")
    label: ClassVar[str] = "Test resource"

    # base URL for API calls to this type of resource
    base_url: ClassVar[str] = ""

    @property
    def label_and_name(self) -> str:
        """
        Shortcut to get the label + name of this resource
        """
        return f"{self.label} '{self.name}'"

    @property
    def url(self) -> str:
        """
        Return the API URL for this resource
        """
        return f"{self.base_url}/{self.name}"

    def delete(self) -> None:
        """
        Calls the API to delete this resource
        """
        logger.info("Deleting %s", self.label_and_name)
        request_and_check_status("delete", self.url, expected_status=204, parse_json=False)
        logger.info("Deleted %s", self.label_and_name)

    def get(self, uri: str=None) -> JsonObject:
        """
        Calls the API to get data on this resource
        """
        if uri:
            return request_and_check_status("get", f"{self.url}/{uri}", expected_status=200, parse_json=True)
        return request_and_check_status("get", self.url, expected_status=200, parse_json=True)
