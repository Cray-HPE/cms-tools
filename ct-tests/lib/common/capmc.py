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
CAPMC-related test helper functions
"""

from .api import API_URL_BASE, requests_post
from .cli import int_list_to_str, run_cli_cmd
from .helpers import debug, error, get_int_field_from_obj, \
                     get_list_field_from_obj, \
                     get_str_field_from_obj, info, \
                     raise_test_error, sleep, warn
import time

CAPMC_URL_BASE = "%s/capmc/capmc" % API_URL_BASE
CAPMC_GET_NODE_STATUS_URL = "%s/get_node_status" % CAPMC_URL_BASE

def do_get_node_status(use_api, nidlist, source=None):
    """
    Calls CAPMC get_node_status on the specified nodes and returns the result.
    """
    if use_api:
        data = { "nids": nidlist }
        if source != None:
            data["source"] = source
        return requests_post(CAPMC_GET_NODE_STATUS_URL, json=data, expected_sc=200)
    else:
        cli_cmd = ["capmc", "get_node_status", "create", "--nids", int_list_to_str(nidlist)]
        if source != None:
            cli_cmd.extend(["--source", source])
        return run_cli_cmd(cli_cmd)

def get_err_vals(response_object):
    """
    Given a CAPMC response object, returns the error value (e field) and error message (err_msg field).
    """
    err_val = get_int_field_from_obj(response_object, "e", noun="CAPMC response")
    err_msg = get_str_field_from_obj(response_object, "err_msg", noun="CAPMC response")
    for k in ['alert', 'warn']:
        if k in response_object:
            warn("CAPMC response included %s field: %s" %(k,  str(response_object[k])))
    return err_val, err_msg

def map_nids_to_states(response_object, nidlist, undefined_okay=False):
    """
    Given a CAPMC get_node_status response, check to see if it reports any errors, and return a map
    from each nid to its state as reported by CAPMC. If undefined_okay is false, an exception is raised
    if any nodes are in the undefined state.
    """
    def get_nids_with_power_state(power_state):
        if power_state not in response_object:
            return list()
        return get_list_field_from_obj(response_object, power_state, noun="CAPMC response", member_type=int)

    def has_invalid_nids(nidlist_to_check):
        invalid_nids = [ n for n in nidlist_to_check if n not in nidlist ]
        if invalid_nids:
            error("CAPMC response included nids (%s) that were not in the request" % str(invalid_nids))
            return True
        return False

    errors_found = False
    err_val, err_msg = get_err_vals(response_object)
    if err_msg:
        error("CAPMC response has error message: %s" % err_msg)
        if err_val == 0:
            error("CAPMC response has error message but error value of 0")
        errors_found = True

    undefined_nids = get_nids_with_power_state("undefined")
    if has_invalid_nids(undefined_nids):
        errors_found = True
    if not undefined_nids and err_val != 0:
        errors_found = True
        error("CAPMC response has no error message and no undefined nids, but its error value is %d" % err_val)

    nid_to_state = { n: "undefined" for n in undefined_nids }

    for k in response_object.keys():
        if k in { 'alert', 'warn', 'e', 'err_msg' }:
            continue
        nids_in_state = get_nids_with_power_state(k)
        for n in nids_in_state:
            try:
                error("NID %d state reported as both %s and %s" % (n, nid_to_state[n], k))
                errors_found = True
            except KeyError:
                nid_to_state[n] = k

    missing_nids = [ n for n in nidlist if n not in nid_to_state.keys() ]
    if missing_nids:
        error("The following NIDs did not have their state reported: %s" % str(missing_nids))
        errors_found = True

    if undefined_nids and not undefined_okay:
        error("The following NIDs are in undefined state: %s" % str(undefined_nids))
        errors_found = True

    if errors_found:
        raise_test_error("At least one error found in CAPMC response")
    return nid_to_state

def get_capmc_node_status(use_api, nids, return_undefined=False, source="Redfish", timeout=90):
    """
    Return a mapping between the specified list of nids and their power state, as reported by CAPMC
    using the specified --source option. For any nodes in the undefined state, CAPMC will retry for a 
    limited period of time before giving up. If return_undefined is true, this function will return a 
    mapping which includes nodes that are undefined (after retries). Otherwise if there are undefined
    nodes (after retries), the function will raise an error.
    """
    expected_states = [ "off", "on", "undefined" ]
    good_states = [ "off", "on" ]
    if return_undefined:
        good_states.append("undefined")
    nid_to_capmc_status = dict()
    nidlist = nids
    stop_time = time.time() + timeout
    while True:
        response_object = do_get_node_status(use_api=use_api, nidlist=nidlist, source=source)
        nid_to_state = map_nids_to_states(response_object, nidlist, undefined_okay=True)
        invalid_states = [ state for state in set(nid_to_state.values()) if state not in expected_states ]
        if invalid_states:
            info("Using Redfish source, valid CAPMC node states are: " % ", ".join(expected_states))
            for istate in invalid_states:
                nids_in_state = [ str(n) for n, s in nid_to_state.items() if s == istate ]
                error("CAPMC reports state %s for nid(s): %s" % (istate, ", ".join(nids_in_state)))
            raise_test_error("Invalid CAPMC node state(s) found: %s" % ", ".join(invalid_states))
        undefined_nids = list()
        for n,s in nid_to_state.items():
            if s in good_states:
                nid_to_capmc_status[n] = s
            elif s != "undefined":
                # Our previous checks should mean the only possible values should be "off", "on", and "undefined"
                raise_test_error("PROGRAMMING LOGIC ERROR: Unexpected CAPNC state (%s) for nid %d" % (s, n))
            else:
                # Note that if return_undefined is true, then this clause will never be hit, since
                # undefined will be in the "good_states" list
                undefined_nids.append(n)
        if not undefined_nids:
            return nid_to_capmc_status
        elif time.time() >= stop_time:
            raise_test_error("Even after %d seconds, CAPMC reports some nodes with undefined power status" % timeout)
        nidlist=undefined_nids
        sleep(5)
