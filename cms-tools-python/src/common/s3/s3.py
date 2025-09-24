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
S3 module for barebones boot test
"""

import base64
import warnings

import boto3
from urllib3.exceptions import InsecureRequestWarning

from common.defs import JsonDict
from common.k8s import get_k8s_secret_data
from common.log import logger

from .defs import S3_CREDS_SECRET_FIELDS, S3_CREDS_SECRET_NAME
from .s3_url import S3Url

def s3_client_kwargs() -> JsonDict:
    """
    Decode the S3 credentials from the Kubernetes secret, and return
    the kwargs needed to initialize the boto3 client.
    """
    secret_data = get_k8s_secret_data(S3_CREDS_SECRET_NAME)
    logger.debug("Reading fields from Kubernetes secret '%s'", S3_CREDS_SECRET_NAME)
    encoded_s3_secret_fields = { field: secret_data[secret_field]
                                 for field, secret_field in S3_CREDS_SECRET_FIELDS.items() }
    logger.debug("Decoding fields from Kubernetes secret '%s'", S3_CREDS_SECRET_NAME)
    kwargs = { field: base64.b64decode(encoded_field).decode()
               for field, encoded_field in encoded_s3_secret_fields.items() }
    # Need to convert the 'verify' field to boolean if it is false
    if not kwargs["verify"] or kwargs["verify"].lower() in ('false', 'off', 'no', 'f', '0'):
        kwargs["verify"] = False

    # And if Verify is false, then we need to make sure that our endpoint isn't https, since
    # it will use SSL verification regardless if the endpoint is https
    if kwargs["verify"] is False and kwargs["endpoint_url"][:6] == "https:":
        kwargs["endpoint_url"] = f"http:{kwargs['endpoint_url'][6:]}"

    return kwargs

def s3_client():
    """
    Initialize the boto3 client and return it
    """
    client_kwargs = s3_client_kwargs()
    logger.debug("Getting boto3 S3 client")
    return boto3.client('s3', **client_kwargs)

def get_s3_artifact_etag(s3_url: S3Url) -> str:
    """
    Return the etag field of the specified object in S3.
    """
    s3_cli = s3_client()
    logger.debug("Retrieving data on S3 artifact with key '%s' in bucket '%s'", s3_url.key,
                 s3_url.bucket)
    # Suppress insecure request warnings from this call
    with warnings.catch_warnings():
        warnings.filterwarnings('ignore', category=InsecureRequestWarning)
        s3_resp = s3_cli.head_object(Bucket=s3_url.bucket, Key=s3_url.key)
    logger.debug("S3 response: %s", s3_resp)
    etag = s3_resp["ETag"].strip('"')
    logger.debug("ETag value is '%s'", etag)
    return etag
