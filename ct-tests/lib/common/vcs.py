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
VCS-related test helper functions for CMS tests
"""

from .api import API_HOSTNAME, URL_BASE, check_response, do_request, get_response_json, \
                 json_content_headers
from .k8s import get_vcs_username_password
from .helpers import debug, get_bool_field_from_obj, get_dict_field_from_obj, \
                     get_field_from_obj, get_str_field_from_obj, \
                     info, raise_test_error, \
                     raise_test_exception_error, run_cmd_list, run_git_cmd_in_repo
import random
import requests

VCS_URL_BASE = URL_BASE + "/vcs/api/v1"

def get_vcs_repo_git_url(orgname, reponame, with_auth=True):
    """
    Returns URL string of repo (for example, for a git clone)
    """
    try:
        vcsuser, vcspass = get_vcs_username_password()
    except Exception as e:
        raise_test_exception_error(e, "to get vcs username and password from kubernetes")
    if with_auth:
        return "https://%s:%s@%s/vcs/%s/%s.git" % (vcsuser, vcspass, API_HOSTNAME, orgname, reponame)
    else:
        return "%s/vcs/%s/%s.git" % (URL_BASE, orgname, reponame)

def vcs_org_url(orgname=None):
    """
    Returns VCS endpoint for orgs in general or the specified org
    """
    url = VCS_URL_BASE + "/orgs"
    if orgname == None:
        return url
    return url + "/" + orgname

def vcs_repo_url(orgname, reponame):
    """
    Returns VCS endpoint for the specified repo in the specified org
    """
    return VCS_URL_BASE + "/repos/" + orgname + "/" + reponame

def vcs_request(method, url, expected_sc=None, return_json=False, **kwargs):
    try:
        vcsuser, vcspass = get_vcs_username_password()
    except Exception as e:
        raise_test_exception_error(e, "to get vcs username and password from kubernetes")
    auth = (vcsuser, vcspass)
    if "json" in kwargs:
        try:
            headers = kwargs["headers"]
        except KeyError:
            headers = None
        kwargs["headers"] = json_content_headers(headers=headers)
    resp = do_request(method=method, url=url, auth=auth, **kwargs)
    if expected_sc == None:
        if return_json:
            return get_response_json(resp)
        return resp
    return check_response(resp, expected_sc=expected_sc, return_json=return_json)

def vcs_get(url, expected_sc=200, return_json=True, **kwargs):
    return vcs_request(method=requests.get, url=url, expected_sc=expected_sc, 
                       return_json=return_json, **kwargs)

def vcs_post(url, expected_sc=201, return_json=True, **kwargs):
    return vcs_request(method=requests.post, url=url, expected_sc=expected_sc, 
                       return_json=return_json, **kwargs)

def vcs_delete(url, expected_sc=204, return_json=False, **kwargs):
    return vcs_request(method=requests.delete, url=url, expected_sc=expected_sc, 
                       return_json=return_json, **kwargs)

org_fieldname_function_map = dict()

def validate_vcs_org(org_object, **expected_values):
    if "username" in expected_values:
        orgname = get_str_field_from_obj(org_object, "username", noun="VCS org", exact_value=expected_values["username"])
    else:
        orgname = get_str_field_from_obj(org_object, "username", noun="VCS org", min_length=1)
    for fieldname, fieldvalue in expected_values.items():
        if fieldname == "username":
            continue
        try:
            get_func = org_fieldname_function_map[fieldname]
        except KeyError:
            get_func = get_field_from_obj
        get_func(org_object, fieldname, noun="VCS org", exact_value=fieldvalue)
    info("VCS org %s has expected values" % orgname)
    return

def create_vcs_org(orgname=None, visibility="public", description="Test org created by CMS test", 
                   query_create=True):
    """
    Create a VCS org with the specified name. If no name is specified, one is generated.
    If query_create is True, then query VCS for the newly-created org, and return the JSON object from that query.
    If query_create is False, then return the JSON object from the create response.
    """
    if orgname == None:
        orgname = "cms-test-org-%d" % random.randint(0,9999999)
        info("Randomly generated org name: %s" % orgname)
    url = vcs_org_url()
    data = { 
        "username": orgname, 
        "visibility": visibility, 
        "description": description }
    resp = vcs_post(url, json=data)
    validate_vcs_org(resp, **data)
    if query_create:
        info("Query VCS for newly-created org to verify it now exists")
        resp = get_vcs_org(orgname)
        validate_vcs_org(resp, **data)
        return resp
    else:
        return resp

def get_vcs_org(orgname, expect_to_pass=True):
    """
    Queries VCS about the specified org
    If we expect it to pass, and it does, then we return the org object
    If we expect it to fail, and it does, then we return the response object
    Otherwise an error is raised (from within a function we call)
    """
    url = vcs_org_url(orgname)
    if expect_to_pass:
        return vcs_get(url)
    else:
        return vcs_get(url, expected_sc=404, return_json=False)

def delete_vcs_org(orgname, query_delete=True):
    url = vcs_org_url(orgname)
    resp = vcs_delete(url)
    if query_delete:
        return get_vcs_org(orgname, expect_to_pass=False)
    return resp

repo_fieldname_function_map = {
    "private": get_bool_field_from_obj }

def validate_vcs_repo(repo_object, expected_org_values=None, **expected_values):
    if "name" in expected_values:
        reponame = get_str_field_from_obj(repo_object, "name", noun="VCS repo", exact_value=expected_values["name"])
    else:
        reponame = get_str_field_from_obj(repo_object, "name", noun="VCS repo", min_length=1)
    for fieldname, fieldvalue in expected_values.items():
        if fieldname == "name":
            continue
        try:
            get_func = repo_fieldname_function_map[fieldname]
        except KeyError:
            get_func = get_field_from_obj
        get_func(repo_object, fieldname, noun="VCS repo", exact_value=fieldvalue)
    if expected_org_values == None:
        expected_org_values = dict()
    min_org_fields = len(expected_org_values)
    if "username" not in expected_org_values:
        min_org_fields += 1
    org_object = get_dict_field_from_obj(repo_object, "owner", noun="VCS repo", min_length=min_org_fields)
    info("Validating VCS org owner of VCS repo %s" % reponame)
    validate_vcs_org(org_object, **expected_org_values)
    info("VCS repo %s has expected values" % reponame)
    return

def create_vcs_repo(orgname, reponame=None, auto_init=True, description="Test repo created by CMS test", 
                    gitignores=None, license=None, private=False, readme=None, query_create=True,
                    default_branch="main", expected_org_values=None):
    """
    Creates a VCS repo with the specified name (or a randomly generated one) in the specified org.
    Returns the object that VCS returns on a successful create
    """
    if reponame == None:
        reponame = "cms-test-repo-%d" % random.randint(0,9999999)
        info("Randomly generated repo name: %s" % reponame)
    url = VCS_URL_BASE + "/org/" + orgname + "/repos"
    data = { 
        "name": reponame, 
        "auto_init": auto_init, 
        "default_branch": default_branch,
        "description": description, 
        "gitignores": gitignores, 
        "license": license, 
        "private": private, 
        "readme": readme }
    resp = vcs_post(url, json=data)
    if expected_org_values == None:
        expected_org_values = { "username": orgname }
    elif "username" not in expected_org_values:
        expected_org_values["username"] = orgname
    # The following parameters are only used for the creation and are not expected to be found as fields in the
    # repo object itself: auto_init, gitignores, license, readme
    del data["auto_init"]
    del data["gitignores"]
    del data["license"]
    del data["readme"]
    # In the initial response, the only field which appears to be reliably populated is the name
    validate_vcs_repo(resp, expected_org_values=expected_org_values, name=reponame)
    if query_create:
        info("Query VCS for newly-created org to verify it now exists")
        resp = get_vcs_repo(orgname, reponame)
        validate_vcs_repo(resp, expected_org_values=expected_org_values, **data)
        return resp
    else:
        return resp

def get_vcs_repo(orgname, reponame, expect_to_pass=True):
    """
    Queries VCS about the specified repo in the specified org
    If we expect it to pass, and it does, then we return the repo object
    If we expect it to fail, and it does, then we return the response object
    Otherwise an error is raised (from within a function we call)
    """
    url = vcs_repo_url(orgname, reponame)
    if expect_to_pass:
        return vcs_get(url)
    else:
        return vcs_get(url, expected_sc=404, return_json=False)

def delete_vcs_repo(orgname, reponame, query_delete=True):
    url = vcs_repo_url(orgname, reponame)
    resp = vcs_delete(url)
    if query_delete:
        return get_vcs_repo(orgname, reponame, expect_to_pass=False)
    return resp

def clone_vcs_repo(orgname, reponame, tmpdir):
    """
    Clone the specified VCS repo into the specified temporary directory.
    Sets user.name and user.email for that repo.
    Returns the repo directory.
    """
    url = get_vcs_repo_git_url(orgname, reponame)
    git_repo_dir="%s/%s_%s" % (tmpdir, orgname, reponame)
    info("Cloning vcs repo %s in org %s to directory %s" % (reponame, orgname, git_repo_dir))
    run_cmd_list(["git", "clone", url, git_repo_dir])
    debug("Setting user.email for cloned repo")
    run_git_cmd_in_repo(git_repo_dir, "config", "user.email", "catfood@dogfood.mil")
    debug("Setting user.name for cloned repo")
    run_git_cmd_in_repo(git_repo_dir, "config", "user.name", "Rear Admiral Joseph Catfood")
    return git_repo_dir

def create_and_clone_vcs_repo(orgname, reponame, tmpdir, testname=None):
    """
    Creates a VCS org
    Create a repo in that org
    Clones that repo into a subdirectory of tmpdir
    Returns the cloned repo directory path
    """
    if testname == None:
        description = "Created by CMS VCS test library"
    else:
        description = "Created by CMS %s test" % testname
    info("Creating VCS org %s" % orgname)
    create_vcs_org(description=description, orgname=orgname)
    info("Created VCS org %s" % orgname)
    
    info("Creating VCS repo %s in org %s" % (reponame, orgname))
    create_vcs_repo(orgname=orgname, reponame=reponame, description=description)
    info("Created VCS repo %s in org %s" % (reponame, orgname))
    
    repodir = clone_vcs_repo(orgname=orgname, reponame=reponame, tmpdir=tmpdir)
    return repodir
