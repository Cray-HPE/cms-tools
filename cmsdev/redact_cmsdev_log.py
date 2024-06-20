#!/usr/bin/env python3
#
# MIT License
#
# (C) Copyright 2024 Hewlett Packard Enterprise Development LP
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
Redact credentials that were overzealously logged by prior cmsdev versions
"""

import os
import shutil
import tempfile

LOGDIR="/opt/cray/tests/install/logs/cmsdev"
LOGPATH=os.path.join(LOGDIR, "cmsdev.log")

class CommandLine:
    def __init__(self, pattern, delete_line=True, delete_output=True):
        self.pattern = pattern
        self.delete_line = delete_line
        self.delete_output = delete_output

REMOVE_LINES = [
    'vcs username (base 64) = ',
    'Decoded vcs username = ',
    'vcs password (base 64) = ',
    'Decoded vcs password = ' ]

RUN_COMMAND = ' msg="Running command: ' 
COMMAND_OUTPUT = ' msg="Command output: '

COMMAND_LINES = [
    CommandLine(' | base64 -d'),
    CommandLine('kubectl get secret -n services vcs-user-credentials', delete_line=False),
    CommandLine('kubectl get secrets admin-client-auth',  delete_line=False),
    CommandLine('curl -k -s -d grant_type=client_credentials', delete_output=False)
]

def scan_line(logline, remove_command_output):
    """
    Returns a tuple of remove_line, remove_command_output
    """
    remove_line = False
    if remove_command_output:
        remove_command_output = False
        if COMMAND_OUTPUT in logline:
            # Remove this line
            remove_line = True
    remove_line = remove_line or any(pattern in logline for pattern in REMOVE_LINES)
    for cmdline in COMMAND_LINES:
        if cmdline.pattern in logline:
            remove_line = remove_line or cmdline.delete_line
            remove_command_output = remove_command_output or cmdline.delete_output
    return remove_line, remove_command_output

def main():
    if not os.path.isfile(LOGPATH):
        return
    remove_command_output = False
    tmpfilepath = tempfile.mkstemp(dir=LOGDIR)[1]
    with open(LOGPATH, "rt") as logfile:
        with open(tmpfilepath, "wt") as tmpfile:
            for logline in logfile:
                remove_line, remove_command_output = scan_line(logline, remove_command_output)
                if not remove_line:
                    tmpfile.write(logline)
    # Copy over original log file
    shutil.copyfile(tmpfilepath, LOGPATH)
    # Remove temporary file
    os.remove(tmpfilepath)

if __name__ == "__main__":
    main()
