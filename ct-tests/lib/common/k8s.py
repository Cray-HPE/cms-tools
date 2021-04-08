# Copyright 2020-2021 Hewlett Packard Enterprise Development LP
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
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

"""
Kubernetes-related CMS test helper functions
"""

from .helpers import debug
import base64
import kubernetes
import warnings
import yaml

saved_k8s_client = None
saved_csm_private_key = None

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

def get_csm_private_key():
    """
    If it has previously been retrieved, return it, otherwise 
    retrieve (and save for future calls) the CSM private key for use when
    sshing to compute nodes
    """
    global saved_csm_private_key
    if saved_csm_private_key == None:
        debug("Getting CSM private key from kubernetes")
        k8s_secret = k8s_client().read_namespaced_secret(name="csm-private-key", namespace="services")
        csmpk64 = k8s_secret.data["value"]
        debug("CSM private key (base 64) is %s" % csmpk64)
        debug('Decoding from base64')
        csmpk = base64.b64decode(csmpk64).decode("ascii").rstrip()
        debug("Decoded CSM private key is %s" % csmpk)
        saved_csm_private_key = csmpk
    return saved_csm_private_key
