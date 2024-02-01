#
# MIT License
#
# (C) Copyright 2021-2022, 2024 Hewlett Packard Enterprise Development LP
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

from barebones_image_test.defs import BB_TEST_TIMESTAMP

# The following line is also used by the Makefile and RPM spec file in this repo. Any changes to it
# should ensure that they do not break this.
DEFAULT_LOG_DIR = "/opt/cray/tests/integration/logs/csm/barebones_image_test"

# Set up the logger.  This is set up to log minimal information to the
# console, but a full description to the file.
DEFAULT_LOG_LEVEL = os.environ.get("LOG_LEVEL", logging.INFO)
logger = logging.getLogger("cray.barebones-boot-test")
logger.setLevel(logging.DEBUG)

# set up logging to file
LOG_FILE_PATH = f'{DEFAULT_LOG_DIR}/{BB_TEST_TIMESTAMP}.log'
file_handler = logging.FileHandler(filename=LOG_FILE_PATH, mode = 'w')
file_handler.setLevel(os.environ.get("FILE_LOG_LEVEL", logging.DEBUG))
formatter = logging.Formatter('%(asctime)s: %(levelname)-8s %(message)s')
file_handler.setFormatter(formatter)
logger.addHandler(file_handler)

# set up logging to console
console_handler = logging.StreamHandler(sys.stdout)
console_handler.setLevel(os.environ.get("CONSOLE_LOG_LEVEL", DEFAULT_LOG_LEVEL))
formatter = logging.Formatter('%(name)-12s: %(levelname)-8s %(message)s')
console_handler.setFormatter(formatter)
logger.addHandler(console_handler)
