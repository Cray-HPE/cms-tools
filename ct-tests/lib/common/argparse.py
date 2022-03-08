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
Argument parsing helper functions for CMS tests.
"""

import argparse
import string
from .hsm import VALID_HSM_GROUP_CHARACTERS

def nonblank(argname, argvalue):
    """
    If argvalue is empty, raise an exception. Otherwise return it.
    """
    if not argvalue:
        raise argparse.ArgumentTypeError("%s may not be blank" % argname)
    return argvalue

def positive_integer(argname, argvalue):
    """
    If argvalue is not a string representing a positive integer, 
    raise an exception. Otherwise return it (as an integer).
    """
    try:
        int_value = int(argvalue)
        if int_nid < 1:
            raise argparse.ArgumentTypeError("%s must be a positive integer. Invalid (not positive): %s" % argvalue)
        return int_value
    except ValueError:
        raise argparse.ArgumentTypeError("%s must be a positive integer. Invalid (not integer): %s" % argvalue)

def valid_string(argname, s, min_string_length=None, 
                 max_string_length=None, valid_characters=None, 
                 invalid_characters=None):
    """
    Raise an exception if:
        - s is shorter than the minimum string length (if specified)
        - s is longer than the maximum string length (if specified)
        - s contains a character that is not in the valid characters
          set (if specified)
        - s contains a character that is in the invalid characters
          set (if specified)
    Otherwise return s.
    """
    if min_string_length and len(s) < min_string_length:
        raise argparse.ArgumentTypeError(
            "Every %s must be at least %d characters long. Invalid: %s" % (
                argname, min_string_length, s))
    elif max_string_length and len(s) > max_string_length:
        raise argparse.ArgumentTypeError(
            "Every %s must be at most %d characters long. Invalid: %s" % (
                argname, max_string_length, s))
    bad_chars = None
    if valid_characters:
        bad_chars = [ c for c in s if c not in valid_characters ]
    if invalid_characters and not bad_chars:
        bad_chars = [ c for c in s if c in invalid_characters ]
    if bad_chars:
        raise argparse.ArgumentTypeError(
            "%s '%s' contains invalid character(s): %s" % (
                argname, s, ' '.join(bad_chars)))
    return s

def valid_string_list(argname, s, delimiter=None, 
                      min_list_length=1, max_list_length=None, 
                      remove_duplicates=True, sort_list=True,
                      min_string_length=1, max_string_length=None, 
                      valid_characters=None, invalid_characters=None):
    """
    Breaks string s into a list of strings, using the specified delimiter (if
    delimiter is None, whitespace is the delimiter). Sorts and removes
    duplicates, if specified.
    Raises errors if:
        - Any string is shorter than the minimum string length (if specified)
        - Any string is longer than the maximum string length (if specified)
        - Any string contains a character that is not in the valid characters
          set (if specified)
        - Any string contains a character that is in the invalid characters
          set (if specified)
        - The list is shorter than the minimum list length (if specified)
        - The list if longer than the maximum list length (if specified)
    Otherwise, returns the list.
    """
    string_list = [ valid_string(
                        argname=argname, s=x, 
                        min_string_length=min_string_length, 
                        max_string_length=max_string_length, 
                        valid_characters=valid_characters, 
                        invalid_characters=invalid_characters) 
                    for x in s.split(",") ]
    if remove_duplicates:
        string_list = list(set(string_list))
    if min_list_length and len(string_list) < min_list_length:
        raise argparse.ArgumentTypeError(
            "%s list length (%d) is less than minimum (%d). Invalid: %s" % (
                argname, len(string_list), min_list_length, s))
    elif max_list_length and len(string_list) > max_list_length:
        raise argparse.ArgumentTypeError(
            "%s list length (%d) is greater than maximum (%d). Invalid: %s" % (
                argname, len(string_list), max_list_length, s))
    if sort_list:
        string_list.sort()
    return string_list

def valid_nid_list(s, min_nid_count=1, max_nid_count=None):
    """
    Checks s to make sure it specifies a number of positive nids that meets 
    the specified restrictions (if any).
    If so, returns the specified nids (as an integer list), otherwise raises
    an exception.
    """
    nid_set = set()
    for nstring in valid_string_list(argname="nid", s=s, delimiter=",", 
                                     sort_list=False,
                                     min_string_length=1, min_list_length=1, 
                                     valid_characters=string.digits+"-"):
        try:
            int_nid = int(nstring)
            if int_nid < 1:
                raise argparse.ArgumentTypeError("All nids must be positive integers. Invalid nid (not positive): %s" % nstring)
            nid_set.add(int_nid)
            continue
        except ValueError:
            pass
        nid_range = nstring.split("-")
        if len(nid_range) != 2:
            raise argparse.ArgumentTypeError("This is not a nid or a nid range. Invalid: %s" % nstring)
        try:
            nstart = int(nid_range[0])
            nend = int(nid_range[1])
        except ValueError:
            raise argparse.ArgumentTypeError("A valid nid range consists of two integers. Invalid range: %s" % nstring)
        if not 1 <= nstart <= nend:
            raise argparse.ArgumentTypeError("A valid nid range consists of two positive integers, a <= b. Invalid range: %s" % nstring)
        nid_set.update(range(nstart, nend+1))
    if min_nid_count and len(nid_set) < min_nid_count:
        raise argparse.ArgumentTypeError("If specifying nids, at least %d must be specified. Invalid argument: %s" % (min_nid_count, s))
    elif max_nid_count and len(nid_set) > max_nid_count:
        raise argparse.ArgumentTypeError("If specifying nids, at most %d must be specified. Invalid argument: %s" % (max_nid_count, s))
    return sorted(list(nid_set))

# As far as I know, xnames only ever contain letters and numbers
VALID_XNAME_CHARACTERS = string.digits + string.ascii_letters

def valid_xname_list(s, min_xname_count=1, max_xname_count=None):
    """
    At this point, other than verifying that the number of xnames meets the
    specified restrictions (if any), we only verify that they are nonblank
    and contain exclusively alphanumeric characters. No further validation
    is done to make sure they look like legitimate xnames.
    If no problems are found, returns the list, otherwise raises an exception.
    """
    return valid_string_list(argname="xname", s=s, delimiter=",", 
                             min_string_length=1, 
                             min_list_length=min_xname_count, 
                             max_list_length=max_xname_count, 
                             valid_characters=VALID_XNAME_CHARACTERS)

def valid_hsm_group_list(s, min_group_count=1, max_group_count=None):
    """
    At this point, other than verifying that the number of xnames meets the
    specified restrictions (if any), we only verify that they are nonblank
    and contain exclusively valid characters.
    If no problems are found, returns the list, otherwise raises an exception.
    """
    return valid_string_list(argname="HSM group name", s=s, delimiter=",", 
                             min_string_length=1, 
                             min_list_length=min_group_count, 
                             max_list_length=max_group_count, 
                             valid_characters=VALID_HSM_GROUP_CHARACTERS)

def valid_session_template_name(s):
    """
    Raise an exception if s is not blank. Otherwise returns s.
    """
    return nonblank("Session template", s)

def valid_api_cli(s):
    """
    If the specified string is not api or cli (case insensitively), raise an exception. 
    Otherwise return the specified string in lowercase
    """
    if s.lower() not in { 'api', 'cli' }:
        raise argparse.ArgumentTypeError('Must specify "api" or "cli"')
    return s.lower()