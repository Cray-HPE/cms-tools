# Copyright 2020 Hewlett Packard Enterprise Development LP

"""
Kubernetes-related CMS test helper functions
"""

from .helpers import debug
import base64
import kubernetes
import warnings
import yaml

saved_k8s_client = None

def k8s_client():
    """
    Initializes the kubernetes client instance (if needed) and
    returns it.
    """
    global saved_k8s_client
    if saved_k8s_client != None:
        return saved_k8s_client
    debug("Initializing kubernetes client")
    with warnings.catch_warnings():
        warnings.simplefilter("ignore", category=yaml.YAMLLoadWarning)
        kubernetes.config.load_kube_config()
    saved_k8s_client = kubernetes.client.CoreV1Api()
    return saved_k8s_client

def get_k8s_secret():
    """
    Returns the kubernetes admin-client-auth secret string
    """
    debug("Getting client secret from kubernetes")
    k8s_secret = k8s_client().read_namespaced_secret(name="admin-client-auth", namespace="default")
    k8sec = k8s_secret.data["client-secret"]
    debug("Client secret is %s" % k8sec)
    debug('Decoding client secret from base64')
    secret = base64.b64decode(k8sec)
    debug("Decoded secret is %s" % secret)
    return secret

def get_vcs_username_password():
    """
    Retrieves the vcs username and password from the k8s vcs-user-credentials secret
    """
    debug("Getting vcs-user-credentials secret from kubernetes")
    k8s_secret = k8s_client().read_namespaced_secret(name="vcs-user-credentials", namespace="services")
    user64, pass64 = k8s_secret.data["vcs_username"], k8s_secret.data["vcs_password"]
    debug("Base64 vcs username is %s, password is %s" % (user64, pass64))
    debug('Decoding from base64')
    username = base64.b64decode(user64).decode("ascii").rstrip()
    password = base64.b64decode(pass64).decode("ascii").rstrip()
    debug("Decoded username is %s, password is %s" % (username, password))
    return username, password
