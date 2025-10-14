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
CFS race condition cleanup related functions
"""

from cmstools.lib.k8s import get_deployment_replicas, set_deployment_replicas
from cmstools.lib.cfs.defs import CFS_OPERATOR_DEPLOYMENT
from cmstools.test.cfs_sessions_rc_test.cfs.session import (delete_cfs_session_by_name,
                                                            delete_cfs_sessions, get_cfs_sessions_list)
from cmstools.test.cfs_sessions_rc_test.log import logger
from cmstools.test.cfs_sessions_rc_test.cfs.options import set_cfs_page_size
from cmstools.test.cfs_sessions_rc_test.cfs.configurations import delete_cfs_configuration
from cmstools.test.cfs_sessions_rc_test.defs import CFS_VERSIONS_STR

def cleanup_cfs_sessions(name_prefix: str, cfs_version: CFS_VERSIONS_STR , page_size: int) -> None:
    """
    Cleanup function to delete any remaining CFS sessions with the specified name prefix
    """
    try:
        delete_cfs_sessions(
            cfs_session_name_contains=name_prefix,
            cfs_version=cfs_version,
            limit=page_size
        )
    except Exception as ex:
        logger.error("Failed to delete remaining CFS sessions with name prefix %s: %s", name_prefix, str(ex))
        logger.info("Deleting session one by one")
        sessions = get_cfs_sessions_list(
            cfs_session_name_contains=name_prefix,
            cfs_version=cfs_version,
            limit=page_size
        )
        if not sessions:
            logger.info("No remaining CFS sessions found for cleanup")
            return
        cfs_session_list = [s["name"] for s in sessions]
        for session_name in cfs_session_list:
            try:
                delete_cfs_session_by_name(session_name, cfs_version=cfs_version)
            except Exception as ex2:
                logger.error("Failed to delete CFS session %s: %s", session_name, str(ex2))

def cleanup_and_restore(orig_replicas_count: int, orig_page_size: int | None,
                        config_name: str| None) -> None:
    """
    Cleanup function to restore the cray-cfs-operator deployment and CFS page size
    """

    if orig_replicas_count != get_deployment_replicas(deployment_name=CFS_OPERATOR_DEPLOYMENT):
        logger.info("Restoring cray-cfs-operator deployment to its original number of replicas: %d", orig_replicas_count)
        set_deployment_replicas(deployment_name="cray-cfs-operator", replicas=orig_replicas_count)

    if orig_page_size is not None:
        logger.info("Restoring CFS v3 global page-size option to its original value: %d", orig_page_size)
        set_cfs_page_size(orig_page_size)

    if config_name is not None:
        logger.info("Deleting CFS configuration %s", config_name)
        delete_cfs_configuration(config_name)
