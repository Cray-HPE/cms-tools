# Copyright 2020 Hewlett Packard Enterprise Development LP

"""
CMS test helper functions for CLI
"""

from .api import get_auth_token
from .helpers import info, raise_test_exception_error, run_cmd_list
import json
import tempfile

auth_tempfile = None

def int_list_to_str(lst):
    """
    Given a list of integers: [1, 10, 3, ... ]
    Return it as a string of the form: "[1,10,3,...]"
    """
    return "[%s]" % ','.join([str(l) for l in lst])

def generate_cli_auth_file(prefix=None):
    """
    Get an auth token for the CLI, dump it into a file, and return the filename
    """
    global auth_tempfile
    if prefix == None:
        prefix = "cms-test-cli-auth-file-"
    else:
        prefix = "%s-cli-auth-file-" % prefix
    # Generate CLI auth token
    info("Generating CLI auth token file")
    auth_token = get_auth_token()
    with tempfile.NamedTemporaryFile(mode="wt", prefix=prefix, delete=False) as f:
        auth_tempfile = f.name
        json.dump(auth_token, f)
    info("CLI auth token file successfully created: %s" % auth_tempfile)
    return auth_tempfile

def auth_file():
    """
    If one has been previously generated, return the cli auth file. Otherwise,
    generate one and return its name.
    """
    if auth_tempfile != None:
        return auth_tempfile
    return generate_cli_auth_file()

def run_cli_cmd(cmdlist, parse_json_output=True, return_rc=False, return_json_output=None):
    """
    Run the specified CLI command, prepending cray and appending "--format json --token <authtoken>"
    Parse and return the json_object if specified. Otherwise return the command output and, if specified,
    its return code.
    """
    if return_json_output == None:
        return_json_output = parse_json_output
    run_cmdlist = [ "cray" ]
    run_cmdlist.extend(cmdlist)
    run_cmdlist.extend(["--format", "json", "--token", auth_file() ])
    cmdresp = run_cmd_list(run_cmdlist, return_rc=return_rc)
    if parse_json_output:
        try:
            json_obj = json.loads(cmdresp["out"])
        except Exception as e:
            raise_test_exception_error(e, "to decode a JSON object in the CLI output")
        if return_json_output:
            return json_obj
        cmdresp["json"] = json_obj
    return cmdresp
