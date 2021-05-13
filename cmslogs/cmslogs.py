#!/usr/bin/env python3
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

import argparse
import os
import pathlib
import subprocess
import sys
import tempfile

outfile_dir = "/tmp"
outfile_base = "cmslogs.tar"
outfile = "%s/%s" % (outfile_dir, outfile_base)
cray_release = "/etc/cray-release"
motd = "/etc/motd"
opt_cray_tests = "/opt/cray/tests"
tmp_cray_tests = "/tmp/cray/tests"
cmsdev = "/usr/local/bin/cmsdev"
cksum_cmd_string = "cksum %s" % cmsdev
rpm_cmd_string = "rpm -qa"

class cmslogsError(Exception):
    pass

def get_tmpfile(tempfile_list, **kwargs):
    tnum, tpath = tempfile.mkstemp(dir="/tmp", **kwargs)
    tempfile_list.append(tpath)
    return tpath

def error_exit(msg):
    sys.exit("ERROR: %s" % msg)

def create_collect_list(args, mylogfilepath, tempfile_list):
    with open(mylogfilepath, "wt") as mylog:
        collect_list = []
        def printmylog(s):
            print(s, file=mylog, flush=True)
        def does_not_exist(thing):
            printmylog("%s does not exist" % thing)
        def skipping(thing):
            printmylog("Skipping %s, as specified" % thing)
        def collecting(thing):
            printmylog("Will collect %s" % thing)
            # [1:] to strip the leading / from the absolute path
            collect_list.append(thing[1:])
        def add_to_collection_list(thing, skip_arg):
            if os.path.exists(thing):
                if skip_arg:
                    skipping(thing)
                else:
                    collecting(thing)
            else:
                does_not_exist(thing)
        def record_cmd_to_file(cmdstring, **tmpfile_args):
            cmdoutfilepath = get_tmpfile(tempfile_list=tempfile_list, **tmpfile_args)
            printmylog("# %s > %s" % (cmdstring, cmdoutfilepath))
            with open(cmdoutfilepath, "wt") as cmdoutfile:
                print("# %s" % cmdstring, file=cmdoutfile, flush=True)
                try:
                    subprocess.run(cmdstring.split(), stdout=cmdoutfile, stderr=subprocess.PIPE, check=True, universal_newlines=True)
                except subprocess.CalledProcessError as e:
                    raise cmslogsError("%s command failed with return code %d: %s" % (cmdstring, e.returncode, e.stderr))
            collecting(cmdoutfilepath)

        cmdline=' '.join(sys.argv)
        printmylog(cmdline)
        printmylog("#"*len(cmdline))
        add_to_collection_list(cray_release, args.no_cray_release)
        add_to_collection_list(motd, args.no_motd)
        add_to_collection_list(tmp_cray_tests, args.no_tmp_cray_tests)
        if not args.no_opt_cray_tests_all:
            if args.no_opt_cray_tests:
                # Need to find log files in /opt/cray/tests
                for path in pathlib.Path(opt_cray_tests).rglob('*.log'):
                    if path.is_file():
                        collecting(str(path))
            else:
                # Collecting all of /opt/cray/tests
                add_to_collection_list(opt_cray_tests, False)
        else:
            add_to_collection_list(opt_cray_tests, True)
        if os.path.exists(cmsdev):
            if os.path.isfile(cmsdev):
                if args.no_cmsdev_sum:
                    printmylog("Not recording cksum of %s, as specified" % cmsdev)
                else:
                    record_cmd_to_file(cmdstring=cksum_cmd_string, prefix="cmslogs-cmsdev-cksum-", suffix=".txt")
            else:
                raise cmslogsError("%s exists but is not a regular file" % cmsdev)
        else:
            does_not_exist(cmsdev)
        if not args.no_rpms:
            record_cmd_to_file(cmdstring=rpm_cmd_string, prefix="cmslogs-rpmlist-", suffix=".txt")
        else:
            printmylog("Not recording output of %s command, as specified" % rpm_cmd_string)
    return collect_list

def do_collect(collect_list):
    tar_cmd_list = [ "tar", "-C", "/", "-cf", outfile ]
    tar_cmd_list.extend(collect_list)
    try:
        print("Running command: %s" % ' '.join(tar_cmd_list), flush=True)
        cmdproc = subprocess.run(tar_cmd_list, check=True, universal_newlines=True)
    except subprocess.CalledProcessError as e:
        raise cmslogsError("tar command failed with return code %d" % (cmsdev, e.returncode))

def remove_tempfiles(tempfile_list):
    for t in tempfile_list:
        os.remove(t)

if __name__ == '__main__':
    tempfile_list = list()
    parser = argparse.ArgumentParser(
        description="Collect files for test debug and stores them in %s" % outfile)
    parser.add_argument("-f", dest="overwrite", 
        action="store_true", 
        help="Overwrite outfile (%s) if it exists" % outfile)
    parser.add_argument("--no-cmsdev-sum", dest="no_cmsdev_sum", 
        action="store_true", 
        help="Do not record output of %s command" % cksum_cmd_string)
    parser.add_argument("--no-cray-release", dest="no_cray_release", 
        action="store_true", 
        help="Do not collect %s" % cray_release)
    parser.add_argument("--no-motd", dest="no_motd", action="store_true",
        help="Do not collect %s" % motd)
    parser.add_argument("--no-opt-cray-tests", dest="no_opt_cray_tests",
        action="store_true", 
        help="Do not collect %s directory (except logs)" % opt_cray_tests)
    parser.add_argument("--no-opt-cray-tests-all", 
        dest="no_opt_cray_tests_all", action="store_true",
        help="Do not collect %s directory (including logs)" % opt_cray_tests)
    parser.add_argument("--no-rpms", dest="no_rpms",
        action="store_true", 
        help="Do not collect output of %s command" % rpm_cmd_string)
    parser.add_argument("--no-tmp-cray-tests", dest="no_tmp_cray_tests",
        action="store_true", 
        help="Do not collect %s directory" % tmp_cray_tests)
    args = parser.parse_args()
    
    if os.path.exists(outfile):
        if not args.overwrite:
            error_exit("Output file already exists: %s" % outfile)
        elif not os.path.isfile(outfile):
            error_exit("Output file already exists and is not a regular file: %s" % outfile)

    mylogfilepath = get_tmpfile(tempfile_list=tempfile_list, prefix="cmslogs-", suffix=".log")
    try:
        collect_list = create_collect_list(args, mylogfilepath, tempfile_list)
        if not collect_list:
            raise cmslogsError("Nothing to collect!")
        # [1:] to strip the leading / from the absolute path
        collect_list.append(mylogfilepath[1:])
        do_collect(collect_list)
    except cmslogsError as e:
        remove_tempfiles(tempfile_list)
        error_exit(str(e))
    remove_tempfiles(tempfile_list)
    print("\nRequested data successfully collected: %s" % outfile, flush=True)
    sys.exit(0)
