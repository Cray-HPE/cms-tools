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
HSM-related CMS test helper functions
"""

from .api import API_URL_BASE, requests_delete, requests_get, requests_post
from .cli import run_cli_cmd
from .helpers import get_dict_field_from_obj, get_list_field_from_obj, \
                     get_str_field_from_obj, info, raise_test_error
import datetime
import string

HSM_URL_BASE = "%s/smd/hsm/v1" % API_URL_BASE
HSM_GROUPS_URL_BASE = "%s/groups" % HSM_URL_BASE
HSM_GROUPS_LABELS_URL_BASE = "%s/labels" % HSM_GROUPS_URL_BASE
HSM_INVENTORY_URL_BASE = "%s/Inventory" % HSM_URL_BASE
HSM_HW_INVENTORY_URL_BASE = "%s/Hardware" % HSM_INVENTORY_URL_BASE

# In addition to alphanumeric characters, the following are also legal:
# .:-_ (period, colon, hyphen, underscore)
VALID_HSM_GROUP_CHARACTERS = string.ascii_letters + string.digits + ".:-_"

def hsm_groups_url(groupname=None):
    """
    Returns the url endpoint for hsm groups generally, or the specified group
    """
    if groupname != None:
        return "%s/%s" % (HSM_GROUPS_URL_BASE, groupname)
    else:
        return HSM_GROUPS_URL_BASE

def hsm_groups_members_url(groupname, xname=None):
    """
    Returns the url endpoint for members of the specified group
    """
    base_url = "%s/members" % hsm_groups_url(groupname)
    if xname != None:
        return "%s/%s" % (base_url, xname)
    return base_url

def get_hsm_xname_list(use_api, populated_only=True):
    """
    Return all Node xnames from the HSM hardware inventory.
    """
    if use_api:
        data = { 'type': 'Node' }
        json_object = requests_get(HSM_HW_INVENTORY_URL_BASE, json=data)
    else:
        json_object = run_cli_cmd("hsm inventory hardware list --type Node".split())

    if not isinstance(json_object, list):
        raise_test_error("We expect hsm response to be a list, but it is %s: %s" % (str(type(json_object)), str(json_object)))

    xname_list = list()
    noun="hsm hardware inventory list entry"
    for nodehw in json_object:
        if populated_only:
            node_status = get_str_field_from_obj(nodehw, "Status", noun=noun, min_length=1)
            if node_status != "Populated":
                continue
        node_xname = get_str_field_from_obj(nodehw, "ID", noun=noun, min_length=2, prefix="x")
        xname_list.append(node_xname)
    return xname_list

def add_hsm_group_member(use_api, group_name, xname):
    """
    Adds the specified xname to the specified HSM group
    """
    # Add member
    info("Adding xname %s to HSM group %s" % (xname, group_name))
    if use_api:
        url = hsm_groups_members_url(group_name)
        data = { 'id': xname }
        requests_post(url, json=data)
    else:
        run_cli_cmd(["hsm", "groups", "members", "create", "--id", xname, group_name])

def add_hsm_group_members(use_api, group_name, xname_list):
    """
    Adds the specified xnames to the specified HSM group
    """
    # Add members
    for xname in xname_list:
        add_hsm_group_member(use_api, group_name, xname)

def validate_hsm_group(group_object, expected_label=None, expected_tags=None,
                       expected_description=None, expected_xnames=None,
                       expected_exclusive_group=None):
    """
    Validates that the HSM group object:
    - Is a dictionary.
    
    - Has a non-blank label field (which matches the expected label, if specified).
    
    - Has a description field (which matches the expected description, if specified).
    
    - Has a members field, which is a dictionary containing exactly 1 field, an ids field.
    - The members ids field is a list of unique strings (matching the expected xnames, if
      specified, ignoring order).
    
    - If the tags field is present, it maps to a nonempty list of strings.
    - If the expected tags list is specified and nonempty, we verify that the tags field is
      present and matches the expected tags, ignoring order.
    - If the expected tags list is specified and empty, we verify that the tags field is not
      present.

    - If the exclusiveGroup field is present, it maps to a nonempty string.
    - If the expected exclusive group is specified and nonempty, we verify that the 
      exclusiveGroup field is present and matches the expected value.
    - If the expected exclusive group is specified and empty, we verify that the
      exclusiveGroup field is not present.

    - No other fields are present.
    
    A test exception is raised if any of these are not true.
    """
    noun = "HSM group object"
    if expected_label != None:
        get_str_field_from_obj(group_object, "label", noun=noun, exact_value=expected_label)
    else:
        get_str_field_from_obj(group_object, "label", noun=noun, min_length=1)

    if expected_description != None:
        get_str_field_from_obj(group_object, "description", noun=noun, exact_value=expected_description)
    else:
        get_str_field_from_obj(group_object, "description", noun=noun, min_length=0)

    members = get_dict_field_from_obj(group_object, "members", noun=noun, exact_length=1, key_type=str, value_type=list)
    xnames = get_list_field_from_obj(members, "ids", noun="%s members field" % noun, member_type=str, null_okay=False)
    if expected_xnames != None and sorted(xnames) != sorted(expected_xnames):
        info("We expect HSM group object members field to contain the following xnames: %s" % ', '.join(sorted(expected_xnames)))
        info("However, it contains the following xnames: %s" % ', '.join(sorted(xnames)))
        raise_test_error("We expect HSM group object members field does not match what we expect")

    if "tags" in group_object:
        tags = get_list_field_from_obj(members, "tags", noun="%s tags field" % noun, member_type=str, null_okay=False, min_length=1)
        if expected_tags != None:
            if len(expected_tags) == 0:
                info("We expect HSM group object tags field to not be present")
                info("However, it is present and contains the following tags: %s" % ', '.join(sorted(tags)))
                raise_test_error("The HSM group object tags field does not match what we expect")
            elif sorted(tags) != sorted(expected_tags):
                info("We expect HSM group object tags field to contain the following tags: %s" % ', '.join(sorted(expected_tags)))
                info("However, it contains the following tags: %s" % ', '.join(sorted(tags)))
                raise_test_error("The HSM group object tags field does not match what we expect")
    elif expected_tags:
        info("We expect HSM group object tags field to contain the following tags: %s" % ', '.join(sorted(expected_tags)))
        info("However, the tags field is not present")
        raise_test_error("The HSM group object tags field does not match what we expect")

    if "exclusiveGroup" in group_object:
        exclusiveGroup = get_str_field_from_obj(group_object, "exclusiveGroup", noun=noun, min_length=1)
        if expected_exclusive_group != None:
            if len(expected_exclusive_group) == 0:
                info("We expect HSM group object exclusiveGroup field to not be present")
                info("However, it is present and has value: %s" % exclusiveGroup)
                raise_test_error("The HSM group object exclusiveGroup field does not match what we expect")
            elif exclusiveGroup != expected_exclusive_group:
                info("We expect HSM group object exclusiveGroup field to have value: %s" % expected_exclusive_group)
                info("However, it is present and has value: %s" % exclusiveGroup)
                raise_test_error("The HSM group object exclusiveGroup field does not match what we expect")
    elif expected_exclusive_group:
        info("We expect HSM group object exclusiveGroup field to have value: %s" % expected_exclusive_group)
        info("However, the exclusiveGroup field is not present")
        raise_test_error("The HSM group object exclusiveGroup field does not match what we expect")

    unknown_keys = sorted([ k for k in group_object.keys() if k not in {"description", "exclusiveGroup", "label", "members", "tags"} ])
    if unknown_keys:
        raise_test_error("Unexpected fields found in HSM group object: %s" % ', '.join(unknown_keys))

def describe_hsm_group(use_api, group_name, partition_name=None, expect_to_pass=True, 
                       expected_description=None, expected_tags=None, expected_xnames=None,
                       expected_exclusive_group=None):
    """
    Calls an API GET or CLI describe on the specified HSM group
    If we are expecting it to pass, the HSM group is extracted from the response,
    validated, and returned.
    Otherwise we just validate that the request failed.
    """
    if partition_name == None:
        info("Describing HSM group %s" % group_name)
    else:
        info("Describing HSM group %s intersected with partition %s" % (group_name, partition_name))
    if use_api:
        url = hsm_groups_url(group_name)
        kwargs = dict()
        if partition_name != None:
            kwargs["params"] = { "partition": partition_name }
        if expect_to_pass:
            response_object = requests_get(url, **kwargs)
        else:
            requests_get(url, expected_sc=404, **kwargs)
            return
    else:
        cmd_list = ["hsm","groups","describe",group_name]
        if partition_name != None:
            cmd_list.extend(["--partition",partition_name])
        if expect_to_pass:
            response_object = run_cli_cmd(cmd_list)
        else:
            cmdresp = run_cli_cmd(cmd_list, return_rc=True, parse_json_output=False)
            if cmdresp["rc"] == 0:
                raise_test_error("We expected this query to fail but the return code was 0")
            return

    validate_hsm_group(group_object=response_object, expected_label=group_name, 
                       expected_description=expected_description, 
                       expected_tags=expected_tags, expected_xnames=expected_xnames,
                       expected_exclusive_group=expected_exclusive_group)
    return response_object

def create_hsm_group(use_api, group_name=None, name_prefix=None, description=None, 
                     xname_list=None, tag_list=None, exclusive_group=None, 
                     verify_create=True, test_name="CMS test"):
    """
    Creates an HSM group with the specified values. Returns the name of the new group.
    """
    if group_name == None:
        if name_prefix == None:
            if xname_list:
                if len(xname_list) == 1:
                    name_prefix = "test-group-%s" % xname_list[0]
                else:
                    name_prefix = "test-group-%dxnames" % len(xname_list)
            else:
                name_prefix = "test-group"
        group_name = "%s-%s" % (name_prefix, datetime.datetime.utcnow().strftime("%Y%m%d-%H%M%S.%f"))
    if description == None:
        description = "Created by %s" % test_name
    # Create group
    if use_api:
        data = { "label": group_name, "description": description }
        if tag_list != None:
            data["tags"] = tag_list
        if exclusive_group != None:
            data["exclusiveGroup"] = exclusive_group
        if xname_list == None or len(xname_list) == 0:
            expected_xnames = []
        elif len(xname_list) > 1:
            data["members"] = { "ids": xname_list[:-1] }
            expected_xnames = xname_list[:-1]
            xname_list = xname_list[-1:]
        else:
            # len(xname_list) == 1
            data["members"] = { "ids": xname_list }
            expected_xnames = xname_list
            xname_list = []
        requests_post(hsm_groups_url(), json=data)
    else:
        cli_cmd_list = ["hsm", "groups", "create", "--label", group_name, "--description", description]
        if tag_list != None and len(tag_list) > 0:
            cli_cmd_list.extend(["--tags", ",".join(tag_list)])
        if exclusive_group != None:
            cli_cmd_list.extend(["--exclusive-group", exclusive_group])
        run_cli_cmd(cli_cmd_list)
        expected_xnames = []

    if verify_create:
        info("Verifying that group was created successfully")
        describe_hsm_group(use_api=use_api, group_name=group_name, 
                           expected_description=description, 
                           expected_xnames=expected_xnames,
                           expected_tags=tag_list,
                           expected_exclusive_group=exclusive_group)
    if not xname_list:
        return group_name

    # Add members
    if use_api:
        url = hsm_groups_members_url(group_name)
        for xname in xname_list:
            data = { 'id': xname }
            requests_post(url, json=data)
            expected_xnames.append(xname)
    else:
        for xname in xname_list:
            run_cli_cmd(["hsm", "groups", "members", "create", "--id", xname, group_name])
            expected_xnames.append(xname)

    if verify_create:
        info("Verifying that group members were added successfully")
        describe_hsm_group(use_api=use_api, group_name=group_name, 
                           expected_description=description, 
                           expected_xnames=expected_xnames,
                           expected_tags=tag_list,
                           expected_exclusive_group=exclusive_group)

    return group_name

def delete_hsm_group(use_api, group_name, verify_delete=True):
    """
    Delete specified HSM group. If specified, verify the deletion succeeded by
    trying to retrieve it.
    """
    info("Deleting HSM group %s" % group_name)
    if use_api:
        url = hsm_groups_url(group_name)
        requests_delete(url, expected_sc=200)
    else:
        run_cli_cmd(["hsm","groups","delete",group_name], parse_json_output=False)

    if not verify_delete:
        return
    info("Validating that the group no longer exists")
    describe_hsm_group(use_api=use_api, group_name=group_name, expect_to_pass=False)

def list_hsm_groups(use_api, name_list=None, tag_list=None):
    """
    Return a list of HSM groups matching specified names and tags. If neither are specified,
    returns list of all groups.
    """
    if use_api:
        params = dict()
        if name_list:
            params["group"] = name_list
        if tag_list:
            params["tag"] = tag_list
        response_object = requests_get(hsm_groups_url(), params=params)
    else:
        cli_cmd_list = "hsm groups list".split()
        if name_list:
            groups_matching_names = dict()
            for gname in name_list:
                response_object = run_cli_cmd(cli_cmd_list + ["--group", gname])
                if len(response_object) > 1:
                    raise_test_error("This command should return at most 1 group object, but it returned %d" % len(response_object))
                elif len(response_object) == 1:
                    group_object = response_object[0]
                    validate_hsm_group(group_object=group_object, expected_label=gname)
                    groups_matching_names[gname] = group_object
        if tag_list:
            groups_matching_tags = dict()
            for tag in tag_list:
                response_object = run_cli_cmd(cli_cmd_list + ["--tag", tag])
                for group_object in response_object:
                    validate_hsm_group(group_object=group_object)
                    gname = group_object["label"]
                    if "tags" not in group_object:
                        raise_test_error(
                            "Command should only return groups with tag %s, but at least one (%s) does not have any tags" % (tag, gname))
                    elif tag not in group_object[tags]:
                        raise_test_error(
                            "Command should only return groups with tag %s, but at least one (%s) does not have it" % (tag, gname))
                    groups_matching_tags[gname] = group_object
            if name_list:
                return list(groups_matching_tags.values())
            else:
                # Must return intersection of groups which match our name list and our tag list
                return [ groups_matching_names[gn] for gn in groups_matching_names.keys() if gn in groups_matching_tags.keys() ]
        elif name_list != None:
            return list(groups_matching_names.values())
        response_object = run_cli_cmd(cli_cmd_list)
    for group_object in response_object:
        validate_hsm_group(group_object=group_object)
        if use_api:
            # We have not yet validated against our name_list or tag_list, if any
            if name_list:
                gname = group_object["label"]
                if gname not in name_list:
                    raise_test_error("Command returned group (%s) which did not match specified group name list" % gname)
            if tag_list:
                if "tags" not in group_object:
                    raise_test_error("Command returned group (%s) which has no tags" % gname)
                elif all(tag in group_object["tags"] not in tag_list):
                    raise_test_error("Command returned group (%s) with no tags matching specified tag list" % gname)
    return response_object

def list_hsm_group_members(use_api, group_name, partition_name=None):
    """
    Returns list of xnames which belong to the specified HSM group.
    If a partition is specified, it returns the intersection of the
    HSM group and that partition.
    """
    if partition_name == None:
        info("Retrieving members list for HSM group %s" % group_name)        
    else:
        info("Retrieving intersection of members list for HSM group %s and partition %s" % (group_name, partition_name))
    if use_api:
        url = hsm_groups_members_url(group_name)
        kwargs = dict()
        if partition_name != None:
            kwargs["params"] = { "partition": partition_name }
        response_object = requests_get(url, **kwargs)
    else:
        cmd_list = ["hsm", "groups", "members", "list", group_name]
        if partition_name != None:
            cmd_list.extend(["--partition",partition_name])
        response_object = run_cli_cmd(cmd_list)

    return get_list_field_from_obj(response_object, "ids", noun="HSM group members object", member_type=str, null_okay=False)

def delete_hsm_group_member(use_api, group_name, xname, verify_delete=True):
    """
    Removes the specified xname from the specified HSM group
    """
    # Delete member
    info("Removing xname %s from HSM group %s" % (xname, group_name))
    if use_api:
        url = hsm_groups_members_url(group_name, xname)
        requests_delete(url, expected_sc=200)
    else:
        run_cli_cmd(["hsm", "groups", "members", "delete", xname, group_name])
    if not verify_delete:
        return
    info("Validating that the member no longer belongs to the group")
    member_list = [ x.lower() for x in list_hsm_group_members(use_api, group_name) ]
    if xname.lower() in member_list:
        raise_test_error("xname %s still appears to be in HSM group %s" % (xname, group_name))

def delete_hsm_group_members(use_api, group_name, xname_list, verify_delete=True):
    """
    Removes the specified xnames from the specified HSM group
    """
    if not xname_list:
        # Nothing to remove
        return
    for xname in xname_list:
        delete_hsm_group_member(use_api, group_name, xname, verify_delete)
    if not verify_delete:
        return
    member_list = [ x.lower() for x in list_hsm_group_members(use_api, group_name) ]
    for xname in xname_list:
        if xname.lower() in member_list:
            raise_test_error("xname %s still appears to be in HSM group %s" % (xname, group_name))

def clear_hsm_group_members(use_api, group_name, verify_delete=True):
    """
    Removes all members from the specified HSM group
    """
    delete_hsm_group_members(use_api=use_api, group_name=group_name, 
                             xname_list=list_hsm_group_members(use_api, group_name), 
                             verify_delete=verify_delete)
    if not verify_delete:
        return
    member_list = list_hsm_group_members(use_api, group_name)
    if member_list:
        raise_test_error("We removed all xnames from HSM group %s, but it still has member(s): %s" % (group_name, ", ".join(member_list)))

def list_hsm_group_labels(use_api):
    """
    Returns a list of all HSM group labels
    """
    info("Listing HSM group labels")
    if use_api:
        return requests_get(HSM_GROUPS_LABELS_URL_BASE)
    else:
        return run_cli_cmd("hsm groups labels list".split())

def set_hsm_group_members(use_api, group_name, xname_list):
    """
    Performs operations that result in specified HSM group having exactly the specified 
    xnames as members.
    """
    sorted_xnames = sorted(xname_list)
    info("Set HSM group %s to have member list: %s" % (group_name, ",".join(sorted_xnames)))
    if len(xname_list) == 0:
        clear_hsm_group_members(use_api, group_name)
        return
    current_group_members = list_hsm_group_members(use_api, group_name)
    if sorted(current_group_members) == sorted_xnames:
        info("Group already has specified member list -- nothing to do")
        return
    members_to_add = [ xn for xn in sorted_xnames if xn not in current_group_members ]
    members_to_remove = [ xn for xn in current_group_members if xn not in sorted_xnames ]
    add_hsm_group_members(use_api, group_name, members_to_add)
    delete_hsm_group_members(use_api, group_name, members_to_remove)
    current_group_members = list_hsm_group_members(use_api, group_name)
    if sorted(current_group_members) != sorted_xnames:
        raise_test_error("The current HSM group membership list does not match what we expect")
