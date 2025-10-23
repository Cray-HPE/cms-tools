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
API module cmstools tests
"""

import base64
import json

import requests
from requests_retry_session import requests_retry_session
from urllib3.exceptions import MaxRetryError

from cmstools.lib.defs import CmstoolsException, JsonDict, JsonObject
from cmstools.lib.k8s import get_k8s_secret_data, S3_CREDS_SECRET_NS
from cmstools.lib.common_logger import logger

# set up gateway address
API_GW_DNSNAME = "api-gw-service-nmn.local"
API_GW_SECURE = f"https://{API_GW_DNSNAME}"
API_BASE_URL = f"{API_GW_SECURE}/apis"

SYSTEM_CA_CERTS = "/etc/ssl/ca-bundle.pem"


def add_api_auth(headers: JsonDict) -> None:
    """
    Get the admin secret from k8s for the api gateway - command line equivalent is:
    #`kubectl get secrets admin-client-auth -o jsonpath='{.data.client-secret}' | base64 -d`
    """
    secret_data = get_k8s_secret_data(sec_name="admin-client-auth", sec_namespace=S3_CREDS_SECRET_NS)
    try:
        encoded_admin_secret = secret_data['client-secret']
        admin_secret = base64.b64decode(encoded_admin_secret)
    except Exception as exc:
        logger.exception("Errpr accessing or decoding admin client secret")
        raise CmstoolsException from exc

    # get an access token from keycloak
    payload = {"grant_type":"client_credentials",
               "client_id":"admin-client",
               "client_secret":admin_secret}
    url = f"{API_GW_SECURE}/keycloak/realms/shasta/protocol/openid-connect/token"
    resp_json = request_and_check_status("post", url, data=payload, add_auth_header=False,
                                         expected_status=200, parse_json=True)

    # pull the access token from the return data
    headers["Authorization"] = f"Bearer {resp_json['access_token']}"

def request(verb, url, headers=None, add_auth_header=True, verify=SYSTEM_CA_CERTS,
            **kwargs) -> requests.Response:
    """
    Wrapper for making requests with retries.
    Automatically adds API auth to header, if specified.
    """
    if add_auth_header:
        logger.debug("API %s request to %s with args: %s", verb, url, kwargs)
        if headers:
            logger.debug("headers: %s", headers)
        else:
            headers = {}
        add_api_auth(headers)
    else:
        # We don't want the client_secret to be logged
        logger.debug("API %s request to %s (args not logged)", verb, url)

    session = requests_retry_session()
    try:
        return session.request(verb, url=url, headers=headers, verify=verify, **kwargs)
    except (requests.exceptions.ConnectionError, MaxRetryError) as exc:
        logger.exception("Unable to connect to %s", url)
        raise CmstoolsException from exc
    except requests.exceptions.HTTPError as exc:
        logger.exception("Unexpected response making %s request to %s", verb, url)
        raise CmstoolsException from exc
    except Exception as exc:
        logger.exception("Unhandled exception making %s request to %s", verb, url)
        raise CmstoolsException from exc

def request_and_check_status(verb, url, expected_status: int, parse_json: bool,
                             **kwargs) -> requests.Response|JsonObject:
    """
    A wrapper for request that also checks the status code of the response.
    Raises an exception if it doesn't match.
    If parse_json is True, then returns the parsed JSON body.
    Otherwise, returns the raw response object.
    """
    resp = request(verb=verb, url=url, **kwargs)
    if resp.status_code != expected_status:
        logger.error("API %s request to %s received incorrect return code %d (expected %d): %s",
                     verb, url, resp.status_code, expected_status, resp.text)
        raise CmstoolsException()
    if not parse_json:
        return resp
    try:
        return json.loads(resp.text)
    except json.JSONDecodeError as exc:
        logger.exception("Non-JSON response to %s request to %s", verb, url)
        raise CmstoolsException from exc
    except Exception as exc:
        logger.exception("Unhandled exception making %s request to %s", verb, url)
        raise CmstoolsException from exc
