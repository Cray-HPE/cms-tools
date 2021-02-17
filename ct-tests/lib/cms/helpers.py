# Copyright 2020-2021 Hewlett Packard Enterprise Development LP

"""
General CMS test helper functions
"""

import datetime
import logging
import os
import shutil
import subprocess
import sys
import tempfile
import time
import traceback

DEFAULT_LOGFILE_DIRECTORY = "/opt/cray/tests/logs/cms"

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)

verbose_output = False

logfile_path = None

TMPDIR_BASE_DIRECTORY="/tmp"
TMPDIR_NAME_PREFIX="cms-test-tmpdir-"
TMPDIR_PREFIX="%s/%s" % (TMPDIR_BASE_DIRECTORY, TMPDIR_NAME_PREFIX)

class CMSTestError(Exception):
    def __init__(self, msg, log_error=True):
        if msg and log_error:
            error(msg)
        super(CMSTestError, self).__init__(msg)

def raise_test_error(msg, log_error=True):
    """
    Raise a CMSTestError with the specified message.
    """
    raise CMSTestError(msg, log_error=log_error)

#
# Test logging / output functions
#

def _timestamped(msg, timestamp=True):
    if timestamp:
        return "%s %s" % (datetime.datetime.utcnow().strftime("%Y%m%d-%H%M%S.%f"), msg)
    else:
        return msg

def info(msg, timestamp=True):
    logger.info(msg)
    print(_timestamped(msg, timestamp))

def section(sname):
    divstr="#"*80
    print("")
    print(divstr)
    info(sname, timestamp=False)
    print(datetime.datetime.now())
    print(divstr)
    print("")

def subtest(sname):
    section("Subtest: %s" % sname)

def debug(msg, timestamp=True):
    logger.debug(msg)
    if verbose_output:
        print(_timestamped(msg, timestamp))

def warn(msg, prefix=True, timestamp=True):
    logger.warn(msg)
    if prefix:
        print(_timestamped("WARNING: %s" % msg, timestamp))
    else:
        print(_timestamped(msg, timestamp))

def error(msg, prefix=True, timestamp=True):
    logger.error(msg)
    if prefix:
        sys.stderr.write(_timestamped("ERROR: %s\n" % msg, timestamp))
    else:
        sys.stderr.write(_timestamped("%s\n" % msg, timestamp))

def init_logger(test_name, logfile_directory=DEFAULT_LOGFILE_DIRECTORY, verbose=False):
    global logfile_path, verbose_output
    verbose_output = verbose
    logfile_directory = "%s/%s" % (logfile_directory, test_name)
    if not os.path.isdir(logfile_directory):
        if os.path.exists(logfile_directory):
            raise_test_error("Logfile directory (%s) exists but is not a directory" % logfile_directory)
        print(_timestamped("Logfile directory (%s) does not exist -- creating it" % logfile_directory))
        os.makedirs(logfile_directory)
    logfile_path = "%s/%s.log" % (
        logfile_directory,
        datetime.datetime.utcnow().strftime("%Y%m%d-%H%M%S.%f"))
    logfile_handler = logging.FileHandler(filename=logfile_path)
    logger.addHandler(logfile_handler)
    logfile_formatter = logging.Formatter(fmt="#LOG#|%(asctime)s.%(msecs)03d|%(levelname)s|%(message)s",datefmt="%Y-%m-%d_%H:%M:%S")
    logfile_handler.setFormatter(logfile_formatter)
    print(_timestamped("Logging to: %s" % logfile_path))
    if verbose_output:
        print(_timestamped("Verbose output enabled"))

#
# Test fatal exit functions
#

def exit_test(rc=0):
    if logfile_path:
        print(_timestamped("Test logfile: %s" % logfile_path))
    sys.exit(rc)

def error_exit(msg="Error encountered. Exiting.", log_error=True):
    """
    Log the specified error message (unless specified not to), then exit.
    """
    if log_error:
        error(msg)
    info("FAILED")
    section("Test failed")
    exit_test(1)

