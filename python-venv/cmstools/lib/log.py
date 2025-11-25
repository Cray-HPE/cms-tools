#
# MIT License
#
# (C) Copyright 2021-2022, 2024-2025 Hewlett Packard Enterprise Development LP
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
Logging setup
"""

import logging
import os
import sys

from cmstools.lib.defs import TEST_TIMESTAMP

# The following line is also used by the Makefile and RPM spec file in this repo. Any changes to it
# should ensure that they do not break this.
DEFAULT_LOG_DIR = "/opt/cray/tests/integration/logs/csm/cmstools"

# Set up the logger.  This is set up to log minimal information to the
# console, but a full description to the file.
DEFAULT_LOG_LEVEL = os.environ.get("LOG_LEVEL", logging.INFO)
LOG_FILE_PATH = ""


def set_log_file_path(path: str) -> None:
    global LOG_FILE_PATH
    LOG_FILE_PATH = path

def get_test_logger(test_name: str) -> logging.Logger:
    log_dir = os.path.join(DEFAULT_LOG_DIR, test_name)
    os.makedirs(log_dir, exist_ok=True)
    log_file_path = f"{log_dir}/{TEST_TIMESTAMP}.log"
    set_log_file_path(log_file_path)
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)

    if not logger.handlers:
        # File handler
        file_handler = logging.FileHandler(filename=log_file_path, mode='w')
        file_handler.setLevel(os.environ.get("FILE_LOG_LEVEL", logging.DEBUG))
        file_handler.setFormatter(logging.Formatter('%(asctime)-15s - %(process)d - %(thread)d - %(levelname)-8s %(message)s'))
        logger.addHandler(file_handler)

        # Console handler
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(os.environ.get("CONSOLE_LOG_LEVEL", DEFAULT_LOG_LEVEL))
        console_handler.setFormatter(logging.Formatter('%(levelname)-8s %(message)s'))
        logger.addHandler(console_handler)

    return logger
