#
# MIT License
#
# (C) Copyright 2020-2022 Hewlett Packard Enterprise Development LP
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
CMS test helper functions for API calls
"""

from .helpers import CMSTestError, debug, info, raise_test_error, raise_test_exception_error
from .k8s import get_k8s_secret
import copy
import requests
import time
import warnings

API_HOSTNAME = "api-gw-service-nmn.local"
URL_BASE = "https://" + API_HOSTNAME
API_URL_BASE = URL_BASE + "/apis"

token_url = URL_BASE + "/keycloak/realms/shasta/protocol/openid-connect/token"

saved_auth_token = None

HEADER_CONTENT_TYPE = "Content-Type"
CONTENT_TYPE_JSON = "application/json"

class CMSTestApiError(CMSTestError):
    pass

class CMSTestApiUnexpectedStatusCodeError(CMSTestApiError):
    def __init__(self, expected_sc, actual_sc, **kwargs):
        msg = "Expected status code %d in response but received %d" % (expected_sc, actual_sc)
        self.expected_sc = expected_sc
        self.actual_sc = actual_sc
        super(CMSTestApiUnexpectedStatusCodeError, self).__init__(msg, **kwargs)

#
# API utility functions
#

def json_content_headers(headers=None, overwrite_content_type=False):
    if headers == None:
        return { HEADER_CONTENT_TYPE: CONTENT_TYPE_JSON }
    new_headers = copy.deepcopy(headers)
    if overwrite_content_type or HEADER_CONTENT_TYPE not in new_headers:
        new_headers[HEADER_CONTENT_TYPE] = CONTENT_TYPE_JSON
    return new_headers

def show_response(resp):
    """
    Displays and logs the contents of an API response
    """
    debug("Status code of API response: %d" % resp.status_code)
    for field in ['reason','headers','text']:
        val = getattr(resp, field)
        if val:
            debug("API response %s: %s" % (field, str(val)))

def do_request(method, url, max_retries_on_5xx=3, **kwargs):
    """
    Wrapper for call to requests functions. Displays, logs, and makes the request,
    then displays, logs, and returns the response.
    """
    req_args = { "verify": False, "timeout": 120 }
    req_args.update(kwargs)
    debug("Sending %s request to %s with following arguments" % (method.__name__.upper(), url))
    for k in req_args:
        debug("%s = %s" % (k, str(req_args[k])))
    with warnings.catch_warnings():
        warnings.simplefilter("ignore", 
            category=requests.packages.urllib3.exceptions.InsecureRequestWarning)
        try:
            resp = method(url=url, **req_args)
            show_response(resp)
            if max_retries_on_5xx > 0 and 500 <= resp.status_code <= 599:
                info("Received status code %d; waiting 2 seconds and retrying request" % resp.status_code)
                time.sleep(2)
                return do_request(method=method, url=url, max_retries_on_5xx=max_retries_on_5xx-1, **kwargs)
            return resp
        except Exception as e:
            raise_test_exception_error(e, "API request")

def get_response_json(resp):
    """
    Return the JSON object from the response or raise an error
    """
    try:
        return resp.json()
    except Exception as e:
        raise_test_exception_error(e, "to decode JSON object in response body")

def check_response(resp, expected_sc=200, return_json=False):
    """
    Checks to make sure the response has the expected status code. If requested,
    returns the JSON object from the response.
    """
    if resp.status_code != expected_sc:
        raise CMSTestApiUnexpectedStatusCodeError(expected_sc=expected_sc, actual_sc=resp.status_code)
    if return_json:
        return get_response_json(resp)
    return resp

#
# Auth functions
#

def validate_auth_token_response(token_resp):
    """
    Verifies that the auth token response we received contains the fields we expect.
    """
    auth_token = check_response(resp=token_resp, return_json=True)
    for k in [ "access_token", "refresh_token" ]:
        try:
            if k not in auth_token:
                raise_test_error("%s field not found in JSON object of response" % k)
        except Exception as e:
            raise_test_exception_error(e, "checking %s field from JSON object in response" % k)
    return auth_token

def get_auth_token():
    """
    Requests and stores a new auth token
    """
    global saved_auth_token
    if saved_auth_token != None:
        return saved_auth_token
    info("Getting auth token")
    secret = get_k8s_secret()
    request_data = { 
        "grant_type": "client_credentials",
        "client_id": "admin-client",
        "client_secret": secret }
    token_resp = do_request(method=requests.post, url=token_url, data=request_data)
    saved_auth_token = validate_auth_token_response(token_resp)
    info("Auth token successfully obtained")
    return saved_auth_token

def refresh_auth_token(auth_token):
    """
    Refreshes a previously-obtained auth token
    """
    info("Refreshing auth token")
    secret = get_k8_secret()
    request_data = { 
        "grant_type": "refresh_token",
        "refresh_token": auth_token["refresh_token"],
        "client_id": "admin-client",
        "client_secret": secret }
    token_resp = do_request(method=requests.post, url=token_url, data=request_data)
    auth_token = validate_auth_token_response(token_resp)
    info("Auth token successfully refreshed")
    return auth_token

def do_request_with_auth_retry(url, method, expected_sc, return_json=None, **kwargs):
    """
    Wrapper to our earlier requests wrapper. This wrapper calls the previous wrapper,
    but if the response indicates an expired token error, then the token is refreshed
    and the request is re-tried with the refreshed token. A maximum of one retry will
    be attempted.
    
    If expected_sc is in the 200 but not 204, return_json defaults to True
    Otherwise, return_json defaults to False.
    
    If a JSON object is being included in the request, the appropriate content-type field is set in the
    header, if not already set.
    """
    if return_json == None:
        if 200 <= expected_sc <= 299 and expected_sc != 204:
            return_json = True
        else:
            return_json = False
    if 500 <= expected_sc <= 599 and "max_retries_on_5xx" not in kwargs:
        # If our expected status code is in the 500 range then we don't want to
        # do any automatic retries if we get it
        kwargs["max_retries_on_5xx"] = 0
    if "json" in kwargs:
        try:
            headers = kwargs["headers"]
        except KeyError:
            headers = None
        kwargs["headers"] = json_content_headers(headers=headers)
    auth_token = get_auth_token()
    try:
        kwargs["headers"]["Authorization"] = "Bearer %s" % auth_token["access_token"]
    except KeyError:
        kwargs["headers"] = { "Authorization": "Bearer %s" % auth_token["access_token"] }
    debug("kwargs = %s" % str(kwargs))
    resp = do_request(method=method, url=url, **kwargs)
    if resp.status_code != 401 or expected_sc == 401:
        if return_json:
            return check_response(resp=resp, expected_sc=expected_sc, return_json=True)
        check_response(resp=resp, expected_sc=expected_sc)
        return resp
    else:
        json_obj = check_response(resp=resp, expected_sc=401, return_json=True)
    try:
        if json_obj["exp"] != "token expired":
            raise CMSTestApiUnexpectedStatusCodeError(expected_sc=expected_sc, actual_sc=resp.status_code)
    except KeyError:
        raise_test_error("Expected response with status code %d" % expected_sc)
    debug("Received token expired response (status code 401). Will attempt to refresh auth token and retry request")
    auth_token = refresh_auth_token()
    kwargs["headers"]["Authorization"] = "Bearer %s" % auth_token["access_token"]
    debug("Retrying request")
    resp = do_request(method, *args, **kwargs)
    if return_json:
        return check_response(resp=resp, expected_sc=expected_sc, return_json=True)
    check_response(resp=resp, expected_sc=expected_sc)
    return resp

#
# Requests functions
#

def requests_delete(url, expected_sc=204, **kwargs):
    """
    Calls our above requests wrapper for a DELETE request, setting the default expected status code to 204
    """
    return do_request_with_auth_retry(url=url, method=requests.delete, expected_sc=expected_sc, **kwargs)

def requests_get(url, expected_sc=200, **kwargs):
    """
    Calls our above requests wrapper for a GET request, and sets the default expected status code to 200
    """
    return do_request_with_auth_retry(url=url, method=requests.get, expected_sc=expected_sc, **kwargs)

def requests_patch(url, expected_sc=200, **kwargs):
    """
    Calls our above requests wrapper for a PATCH request, and sets the default expected status code to 200.
    """
    return do_request_with_auth_retry(url=url, method=requests.patch, expected_sc=expected_sc, **kwargs)

def requests_post(url, expected_sc=201, **kwargs):
    """
    Calls our above requests wrapper for a POST request, and sets the default expected status code to 201.
    """
    return do_request_with_auth_retry(url=url, method=requests.post, expected_sc=expected_sc, **kwargs)

def requests_put(url, expected_sc=200, **kwargs):
    """
    Calls our above requests wrapper for a PUT request, and sets the default expected status code to 200.
    """
    return do_request_with_auth_retry(url=url, method=requests.put, expected_sc=expected_sc, **kwargs)