def log_exception_error(err, attempting=None):
    """
    Log an error for the specified exception (including the optional string describing
    the operation being attempted at the time of the error).
    Return a string describing the error.
    """
    if attempting != None:
        msg = "%s attempting %s: %s" % (str(type(err)), attempting, str(err))
    else:
        msg = "Unexpected error encountered: %s: %s" % (str(type(err)), str(err))
    error(msg)
    for line in traceback.format_exception(type(err), err, err.__traceback__):
        info(line.rstrip(), timestamp=False)
    return msg

def raise_test_exception_error(err, attempting=None):
    """
    Raise a CMSTestError for the specified exception (including the optional string describing
    the operation being attempted at the time of the exception).
    """
    msg = log_exception_error(err, attempting)
    raise CMSTestError(msg, log_error=False) from err

#
# Utility functions
#

def sleep(seconds):
    """
    Log that we are about to wait for the specified number of seconds,
    then do so
    """
    if seconds == 1:
        debug("Waiting for 1 second")
    else:
        debug("Waiting for {} seconds".format(seconds))
    time.sleep(seconds)

def create_tmpdir():
    """
    Create a temporary directory and return its name
    """
    t = tempfile.mkdtemp(prefix=TMPDIR_NAME_PREFIX, dir=TMPDIR_BASE_DIRECTORY)
    info("Created temporary directory %s" % t)
    return t

def remove_tmpdir(tmpdir):
    """
    Remove the specified temporary directory if it exists and is in the
    expected location
    """
    if tmpdir == None:
        debug("No temporary directory defined")
        return
    elif not os.path.exists(tmpdir):
        debug("Temporary directory (%s) does not exist" % str(tmpdir))
        return
    elif not os.path.isdir(tmpdir):
        warn("Temporary directory (%s) exists but is not a directory" % str(tmpdir))
        return
    elif tmpdir.find(TMPDIR_PREFIX) != 0:
        warn("Unexpected value of tmpdir variable (%s) -- not removing" % str(tmpdir))
        return
    shutil.rmtree(tmpdir)
    info("Temporary directory %s removed" % tmpdir)

def cmd_failed_msg(cmd_string, rc, stdout, stderr):
    msg = "%s command failed with return code %d" % (cmd_string, rc)
    error(msg)
    info("Command stdout:\n%s" % stdout, timestamp=False)
    info("Command stderr:\n%s" % stderr, timestamp=False)
    return msg

def do_run_cmd(cmd_list, cmd_string, show_output=None, return_rc=False, cwd=None, env_var=None):
    """ 
    Runs the specified command, then displays, logs, and returns the output
    """
    if show_output == None:
        show_output = verbose_output
    info("Running command: %s" % cmd_string)
    run_kwargs = {
        "stdout": subprocess.PIPE,
        "stderr": subprocess.PIPE,
        "check": not return_rc }
    if env_var != None:
        run_kwargs['env'] = env_var
    if cwd != None:
        run_kwargs["cwd"] = cwd
    if show_output:
        output = info
    else:
        output = debug
    try:
        cmddone = subprocess.run(cmd_list, **run_kwargs)
        cmdout = cmddone.stdout.decode()
        cmderr = cmddone.stderr.decode()
        cmdrc = cmddone.returncode
        output("Command return code: %d" % cmdrc)
        output("Command stdout:\n%s" % cmdout, timestamp=False)
        output("Command stderr:\n%s" % cmderr, timestamp=False)
        if return_rc:
            return { "rc": cmdrc, "out": cmdout, "err": cmderr }
        return { "out": cmdout }
    except subprocess.CalledProcessError as e:
        msg = cmd_failed_msg(
            cmd_string=cmd_string, 
            rc=e.returncode, 
            stdout=e.stdout.decode(), 
            stderr=e.stderr.decode())
        raise CMSTestError(msg, log_error=False) from e

def run_cmd_list(cmd_list, **kwargs):
    """
    Wrapper for do_run_cmd that accepts a command in list format
    """
    cmd_string = " ".join(cmd_list)
    return do_run_cmd(cmd_list=cmd_list, cmd_string=cmd_string, **kwargs)

