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
CMS test helper functions for CLI
"""

from .api import get_auth_token
from .helpers import cmd_failed_msg, info, raise_test_error, \
                     raise_test_exception_error, run_cmd_list, warn
import json
import os
import tempfile

auth_tempfile = None
config_tempfile = None
cray_cli = "/usr/bin/cray"

config_error_strings = [
    "Unable to connect to cray",
    "verify your cray hostname",
    "core.hostname",
    "No configuration exists",
    "cray init" ]

cli_config_file_text = """\
[core]
hostname = "https://api-gw-service-nmn.local"
"""

def int_list_to_str(lst):
    """
    Given a list of integers: [1, 10, 3, ... ]
    Return it as a string of the form: "[1,10,3,...]"
    """
    return "[%s]" % ','.join([str(l) for l in lst])

def generate_cli_config_file(prefix=None):
    """
    Write CLI config file text to a file and return the filename
    """
    global config_tempfile
    if prefix == None:
        prefix = "cms-test-cli-config-file-"
    else:
        prefix = "%s-cli-config-file-" % prefix
    with tempfile.NamedTemporaryFile(mode="wt", prefix=prefix, delete=False) as f:
        config_tempfile = f.name
        f.write(cli_config_file_text)
    info("CLI config file successfully created: %s" % config_tempfile)
    return config_tempfile

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

def config_env():
    env_var = os.environ.copy()
    if config_tempfile:
        env_var["CRAY_CONFIG"] = config_tempfile
    else:
        env_var["CRAY_CONFIG"] = generate_cli_config_file()
    info("Environment variable CRAY_CONFIG set to '%s' for CLI command execution" % env_var["CRAY_CONFIG"])
    return env_var

def run_cli_cmd(cmdlist, parse_json_output=True, return_rc=False, return_json_output=None):
    """
    Run the specified CLI command, prepending cray and appending "--format json --token <authtoken>"
    If it fails with an error that appears config-related, retry the command with our own config file.
    Parse and return the json_object if specified. Otherwise return the command output and, if specified,
    its return code.
    """
    global config_tempfile
    if return_json_output == None:
        return_json_output = parse_json_output
    run_cmdlist = [ cray_cli ]
    run_cmdlist.extend(cmdlist)
    run_cmdlist.extend(["--format", "json", "--token", auth_file() ])
    if config_tempfile:
        # Specify our own config file via environment variable
        cmdresp = run_cmd_list(run_cmdlist, return_rc=return_rc, env_var=config_env())
    else:
        # Let's try CLI command without our own config file first.
        cmdresp = run_cmd_list(run_cmdlist, return_rc=True)
        if cmdresp["rc"] != 0:
            if any(estring in cmdresp["err"] for estring in config_error_strings):
                info("CLI command failure may be due to configuration error. Will generate our own config file and retry")
                cmdresp = run_cmd_list(run_cmdlist, return_rc=True, env_var=config_env())
                if cmdresp["rc"] != 0:
                    info("CLI command failed even using our CLI config file.")
                    # No need to continue using our config file                    
                    config_tempfile = None
                    if not return_rc:
                        msg = cmd_failed_msg(
                            cmd_string=" ".join(run_cmdlist), 
                            rc=cmdresp["rc"], 
                            stdout=cmdresp["out"], 
                            stderr=cmdresp["err"])
                        raise_test_error(msg, log_error=False)
                elif not return_rc:
                    # Command passed but user did not request return code in the response, so let's remove it
                    del cmdresp["rc"]
            else:
                info("CLI command failed and does not appear to be obviously related to the CLI config")
                if not return_rc:
                    msg = cmd_failed_msg(
                        cmd_string=" ".join(run_cmdlist), 
                        rc=cmdresp["rc"], 
                        stdout=cmdresp["out"], 
                        stderr=cmdresp["err"])
                    raise_test_error(msg, log_error=False)
        elif not return_rc:
            # Command passed but user did not request return code in the response, so let's remove it
            del cmdresp["rc"]

    if parse_json_output:
        try:
            json_obj = json.loads(cmdresp["out"])
        except Exception as e:
            raise_test_exception_error(e, "to decode a JSON object in the CLI output")
        if return_json_output:
            return json_obj
        cmdresp["json"] = json_obj
    return cmdresp
