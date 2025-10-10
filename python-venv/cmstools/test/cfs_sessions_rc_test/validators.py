#
# MIT License
#
# (C) Copyright 2025 Hewlett Packard Enterprise Development LP
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
Validation function for parsed arguments
"""

import re
import argparse

from .defs import (MAX_NAME_LENGTH, MIN_NAME_LENGTH)

# Validations for command line arguments
def check_min_page_size(value) -> int:
    int_value = int(value)
    if int_value < 1:
        raise argparse.ArgumentTypeError("--page-size must be at least 1")
    return int_value

def validate_name(value) -> str:
    pattern = r'^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'
    # Script appends a number to the name, so allow up to 40 characters here
    if not (MIN_NAME_LENGTH <= len(value) <= MAX_NAME_LENGTH):
        raise argparse.ArgumentTypeError("--name must be between 1 and 40 characters")
    if not re.match(pattern, value):
        raise argparse.ArgumentTypeError("--name must match pattern: " + pattern)
    return value

def check_minimum_max_sessions(value) -> int:
    int_value = int(value)
    if int_value < 1:
        raise argparse.ArgumentTypeError("--max-sessions must be at least 1")
    return int_value

def check_minimum_max_parallel_reqs(value) -> int:
    int_value = int(value)
    if int_value < 1:
        raise argparse.ArgumentTypeError("--max-multi-delete-reqs must be at least 1")
    return int_value

def check_subtest_names(value) -> list[str]:
    # Avoiding circular import
    from .__main__ import subtest_functions_dict

    names = [name.strip() for name in value.split(",")]
    invalid_names = [name for name in names if name not in subtest_functions_dict.keys()]
    if invalid_names:
        raise argparse.ArgumentTypeError(f"Invalid subtest names: {', '.join(invalid_names)}, valid names are: {', '.join(subtest_functions_dict.keys())}")
    return names