def run_cmd(cmd_string, **kwargs):
    """
    Wrapper for do_run_cmd that accepts a command in string format
    """
    cmd_list = cmd_string.split()
    return do_run_cmd(cmd_list=cmd_list, cmd_string=cmd_string, **kwargs)

def validate_list(val, noun="object", min_length=None, max_length=None, exact_length=None, 
                  member_type=None, prefix=None, show_val_on_error=False):
    """
    Validates that obj is a list which meets all of the specified criteria.
    Raises an error if not.
    """
    def _error(msg):
        if show_val_on_error:
            msg = "%s: %s" % (msg, str(val))
        raise CMSTestError(msg)
    
    if not isinstance(val, list):
        _error("%s should be a list but it is type %s" % (noun, str(type(val))))
    if min_length != None and len(val) < min_length:
        _error("%s should have length >= %d but its length is %d" % (noun, min_length, len(val)))
    elif max_length != None and len(val) > max_length:
        _error("%s should have length <= %d but its length is %d" % (noun, max_length, len(val)))
    elif exact_length != None and len(val) != exact_length:
        _error("%s should have length %d but its length is %d" % (noun, exact_length, len(val)))
    elif member_type != None and not all(isinstance(m, member_type) for m in val):
        _error("%s should have members of type %s but at least one member is not" % (noun, str(member_type)))
    elif prefix != None and val[:len(prefix)] != prefix:
        _error("%s should begin with '%s' but it does not" % (noun, str(prefix)))

def get_field_from_obj(obj, field, noun="response object", expected_type=None, 
                       show_val_on_error=False, null_okay=None, 
                       min_value=None, max_value=None, exact_value=None,
                       min_length=None, max_length=None, exact_length=None,
                       key_type=None, value_type=None, member_type=None, prefix=None):
    """
    Returns the requested field from the specified object, raising a fatal error
    if it is not found. Performs all specified checks, raising errors if any of them fail.
    """
    if null_okay == None:
        if exact_value != None or min_length != None or exact_length != None or prefix != None:
            null_okay = False
        else:
            null_okay = True

    def _error(msg):
        if show_val_on_error:
            msg = "%s: %s" % (msg, str(val))
        raise CMSTestError(msg)

    if not isinstance(obj, dict):
        _error("%s should be dict but is type %s" % (noun, str(type(obj))))
    try:
        val = obj[field]
    except KeyError:
        _error("No '%s' field found in %s" % (field, noun))
    if null_okay and val == None:
        return val

    field_in_noun = "'%s' field in %s" % (field, noun)

    if expected_type == list:
        validate_list(val=val, noun=field_in_noun, show_val_on_error=show_val_on_error, min_length=min_length, 
                      max_length=max_length, exact_length=exact_length, member_type=member_type, prefix=prefix)
    elif expected_type != None and not isinstance(val, expected_type):
        _error("%s should be type %s but it is type %s" % (field_in_noun, str(expected_type), str(type(val))))
    if min_length != None and len(val) < min_length:
        _error("%s should have length >= %d but its length is %d" % (field_in_noun, min_length, len(val)))
    elif max_length != None and len(val) > max_length:
        _error("%s should have length <= %d but its length is %d" % (field_in_noun, max_length, len(val)))
    elif exact_length != None and len(val) != exact_length:
        _error("%s should have length %d but its length is %d" % (field_in_noun, exact_length, len(val)))
    elif key_type != None and not all(isinstance(k, key_type) for k in val.keys()):
        _error("%s should have keys of type %s but at least one key is not" % (field_in_noun, str(key_type)))
    elif value_type != None and not all(isinstance(v, value_type) for v in val.values()):
        _error("%s should have values of type %s but at least one value is not" % (field_in_noun, str(value_type)))
    elif member_type != None and not all(isinstance(m, member_type) for m in val):
        _error("%s should have members of type %s but at least one member is not" % (field_in_noun, str(member_type)))
    elif prefix != None and val[:len(prefix)] != prefix:
        _error("%s should begin with '%s' but it does not" % (field_in_noun, str(prefix)))
    elif min_value != None and val < min_value:
        _error("%s should be >= '%s' but it is not" % (field_in_noun, str(min_value)))
    elif max_value != None and val > max_value:
        _error("%s should be <= '%s' but it is not" % (field_in_noun, str(max_value)))
    elif exact_value != None and val != exact_value:
        _error("%s should be '%s' but it is not" % (field_in_noun, str(exact_value)))
    return val

