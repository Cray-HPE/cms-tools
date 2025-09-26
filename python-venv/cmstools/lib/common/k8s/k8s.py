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
Kubernetes module for barebones boot test
"""

from kubernetes import client, config

from cmstools.lib.common.defs import TestException, JsonDict
from cmstools.lib.common.common_logger import logger


def get_k8s_configmap_data(cm_name: str, cm_namespace: str = "default") -> JsonDict:
    """
    Get the specified config map from Kubernetes and return its data field.
    Raise an exception if any problems are encountered.
    """
    try:
        return k8s_client_api.read_namespaced_config_map(name=cm_name, namespace=cm_namespace).data
    except Exception as exc:
        logger.exception("Error retrieving Kubernetes configmap '%s' from namespace '%s'",
                         cm_name, cm_namespace)
        raise TestException from exc


def get_k8s_secret_data(secret_name: str, secret_namespace: str = "default") -> JsonDict:
    """
    Get the specified secret from Kubernetes and return its data field.
    Raise an exception if any problems are encountered.
    """
    try:
        return k8s_client_api.read_namespaced_secret(secret_name, secret_namespace).data
    except Exception as exc:
        logger.exception("Error retrieving Kubernetes secret '%s' from namespace '%s'",
                         secret_name, secret_namespace)
        raise TestException from exc


# initialize k8s
config.load_kube_config()
k8s_client_api = client.CoreV1Api()
