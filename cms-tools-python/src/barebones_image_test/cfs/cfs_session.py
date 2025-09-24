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
CfsSession and CfsSessionStatusFields classes
"""

from dataclasses import dataclass

from typing import ClassVar

from common.api import request_and_check_status
from common.defs import BBException, BB_TEST_RESOURCE_NAME
from barebones_image_test.ims import ImsImage
from common.log import logger
from barebones_image_test.test_session import SessionStatusFields, TestSession

from .cfs_config import CfsConfig
from .defs import CFS_SESSIONS_URL

@dataclass(frozen=True)
class CfsSessionStatusFields(SessionStatusFields):
    """
    Selected CFS session status fields

    .status.session.status (status field inherited from parent class)
    .status.session.succeeded
    """
    succeeded: str|None = None

    def passed(self) -> bool:
        """
        Returns True if session is complete and succeeded.
        Logs an error and raises an exception if the session is complete and failed.
        Returns False otherwise.
        """
        if not self.completed:
            return False
        if self.succeeded == "true":
            return True
        logger.error("%s completed but not successful ('succeeded' session status field = '%s')",
                     self.session.label_and_name, self.succeeded)
        raise BBException()


class CfsSession(TestSession):
    """
    Data about the image customization session in CFS
    """
    base_url: ClassVar[str] = CFS_SESSIONS_URL
    label: ClassVar[str] = "CFS session"

    def __init__(self, base_ims_image: ImsImage, cfs_config: CfsConfig,
                 session_name: str|None=None):
        """
        Create the specified CFS iamge customization session
        """
        if session_name is None:
            super().__init__()
        else:
            super().__init__(session_name)
        logger.info("Creating %s to customize %s with %s", self.label_and_name,
                    base_ims_image.label_and_name, cfs_config.label_and_name)
        cfs_session_create_data = {
            "name": self.name,
            "configuration_name": cfs_config.name,
            "target": {
                "definition": "image",
                "groups": [ { "name": "Compute", "members": [ base_ims_image.name ] } ],
                "image_map": [ {
                    "source_id": base_ims_image.name,
                    "result_name": BB_TEST_RESOURCE_NAME
                } ]
            }
        }
        resp_data = request_and_check_status("post", CFS_SESSIONS_URL,
                                             json=cfs_session_create_data,
                                             expected_status=201, parse_json=True)
        logger.debug("Created %s: %s", self.label_and_name, resp_data)

    @property
    def current_status_fields(self) -> CfsSessionStatusFields:
        """
        Query CFS API about this session and return selected status fields
        """
        status_dict = self.get()["status"]
        try:
            session_status = status_dict["session"]
        except KeyError:
            # In this case, no real status to reported
            return CfsSessionStatusFields(session=self)
        return CfsSessionStatusFields(session=self,
                                      status=session_status.get("status", None),
                                      succeeded=session_status.get("succeeded", None))

    @property
    def result_id(self) -> str:
        """
        .status.artifacts[0].result_id     result_id: ImsImageId|None = None

        # If we get here, then the CFS session has reported successful completion, so
        # we should be able to get the resulting image ID and return it.
        """
        status_dict = self.get()["status"]
        try:
            artifact_list = status_dict["artifacts"]
        except KeyError as exc:
            logger.exception("%s has no 'artifacts' field in its status: %s", self.label_and_name,
                             status_dict)
            raise BBException() from exc

        if len(artifact_list) > 1:
            logger.error("%s should produce exactly one artifact, but multiple listed: %s",
                         self.label_and_name, artifact_list)
            raise BBException()

        if len(artifact_list) == 0:
            logger.error("%s has no artifacts recorded in its status: %s", self.label_and_name,
                         status_dict)
            raise BBException()

        artifact = artifact_list[0]
        try:
            result_id = artifact["result_id"]
        except KeyError as exc:
            logger.exception("%s: artifact is missing 'result_id' field: %s", self.label_and_name,
                             status_dict)
            raise BBException() from exc

        if result_id:
            return result_id

        logger.error("%s: artifact has blank 'result_id' field: %s", self.label_and_name,
                     status_dict)
        raise BBException()
