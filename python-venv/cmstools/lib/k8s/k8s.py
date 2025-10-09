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
Kubernetes module cmstools test
"""
import time

from kubernetes import client, config

from cmstools.lib.defs import CmstoolsException, JsonDict
from cmstools.lib.common_logger import logger
from cmstools.lib.k8s.defs import (DEFAULT_DEPLOYMENT_NS, DEFAULT_SECRET_NS, DEFAULT_CONFIGMAP_NS)


def get_k8s_configmap_data(cm_name: str, cm_namespace: str = DEFAULT_CONFIGMAP_NS) -> JsonDict:
    """
    Get the specified config map from Kubernetes and return its data field.
    Raise an exception if any problems are encountered.
    """
    try:
        return k8s_client_api.read_namespaced_config_map(name=cm_name, namespace=cm_namespace).data
    except Exception as exc:
        logger.exception("Error retrieving Kubernetes configmap '%s' from namespace '%s'",
                         cm_name, cm_namespace)
        raise CmstoolsException from exc


def get_k8s_secret_data(secret_name: str, secret_namespace: str = DEFAULT_SECRET_NS) -> JsonDict:
    """
    Get the specified secret from Kubernetes and return its data field.
    Raise an exception if any problems are encountered.
    """
    try:
        return k8s_client_api.read_namespaced_secret(secret_name, secret_namespace).data
    except Exception as exc:
        logger.exception("Error retrieving Kubernetes secret '%s' from namespace '%s'",
                         secret_name, secret_namespace)
        raise CmstoolsException from exc

def get_deployment_replicas(deployment_name: str, namespace: str = DEFAULT_DEPLOYMENT_NS) -> int:
    """
    Get the current replica count for the specified deployment.
    """
    try:
        deployment = apps_v1_api.read_namespaced_deployment(name=deployment_name, namespace=namespace)
        return deployment.spec.replicas
    except Exception as exc:
        logger.exception("Error retrieving replica count for deployment '%s' in namespace '%s'",
                         deployment_name, namespace)
        raise CmstoolsException from exc

def set_deployment_replicas(deployment_name: str, replicas: int, namespace: str = DEFAULT_DEPLOYMENT_NS) -> None:
    """
    Set the replica count for the specified deployment.
    """
    try:
        body = {'spec': {'replicas': replicas}}
        apps_v1_api.patch_namespaced_deployment_scale(name=deployment_name,
                                                     namespace=namespace,
                                                     body=body)
    except Exception as exc:
        logger.exception("Error scaling deployment '%s' in namespace '%s' to %d replicas",
                         deployment_name, namespace, replicas)
        raise CmstoolsException from exc

def get_pod_count_for_deployment(deployment_name: str, namespace: str = DEFAULT_DEPLOYMENT_NS) -> int:
    """
    Return the number of pods in the specified namespace for a given deployment.
    """
    try:
        # Get the deployment to extract its selector
        deployment = apps_v1_api.read_namespaced_deployment(name=deployment_name, namespace=namespace)
        selector = deployment.spec.selector.match_labels
        label_selector = ",".join([f"{k}={v}" for k, v in selector.items()])
        pods = k8s_client_api.list_namespaced_pod(namespace=namespace, label_selector=label_selector)
        return len(pods.items)
    except Exception as exc:
        logger.exception("Error retrieving pod count for deployment '%s' in namespace '%s'", deployment_name, namespace)
        raise CmstoolsException from exc

def check_replicas_and_pods_scaled(deployment_name: str, expected_replicas: int) -> None:
    """
    Ensure deployment is scaled and all pods are terminated.
    """
    start_time = time.time()
    while True:
        actual_replicas = get_deployment_replicas(deployment_name=deployment_name)

        if actual_replicas == expected_replicas and get_pod_count_for_deployment(deployment_name=deployment_name) == 0:
            logger.info(f"Deployment {deployment_name} scaled to {expected_replicas} replicas and all pods terminated")
            return
        if time.time() - start_time > 300:
            logger.error(f"Timeout: Deployment {deployment_name} did not scale to {expected_replicas} replicas and terminate pods within 5 minutes")
            raise CmstoolsException()

        logger.info(f"Waiting for deployment {deployment_name} to scale down and pods to terminate.")
        time.sleep(5)

# initialize k8s
config.load_kube_config()
k8s_client_api = client.CoreV1Api()
apps_v1_api = client.AppsV1Api()