def get_bool_field_from_obj(obj, field, **kwargs):
    """
    Wrapper for get_field_from_obj for boolean fields
    """
    return get_field_from_obj(obj, field, expected_type=bool, **kwargs)

def get_dict_field_from_obj(obj, field, key_type=str,**kwargs):
    """
    Wrapper for get_field_from_obj for dict fields
    """
    return get_field_from_obj(obj, field, expected_type=dict, key_type=key_type, **kwargs)

def get_int_field_from_obj(obj, field, **kwargs):
    """
    Wrapper for get_field_from_obj for int fields
    """
    return get_field_from_obj(obj, field, expected_type=int, **kwargs)

def get_list_field_from_obj(obj, field, **kwargs):
    """
    Wrapper for get_field_from_obj for list fields
    """
    return get_field_from_obj(obj, field, expected_type=list, **kwargs)

def get_str_field_from_obj(obj, field, **kwargs):
    """
    Wrapper for get_field_from_obj for str fields
    """
    return get_field_from_obj(obj, field, expected_type=str, **kwargs)

def any_dict_value(d):
    """
    Returns a value from the specified dictionary. This is intended for use with dictionaries of one element, so
    we do not attempt to make this a random choice.
    """
    for v in d.values():
        return v

def is_pingable(target):
    """
    Returns True if target is pingable, False otherwise.
    """
    ping_cmd_list = [ "ping", "-q", "-c", "3", target ]
    ping_cmd_resp = run_cmd_list(ping_cmd_list, return_rc=True)
    return ping_cmd_resp["rc"] == 0

def run_scp(local_file, target_host, remote_target=None, scp_arg_list=None, user="root", 
            identity_file=None, port=None, strict_host_key_check=False):
    """
    Runs the necessary scp command to copy the local file to the specified
    remote destination. If remote_target is not specified, it will be
    equal to local_file
    """
    if remote_target == None:
        remote_target = local_file
    info("Using scp to copy %s to %s:%s" % (local_file, target_host, remote_target))
    scp_command_list = [ "scp" ]
    if scp_arg_list != None:
        debug("Extra scp args = %s" % str(scp_arg_list))
        scp_command_list.extend(scp_arg_list)
    if identity_file:
        scp_command_list.extend([ "-i", identity_file ])
    if port:
        scp_command_list.extend([ "-p", str(port) ])
    if not strict_host_key_check:
        scp_command_list.extend(["-o", "StrictHostKeyChecking=no" ])
    scp_command_list.extend([
        local_file, 
        "%s@%s:%s" % (user, target_host, remote_target)])
    run_cmd_list(scp_command_list, return_rc=False)

def run_command_via_ssh(target, cmdstring, user="root", identity_file=None, port=None, 
                        strict_host_key_check=False, **kwargs):
    """
    Runs the necessary ssh command to run the specified command on the specified target.
    """
    ssh_cmd_list = [ "ssh" ]
    if identity_file:
        ssh_cmd_list.extend([ "-i", identity_file ])
    if port:
        ssh_cmd_list.extend([ "-p", str(port) ])
    if not strict_host_key_check:
        ssh_cmd_list.extend([ "-o", "StrictHostKeyChecking=no"])
    ssh_cmd_list.extend([ "%s@%s" % (user, target), cmdstring ])
    return run_cmd_list(ssh_cmd_list, **kwargs)

def ssh_command_passes(target, cmdstring, **kwargs):
    """
    Runs the specified command via ssh on the specified target. 
    Returns True if this succeeds (both the ssh and the command), False otherwise.
    """
    ssh_cmd_resp = run_command_via_ssh(target, cmdstring, return_rc=True, **kwargs)
    return ssh_cmd_resp["rc"] == 0
