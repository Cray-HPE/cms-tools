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
#
# (MIT License)

"""
BSS-related CMS test helper functions
"""

from .api import API_URL_BASE, requests_get
from .cli import int_list_to_str, run_cli_cmd
from .helpers import debug, get_bool_field_from_obj, get_int_field_from_obj, \
                     get_list_field_from_obj, get_str_field_from_obj, \
                     info, raise_test_error

BSS_URL_BASE = "%s/bss" % API_URL_BASE
BSS_BOOTPARAMETERS_URL = "%s/boot/v1/bootparameters" % BSS_URL_BASE
BSS_HOSTS_URL = "%s/boot/v1/hosts" % BSS_URL_BASE

def verify_bss_response(response_object, expected_length=None):
    """
    Verify some basic things about the bss response object:
    1) That it is a list
    2) That it has the expected length, if specified
    3) That all list entries are type dict
    """
    debug("Validating BSS response")
    if not isinstance(response_object, list):
        raise_test_error("We expect the bss response to be a list but it is %s" % str(type(response_object)))
    elif expected_length != None and len(response_object) != expected_length:
        raise_test_error("We expect bss response list length to be %d, but it is %d" % (expected_length, len(response_object)))
    elif not all(isinstance(x, dict) for x in response_object):
        raise_test_error("We expect the bss list items to all be dicts, but at least one is not")

def validate_bss_host_entry(host, nid=None, xname=None):
    """
    Validates the following about the specified bss host object:
    1) The nid field is an integer equal to the specified nid, or if none specified, is a positive integer
    2) The Role field is a non-empty string
    3) The xname field is a string equal to the specified xname, or if none specified, looks at least a little
       bit like an xname
    4) The Enabled field is a boolean
    5) The Type field is Node
    """
    noun="bss host list entry"
    if nid != None:
        get_int_field_from_obj(host, "NID", noun="%s for nid %d" % (noun, nid), exact_value=nid)
    else:
        get_int_field_from_obj(host, "NID", noun=noun, min_value=1)
    get_str_field_from_obj(host, "Role", noun=noun, min_length=1)
    if xname != None:
        get_str_field_from_obj(host, "ID", noun="%s for xname %s" % (noun, xname), exact_value=xname)
    else:
        get_str_field_from_obj(host, "ID", noun=noun, min_length=2, prefix="x")
    get_str_field_from_obj(host, "Type", noun=noun, exact_value="Node")
    get_bool_field_from_obj(host, "Enabled", noun=noun, null_okay=False)

def get_bss_host_by_nid(use_api, nid, expected_xname, enabled_only=True):
    """
    List all host entries in bss for the specified nid. Validate that there is only one such
    entry, verify that it specifies the same xname that we expect, verify that its Role field
    is not empty, and then return the host object.
    """
    if use_api:
        params={'mac': None, 'name': None, 'nid': nid}
        response_object = requests_get(BSS_HOSTS_URL, params=params)
    else:
        response_object = run_cli_cmd(["bss", "hosts", "list", "--nid", str(nid)])

    verify_bss_response(response_object, 1)
    host = response_object[0]
    validate_bss_host_entry(host=host, nid=nid, xname=expected_xname)
    if enabled_only and not host["Enabled"]:
        raise_test_error("BSS host entry for NID %d is marked not Enabled" % nid)
    return host

def list_bss_hosts(use_api, enabled_only=True):
    """
    List all host entries in bss, verify that they look okay, then return the list.
    """
    info("Listing all BSS hosts")
    if use_api:
        params={'mac': None, 'name': None, 'nid': None}
        response_object = requests_get(BSS_HOSTS_URL, params=params)
    else:
        response_object = run_cli_cmd(["bss", "hosts", "list"])

    verify_bss_response(response_object)
    host_list = list()
    for host in response_object:
        debug("Examining host: %s" % str(host))
        validate_bss_host_entry(host=host)
        if host["Enabled"] or not enabled_only:
            host_list.append(host)
    return host_list

def bss_host_nid(host):
    """
    Retrieves the nid from the BSS host object
    """
    return host["NID"]

def bss_host_xname(host):
    """
    Retrieves the xname from the BSS host object
    """
    return host["ID"]

def bss_host_role(host):
    """
    Retrieves the xname from the BSS host object
    """
    return host["Role"]

def get_bss_nodes_by_role(use_api, role, enabled_only=True):
    """
    List all bss host entries, validate that they look legal, and returns a list of the
    entries with the specified role
    """
    bss_host_list = list_bss_hosts(use_api, enabled_only=enabled_only)
    return [ host for host in bss_host_list if host["Role"] == role ]

def get_bss_compute_nodes(use_api, min_number=1, enabled_only=True):
    """
    List all bss host entries for compute nodes, validate that they look legal, and returns the list.
    """
    bss_host_list = get_bss_nodes_by_role(use_api, role="Compute", enabled_only=enabled_only)
    if len(bss_host_list) < min_number:
        raise_test_error("We need at least %d compute node(s), but only %d found in BSS!" % (min_number, len(bss_host_list)))
    return bss_host_list

def verify_bss_bootparameters_list(response_object, xname_to_nid):
    """
    Validates that the list of bss bootparameters looks valid and has all of the
    xnames we expect to find in it.
    """
    verify_bss_response(response_object, len(xname_to_nid))
    xnames_to_find = set(xname_to_nid.keys())
    for bootparams in response_object:
        if "params" not in bootparams:
            raise_test_error("We expect boot parameters entry to have 'params' field, but this one does not: %s" % bootparams)
        elif "kernel" not in bootparams:
            raise_test_error("We expect boot parameters entry to have 'kernel' field, but this one does not: %s" % bootparams)
        hostlist = get_list_field_from_obj(bootparams, "hosts", noun="boot parameters list entry", member_type=str, min_length=1)
        xnames_to_find.difference_update(hostlist)
    if xnames_to_find:
        raise_test_error("Did not find bootparameter entries for the following nids/xnames: %s" % ", ".join(
            ["%d/%s" % (xname_to_nid[xname], xname) for xname in xnames_to_find]))

def list_all_bss_bootparameters(use_api):
    """
    Returns list of all BSS bootparameters
    """
    if use_api:
        params={'mac': None, 'name': None, 'nid': None}
        response_object = requests_get(BSS_BOOTPARAMETERS_URL, params=params)
    else:
        response_object = run_cli_cmd(["bss", "bootparameters", "list"])
    verify_bss_bootparameters_list(response_object, dict())
    return response_object

def list_bss_bootparameters(use_api, nid, xname):
    """
    Generate a list of bss bootparameters for the specified nid, validate the list, and return its first entry.
    """
    if use_api:
        params={'mac': None, 'name': None, 'nid': nid}
        response_object = requests_get(BSS_BOOTPARAMETERS_URL, params=params)
    else:
        response_object = run_cli_cmd(["bss", "bootparameters", "list", "--nid", str(nid)])
    verify_bss_bootparameters_list(response_object, {xname: nid})
    return response_object[0]

def list_bss_bootparameters_nidlist(use_api, xname_to_nid):
    """
    List bss boot parameters for all specified nids, validate them, and return the list.
    """
    nids = list(xname_to_nid.values())
    if use_api:
        params={'mac': None, 'name': None, 'nid': None}
        json_data={ "nids": nids }
        response_object = requests_get(BSS_BOOTPARAMETERS_URL, params=params, json=json_data)
    else:
        response_object = run_cli_cmd(["bss", "bootparameters", "list", "--nids", int_list_to_str(nids)])
    verify_bss_bootparameters_list(response_object, xname_to_nid)
    return response_object
