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
BosSession and BosSessionStatusFields classes
"""

from dataclasses import dataclass
from typing import ClassVar

from barebones_image_test.api import request_and_check_status
from barebones_image_test.hsm import ComputeNode
from barebones_image_test.defs import BBException
from barebones_image_test.log import logger
from barebones_image_test.test_session import SessionStatusFields, TestSession

from .bos_template import BosTemplate
from .defs import BOS_SESSIONS_URL

@dataclass(frozen=True)
class BosSessionStatusFields(SessionStatusFields):
    """
    Selected BOS session status fields

    .status.status (status field inherited from parent class)
    .status.error
    """
    error: str|None = None
    error_summary: list|None = None
    percent_failed: int|None = None

    def passed(self) -> bool:
        """
        Returns True if session is complete and succeeded.
        Logs an error and raises an exception if the session is complete and failed.
        Returns False otherwise.
        """
        if not self.completed:
            return False
        if not self.error:
            if self.percent_failed != 0:
                if not self.error_summary:
                    logger.error("%s completed with percent_failed = %.2f but no "
                                 "errors listed in extended session status",
                                 self.session.label_and_name,
                                 self.percent_failed)
                for err in self.error_summary:
                    logger.error("%s completed unsuccessfully with error: %s",
                                 self.session.label_and_name, err)
                raise BBException()

            return True
        logger.error("%s completed unsuccessfully with one or more errors: %s",
                     self.session.label_and_name, self.error)
        raise BBException()


class BosSession(TestSession):
    """
    Data about the barebones reboot session in BOS
    """
    base_url: ClassVar[str] = BOS_SESSIONS_URL
    label: ClassVar[str] = "BOS session"

    def __init__(self, template: BosTemplate, compute_node: ComputeNode,
                 session_name: str|None=None):
        """
        Create the BOS session that attempts to reboot a compute node.
        """
        if session_name is None:
            super().__init__()
        else:
            super().__init__(name=session_name)
        logger.info("Creating %s with %s on node '%s'", self.label_and_name,
                    template.label_and_name, compute_node.xname)

        # put together the session information
        bos_params = {
            "name": self.name,
            "template_name": template.name,
            "limit": compute_node.xname,
            "operation": "reboot" }

        # make the call to BOS to create the session
        resp_data = request_and_check_status("post", BOS_SESSIONS_URL, json=bos_params,
                                             expected_status=201, parse_json=True)
        logger.debug("Created %s: %s", self.label, resp_data)


    @property
    def current_status_fields(self) -> BosSessionStatusFields:
        """
        Query BOS session and return its status fields
        """
        status_dict = self.get()["status"]
        # get BOS session status list
        list_status_dict = self.get(uri="status")
        return BosSessionStatusFields(session=self, status=status_dict["status"],
                                      error=status_dict["error"],
                                      error_summary=list(list_status_dict["error_summary"].keys()),
                                      percent_failed=list_status_dict["percent_failed"])
