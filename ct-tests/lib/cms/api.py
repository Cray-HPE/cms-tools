# Copyright 2020 Hewlett Packard Enterprise Development LP

"""
CMS test helper functions for API calls
"""

from .helpers import debug, info, raise_test_error, raise_test_exception_error
from .k8s import get_k8s_secret
import requests
import warnings

URL_BASE = "https://api-gw-service-nmn.local"
API_URL_BASE = "%s/apis" % URL_BASE

token_url="%s/keycloak/realms/shasta/protocol/openid-connect/token" % URL_BASE

saved_auth_token = None

#
# API utility functions
#

def show_response(resp):
    """
    Displays and logs the contents of an API response
    """
    debug("Status code of API response: %d" % resp.status_code)
    for field in ['reason','headers','text']:
        val = getattr(resp, field)
        if val:
            debug("API response %s: %s" % (field, str(val)))

def do_request(method, url, **kwargs):
    """
    Wrapper for call to requests functions. Displays, logs, and makes the request,
    then displays, logs, and returns the response.
    """
    req_args = { "verify": False, "timeout": 30 }
    req_args.update(kwargs)
    debug("Sending %s request to %s with following arguments" % (method.__name__, url))
    for k in req_args:
        debug("%s = %s" % (k, str(req_args[k])))
    with warnings.catch_warnings():
        warnings.simplefilter("ignore", 
            category=requests.packages.urllib3.exceptions.InsecureRequestWarning)
        try:
            resp = method(url=url, **req_args)
            show_response(resp)
            return resp
        except Exception as e:
            raise_test_exception_error(e, "API request")

def check_response(resp, expected_sc=200, return_json=False):
    """
    Checks to make sure the response has the expected status code. If requested,
    returns the JSON object from thje response.
    """
    if resp.status_code != expected_sc:
        raise_test_error("Request status code expected to be %d, but was not" % expected_sc)
    if return_json:
        try:
            return resp.json()
        except Exception as e:
            raise_test_exception_error(e, "to decode JSON object in response body")

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
    if "json" in kwargs:
        try:
            if "Content-Type" not in kwargs["headers"]:
                kwargs["headers"]["Content-Type"] = "application/json"
        except KeyError:
            kwargs["headers"] = { "Content-Type": "application/json" }
    auth_token = get_auth_token()
    try:
        kwargs["headers"]["Authorization"] = "Bearer %s" % auth_token["access_token"]
    except KeyError:
        kwargs["headers"] = { "Authorization": "Bearer %s" % auth_token["access_token"] }
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
            raise_test_error("Expected response with status code %d" % expected_sc)
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

