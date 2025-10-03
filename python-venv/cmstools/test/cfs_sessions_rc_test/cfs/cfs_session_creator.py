#
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
cfs session creator class for session creation and verification
"""

from cmstools.lib.api import request, request_and_check_status
from cmstools.lib.cfs.defs import CFS_CONFIGS_URL, CFS_SESSIONS_URL_TEMPLATE
from cmstools.lib.defs import CmstoolsException as CFSRCException
from cmstools.test.cfs_sessions_rc_test.cfs.cfs_session import get_cfs_sessions_list
from cmstools.test.cfs_sessions_rc_test.log import logger

DEFAULT_PLAYBOOK = "compute_nodes.yml"

class CfsSessionCreator:
    def __init__(self, name_prefix: str, max_sessions: int, cfs_version: str, page_size: int):
        self.name_prefix = name_prefix
        self.max_sessions = max_sessions
        self.cfs_version = cfs_version
        self.page_size = page_size

    def _find_or_create_cfs_config(self) -> str:
        url = CFS_CONFIGS_URL
        resp = request("get", url)

        if resp.status_code != 200:
            logger.error(f"Failed to list CFS configs: {resp.text}")
            raise CFSRCException()

        configs = resp.json()
        configurations_data = configs["configurations"]

        if configs:
            config_name = configurations_data[0]["name"]
            logger.info(f"Using existing CFS config: {config_name}")
            return config_name

        # Create a new config
        config_name = f"{self.name_prefix}config"
        url = f"{CFS_CONFIGS_URL}/{config_name}"
        # Using dummy values for clone_url and commit
        cfs_config_json = {
            "layers": [
                {
                    "clone_url": "https://dummy-server-nmn.local/vcs/cray/example-repo.git",
                    "commit": "43ecfa8236bed625b54325ebb70916f599999999",
                    "playbook": DEFAULT_PLAYBOOK,
                    "name": "compute"
                }
            ]
        }
        resp_data = request_and_check_status("put", url, expected_status=200,
                                             parse_json=True, json=cfs_config_json)
        logger.debug(f"Created {config_name}: {resp_data}")
        return config_name

    @property
    def expected_http_status(self) -> int:
        if self.cfs_version == "v2":
            return 200
        return 201

    def create_cfs_session_payload(self, session_name: str, config_name: str) -> dict:
        if self.cfs_version == "v3":
            return {
            "name": session_name,
            "configuration_name": config_name,
            "target": {
                "definition": "spec",
                "groups": [ { "name": "Compute", "members": [ "fakexname" ] } ],
            }
        }

        return {
            "name": session_name,
            "configurationName": config_name,
            "target": {
                "definition": "spec",
                "groups": [ { "name": "Compute", "members": [ "fakexname" ] } ],
            }
        }

    def create_sessions(self) -> (list[str], str):
        """
        Create CFS sessions up to max_sessions using the specified name prefix.
        List all sessions in pending state that have the text prefix string in their names. Verify that the names of
        these sessions match the ones we just created.
        """
        config_name = self._find_or_create_cfs_config()
        cfs_session_names_list = []
        url = CFS_SESSIONS_URL_TEMPLATE.format(api_version=self.cfs_version)
        for i in range(self.max_sessions):
            session_name = f"{self.name_prefix}{i}"
            session_payload = self.create_cfs_session_payload(session_name=session_name, config_name=config_name)
            _ = request_and_check_status("post", url,
                                                 json=session_payload,
                                                 expected_status=self.expected_http_status, parse_json=True)
            cfs_session_names_list.append(session_name)
            logger.info(f"Created CFS session: {session_name}")

        # Verify all sessions are in pending state and names match
        sessions = get_cfs_sessions_list(cfs_session_name_contains=self.name_prefix, cfs_version=self.cfs_version, limit=self.page_size)
        pending_cfs_session_names = sorted([s["name"] for s in sessions if s["name"] in cfs_session_names_list])

        if sorted(cfs_session_names_list) != sorted(pending_cfs_session_names):
            logger.error(f"Mismatch in created and pending session names. Created: {cfs_session_names_list}, Pending: {pending_cfs_session_names}")
            raise CFSRCException()

        logger.info(f"All {self.max_sessions} CFS sessions created and in pending state")
        return cfs_session_names_list, config_name