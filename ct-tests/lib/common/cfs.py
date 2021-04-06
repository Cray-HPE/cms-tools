# Copyright 2021 Hewlett Packard Enterprise Development LP
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
CFS-related CMS test helper functions
"""

from .api import API_URL_BASE, requests_delete, requests_get, requests_put
from .cli import run_cli_cmd
from .helpers import debug, get_dict_field_from_obj, get_list_field_from_obj, \
                     get_str_field_from_obj, info, raise_test_error, \
                     run_cmd_list
import datetime
import json
import string
import tempfile

CFS_URL_BASE = "%s/cfs/v2" % API_URL_BASE
CFS_CONFIGS_URL_BASE = "%s/configurations" % CFS_URL_BASE

def cfs_configs_url(name=None):
    """
    Returns the url endpoint for cfs configurations generally, or the specified configuration
    """
    if name != None:
        return "%s/%s" % (CFS_CONFIGS_URL_BASE, name)
    else:
        return CFS_CONFIGS_URL_BASE

def validate_cfs_config(config_object, min_layers=0, expected_name=None, expected_lastUpdated=None, expected_layers=None):
    """
    Validate that the config object has all expected fields with the expected types:
    - name field, type string, nonzero length
    - lastUpdated field, type string, nonzero length 
      (eventually would be good to also validate RFC 3339 format)
    And validates that if any optional fields are present, they also are of the expected types:
    - layers field, type list, and for each entry we verify the following
        - Has cloneUrl field, type string, nonzero length
        - Has commit field, type string, nonzero length
        - Optionally has name field, type string
        - Optionally has playbook field, type string
    And for name, lastUpdated, and layers, if an expected value has been specified, validate that too
    """
    # First, sanity check our min_layers and expected_layers arguments to make sure they are not in conflict
    if expected_layers != None:
        if len(expected_layers) < min_layers:
            debug("expected_layers = %s" % str(expected_layers))
            raise_test_error(
                "PROGRAMMING LOGIC ERROR: validate_cfs_config: min_layers = %d > len(expected_layers) = %d" % (
                    min_layers, len(expected_layers)))
        elayers_ok = True
        for i, elayer in enumerate(expected_layers):
            for k, v in elayer.items():
                if k not in { "name", "playbook", "commit", "cloneUrl" }:
                    error("expected_layer %d has unknown field %s with value '%s' (type %s)" % (i, k, str(v), str(type(v))))
                    elayers_ok = False
        if not elayers_ok:
            raise_test_error("PROGRAMMING LOGIC ERROR: validate_cfs_config: One or more expected layers has invalid fields")

    if expected_name != None:
        get_str_field_from_obj(config_object, "name", noun="CFS configuration", exact_value=expected_name)
    else:
        get_str_field_from_obj(config_object, "name", noun="CFS configuration", min_length=1)
    if expected_lastUpdated != None:
        get_str_field_from_obj(config_object, "lastUpdated", noun="CFS configuration", exact_value=expected_lastUpdated)
    else:
        get_str_field_from_obj(config_object, "lastUpdated", noun="CFS configuration", min_length=1)

    if expected_layers != None:
        layers = get_list_field_from_obj(config_object, "layers", noun="CFS configuration", member_type=dict, null_okay=False, exact_length=len(expected_layers))
    elif min_layers > 0:
        layers = get_list_field_from_obj(config_object, "layers", noun="CFS configuration", member_type=dict, null_okay=False, min_length=min_layers)
    elif "layers" not in config_object:
        return
    else:
        layers = get_list_field_from_obj(config_object, "layers", noun="CFS configuration", member_type=dict, null_okay=False)

    match_expected = True
    for i, layer in enumerate(layers):
        if expected_layers != None:
            # We know that if we got this far, it means that if expected_layers != None,
            # then len(layers) == len(expected_layers)
            elayer = expected_layers[i]
        if "name" in layer:
            name = get_str_field_from_obj(layer, "name", noun="CFS configuration layer", null_okay=False)
            if expected_layers != None:
                if "name" not in elayer:
                    error("layer %d is not expected to have a name field, but it has name '%s'" % (i, name))
                    match_expected = False
                elif elayer["name"] != name:
                    error("name field in layer %d is expected to be '%s' but it is actually '%s'" % (i, elayer["name"], name))
                    match_expected = False
        if "playbook" in layer:
            playbook = get_str_field_from_obj(layer, "playbook", noun="CFS configuration layer", null_okay=False)
            if expected_layers != None:
                if "playbook" not in elayer:
                    error("layer %d is not expected to have a playbook field, but it has playbook '%s'" % (i, playbook))
                    match_expected = False
                elif elayer["playbook"] != playbook:
                    error("playbook field in layer %d is expected to be '%s' but it is actually '%s'" % (i, elayer["playbook"], playbook))
                    match_expected = False
        commit = get_str_field_from_obj(layer, "commit", noun="CFS configuration layer", min_length=1)
        if expected_layers != None:
            if "commit" not in elayer:
                error("layer %d is not expected to have a commit field, but it has commit '%s'" % (i, commit))
                match_expected = False
            elif elayer["commit"] != commit:
                error("commit field in layer %d is expected to be '%s' but it is actually '%s'" % (i, elayer["commit"], commit))
                match_expected = False
        cloneUrl = get_str_field_from_obj(layer, "cloneUrl", noun="CFS configuration layer", min_length=1)
        if expected_layers != None:
            if "cloneUrl" not in elayer:
                error("layer %d is not expected to have a cloneUrl field, but it has cloneUrl '%s'" % (i, cloneUrl))
                match_expected = False
            elif elayer["cloneUrl"] != cloneUrl:
                error("cloneUrl field in layer %d is expected to be '%s' but it is actually '%s'" % (i, elayer["cloneUrl"], cloneUrl))
                match_expected = False
    if not match_expected:
        raise_test_error("One or more layers do not match what we expect")

def describe_cfs_config(use_api, name, expect_to_pass=True, **validate_cfs_config_kwargs):
    """
    Calls an API GET or CLI describe on the specified CFS configuration
    If we are expecting it to pass:
        1) The CFS configuration is extracted from the response
        2) We verify that the name matches the one we expect
        3) If min_layers > 0, then we verify that the layers field is a list with at least that many entries,
           and that those entries have the appropriate fields
        4) Return the config
    Otherwise we just validate that the request failed.
    """
    if use_api:
        url = cfs_configs_url(name)
        if expect_to_pass:
            response_object = requests_get(url)
        else:
            requests_get(url, expected_sc=404)
            return
    else:
        cmd_list = ["cfs","configurations","describe",name]
        if expect_to_pass:
            response_object = run_cli_cmd(cmd_list)
        else:
            cmdresp = run_cli_cmd(cmd_list, return_rc=True, parse_json_output=False)
            if cmdresp["rc"] == 0:
                raise_test_error("We expected this query to fail but the return code was 0")
            return
    validate_cfs_config(response_object, expected_name=name, **validate_cfs_config_kwargs)
    return response_object

def delete_cfs_config(use_api, name, verify_delete=True):
    if use_api:
        url = cfs_configs_url(name)
        requests_delete(url)
    else:
        cmd_list = ["cfs","configurations","delete",name]
        run_cli_cmd(cmd_list, parse_json_output=False)
    if verify_delete:
        debug("Verify that we can no longer retrieve the config")
        describe_cfs_config(use_api=use_api, name=name, expect_to_pass=False)

def create_cfs_config(use_api, name, layers=None, verify_create=True):
    """
    Creates (or updates, if it already exists) a CFS configuration with the specified name and layers.
    If no layers are specified, an empty list is used.
    """
    if layers == None:
        layers = list()
    data = { "layers": layers }
    if use_api:
        url = cfs_configs_url(name)
        response_object = requests_put(url, json=data)
    else:
        with tempfile.NamedTemporaryFile(mode="wt", encoding="ascii", prefix="cfs-config-", 
                                         suffix=".tmp", delete=True) as f:
            json.dump(data, f)
            f.flush()
            run_cmd_list(["cat", f.name])
            cmd_list = ["cfs","configurations","update",name,"--file",f.name]
            response_object = run_cli_cmd(cmd_list)
    validate_cfs_config(response_object, expected_name=name, expected_layers=layers)
    if verify_create:
        return describe_cfs_config(use_api=use_api, name=name, expected_layers=layers)
    return response_object

def create_cfs_config_with_appended_layer(use_api, base_config_name, new_config_name, layer_clone_url, 
                                          layer_commit, layer_playbook, layer_name):
    """
    Creates a new CFS config with the specified name. This config is identical to the specified base config
    except it has a layer appended to it with the specified values.
    """
    base_config = describe_cfs_config(use_api=use_api, name=base_config_name)
    try:
        layers = base_config["layers"]
    except KeyError:
        # I think configs will always have this field, but the spec doesn't
        # assure us of that, so we'll play it safe
        layers = list()
    new_layer = {
        "cloneUrl": layer_clone_url,
        "commit": layer_commit,
        "name": layer_name,
        "playbook": layer_playbook }
    layers.append(new_layer)
    return create_cfs_config(use_api=use_api, name=new_config_name, layers=layers)