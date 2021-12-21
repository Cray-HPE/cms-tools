#!/usr/bin/env python3
# Copyright 2021 Hewlett Packard Enterprise Development LP
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


# This is a fairly straightforward process for testing the barebones boot image.
# The steps are run from main and if the boot succeeds will return '0', any other
# return code indicates failure of the boot test. Any problem encountered will be
# logged with as much information as possible.


import json
import sys
import os
import logging
import requests
import base64
import argparse
import threading
import time

from kubernetes import client, config
from kubernetes.stream import stream

# url for additional troubleshooting help
HELP_URL = "https://github.com/Cray-HPE/docs-csm/blob/main/troubleshooting/cms_barebones_image_boot.md"

# Set up the logger.  This is set up to log minimal information to the 
# console, but a full description to the file.
DEFAULT_LOG_LEVEL = os.environ.get("LOG_LEVEL", logging.INFO)
logger = logging.getLogger("cray.barebones-boot-test")
logger.setLevel(logging.DEBUG)

# set up logging to file
logFilePath = '/tmp/cray.barebones-boot-test.log'
file_handler = logging.FileHandler(filename=logFilePath, mode = 'w')
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

# security token for accessing the api gateway
API_GW_TOKEN = ""

# set up gateway address
PROTOCOL = "http"
API_GW_DNSNAME = "api-gw-service-nmn.local"
API_GW = "%s://%s/apis/" % (PROTOCOL, API_GW_DNSNAME)
API_GW_SECURE = "%ss://%s/apis/" % (PROTOCOL, API_GW_DNSNAME)

# names
BOS_SESSION_TEMPLATE_NAME = "csm-barebones-image-test"

# optional input arguments
INPUT_COMPUTE_NODE = None

# global vars for console processing
# NOTE: these are used by multiple threads
CONSOLE_HAS_INPUT = False
CONSOLE_ERROR_MSG = ""
CONSOLE_FOUND_DRACUT = False
CONSOLE_TERMINATE = False

class BBException(Exception):
    """
    This is the base exception for all custom exceptions that can be raised from
    this application.
    """

def run(k8sClientApi):
    """
    Main workflow of the barebones image boot test process.
    """
    # get the name of the barebones image from ims
    bootImageName, imsEtag, imsPath = get_boot_image_name()
    logger.debug(f"Found boot image: {bootImageName}")

    # create bos session template for this image
    bos_st = create_session_template(imsEtag, imsPath)
    logger.debug(f"Created bos session template: {bos_st}")

    # find a compute node that is available
    computeNode = find_compute_node()
    logger.debug(f"Found compute node for boot: {computeNode}")

    # connect to the nodes console to watch the boot
    # NOTE: make this a daemon thread so it will die when the main process dies
    #  if things go horribly wrong
    t = threading.Thread(target = watch_console, args = (k8sClientApi, computeNode))
    t.daemon = True
    t.start()

    # reboot the node using the created bos template
    bosJobId, bosSessionUrl, bosStatusUrl = start_bos_session(bos_st, computeNode)
    logger.info(f"Starting boot on compute node: {computeNode}")
    logger.debug(f"BOS session job: {bosJobId}, bos session status: {bosStatusUrl}")

    # wait while the node reboots - add a timeout here rather than just joining the
    # thread to handle the case where nothing happens on the console
    consoleTimeoutSec = 900 # 15 min
    global CONSOLE_HAS_INPUT
    global CONSOLE_ERROR_MSG
    global CONSOLE_FOUND_DRACUT
    global CONSOLE_TERMINATE

    # wait for the console output to finish or timeout
    t.join(timeout = consoleTimeoutSec)

    # figure out if it timed out or completed
    if t.is_alive():
        # send terminate signal and wait for it to finish
        logger.warning("Timed out waiting for console response - sending terminate signal")
        CONSOLE_TERMINATE = True
        CONSOLE_ERROR_MSG = "Timed out waiting for reboot"

        # give it 30 seconds to exit cleanly then check
        t.join(30)
        if t.is_alive():
            # Uh-oh - thread wouldn't terminate
            logger.error("Console thread not terminating...")

    # figure out what happened
    if CONSOLE_FOUND_DRACUT:
        # everyone happy - just exit with success
        logger.info(f"Found dracut message in console output - success!!!")
        return

    # if an error was found in the console output report it here
    if len(CONSOLE_ERROR_MSG)!=0:
        logger.error(f"Error encountered during reboot of compute node: {CONSOLE_ERROR_MSG}")
        raise BBException()

    # if we don't find the dracut information from the console then fail
    # NOTE: there may be a case where the watching thread exits without
    #  setting any information at all and we don't want that to pass
    logger.error(f"No information returned from console log")
    raise BBException()

def get_console_operator_pod(k8sClientApi):
    """
    Find the complete name of the running console operator pod.
    """
    # use the k8s api to get the name of the console-operator pod
    #ret = k8sClientApi.list_pod_for_all_namespaces(watch=False)
    ret = k8sClientApi.list_namespaced_pod(namespace = 'services', watch=False)
    for i in ret.items:
        if "console-operator" in i.metadata.name.lower():
            return i.metadata.name
    
    # if nothing was found, then throw an error
    logger.error("Unable to find the cray-console-operator pod")
    raise BBException()

def watch_console(k8sClientApi, computeNode):
    """
    Stream the console log file for the compute node being rebooted for this test
    and parse the output to determine success or failure.  This is meant to be run
    in a separate thread so something else can determine if nothing is going on
    in a reasonable amount of time.
    """
    global CONSOLE_ERROR_MSG
    logger.debug(f"Watching console for node: {computeNode}")

    # grab output stream from desired console - use k8s interface to tail
    # the correct console log file
    opPod = get_console_operator_pod(k8sClientApi)
    logger.debug(f"Found console operator pod: {opPod}")

    # create the stream to feed commands in the console-operator pod
    tailCmd = ['/bin/sh', '-c', f"tail -f -n 0 /var/log/conman/console.{computeNode}"]
    resp = stream(k8sClientApi.connect_get_namespaced_pod_exec, opPod, 'services',
                    container = "cray-console-operator", command = tailCmd, stderr=True,
                    stdin=True, stdout=True, tty=False, _preload_content=False)

    # Read the stream while checking if we should terminate
    while resp.is_open():
        try:
            resp.update(timeout=1)
            # check if we need to terminate the stream processing
            if CONSOLE_TERMINATE:
                logger.warning(f"Responding to console thread terminate command")
                break

            # Look through the console output for something saying we are done
            if resp.peek_stdout() and process_console_output(resp.read_stdout()):
                break
            if resp.peek_stderr() and process_console_output(resp.read_stderr()):
                break

        except Exception as inst:
            # don't know what went wrong, but exit the thread cleanly with an error
            CONSOLE_ERROR_MSG = f"Unexpected exception watching console log: {type(inst)} : {inst}"
            break

    # close down cleanly
    resp.close()
    logger.debug(f"Console watching thread exiting")

def process_console_output(buff):
    """
    Look through a batch of output from the console log file to find if the boot has
    succeeded or if it has encountered a problem.  If this returns 'True' it will stop
    monitoring the console output, a 'False' return will continue to look.
    """
    # Watch for:
    # "Start PXE over IPv4" - starting network boot
    # PXE-E99 - signifies network error
    # PXE-E18: Server response timeout
    # PXE-E16: No offer received
    # "Warning: dracut: FATAL: Don't know how to handle" - successful test

    # all our global vars
    global CONSOLE_HAS_INPUT
    global CONSOLE_ERROR_MSG
    global CONSOLE_FOUND_DRACUT
    
    # Record that something has been seen in console output
    CONSOLE_HAS_INPUT = True

    # list of strings to look for:
    errorStr = {
        "PXE-E99",
        "PXE-E18: Server response timeout",
        "PXE-E16: No offer received"
    }
    successStr = {
        "dracut: FATAL:"
    }

    # split the output buffer into lines
    lines = buff.split("\n")

    # Look through the single line of input for something meaningful
    retVal = False
    for line in lines:
        logger.debug(f"Console log: {line}")
        for errLine in errorStr:
            if errLine in line:
                logger.debug(f"Found console error: {line}")
                CONSOLE_ERROR_MSG = line
                retVal = True
        for successLine in successStr:
            if successLine in line:
                logger.debug(f"Found console success: {line}")
                CONSOLE_FOUND_DRACUT = True
                retVal = True
    
    # Return if we found something signifying we are done looking
    return retVal


def get_boot_image_name():
    """
    Look through the currently defined boot images to find the one we will use
    for the barebones image boot test.  This replicates the command:
    # cray ims images list --format json | jq '.[] | select(.name | contains("barebones"))'
    """
    url = API_GW_SECURE + "ims/images"
    headers = {"Authorization": f"Bearer {API_GW_TOKEN}"}
    params = {"Role":"Compute", "Enabled":"True"}
    r = requests.get(url, headers = headers, params = params)
    if r.status_code != 200:
        logger.error(f"IMS image query incorrect return code: {r.status_code}: {r.text}")
        raise BBException()

    # Look through the result for an image with 'barebones' in the name
    for image in r.json():
        if "barebones" in image['name']:
            logger.debug(f"Found image: {image}")
            return image['name'], image['link']['etag'],image['link']['path']

    # if it gets here then we did not find an appropriate image
    logger.error("Did not find barebones image for boot test")
    raise BBException()

def create_session_template(imsEtag, imsPath):
    """
    Create a new bos session template to do the barebones image boot using the
    image found previously.
    """
    logger.debug(f"Creating bos session template with etag:{imsEtag}, path:{imsPath}")

    # put together the session template information
    compute_set = {
        "boot_ordinal": 2,
        "etag": imsEtag,
        "kernel_parameters": "console=ttyS0,115200 bad_page=panic crashkernel=340M hugepagelist=2m-2g intel_iommu=off intel_pstate=disable iommu=pt ip=dhcp numa_interleave_omit=headless numa_zonelist_order=node oops=panic pageblock_order=14 pcie_ports=native printk.synchronous=y rd.neednet=1 rd.retry=10 rd.shell turbo_boost_limit=999 spire_join_token=${SPIRE_JOIN_TOKEN}",
        "network": "nmn",
        "node_roles_groups": ["Compute"],
        "path": imsPath,
        "rootfs_provider": "cpss3",
        "rootfs_provider_passthrough": "dvs:api-gw-service-nmn.local:300:nmn0",
        "type": "s3"
    }
    bos_params = {
        "name": BOS_SESSION_TEMPLATE_NAME,
        "enable_cfs": False,
        "cfs": {"configuration":"cos-integ-config-1.4.0"},
        "boot_sets": {"compute":compute_set}
        }

    # make the call to bos to create the session template
    url = API_GW_SECURE + "bos/v1/sessiontemplate"
    headers = {"Authorization": f"Bearer {API_GW_TOKEN}"}
    r = requests.post(url, headers = headers, json = bos_params)
    if r.status_code != 201:
        logger.error(f"BOS session template creation incorrect return code: {r.status_code}: {r.text}")
        raise BBException()

    # return the name of the session template
    return BOS_SESSION_TEMPLATE_NAME

def find_compute_node():
    """
    Find a compute node to use for the boot test.  If the user has specified a particular
    compute node, look for that first.  If the user has not specified a node or the one
    the user specified is not available just use the first one returned by hsm.
    """

    # log if the user has specified a compute node
    if not INPUT_COMPUTE_NODE == None:
        logger.debug(f"User specified compute node: {INPUT_COMPUTE_NODE}")

    # query for compute nodes that are enabled
    # cray hsm state components list --role Compute --enabled true
    url = API_GW_SECURE + "smd/hsm/v1/State/Components"
    headers = {"Authorization": f"Bearer {API_GW_TOKEN}"}
    params = {"Role":"Compute", "Enabled":"True"}
    r = requests.get(url, headers = headers, params = params)
    if r.status_code != 200:
        logger.error(f"HSM state components query incorrect return code: {r.status_code}: {r.text}")
        raise BBException()

    # make sure that is json return data from the query
    if r.json() == None:
        logger.error(f"No data returned from state manager.")
        raise BBException()

    # sort through to find a compute node to use - if the user has input a preferred
    # node try to find that one
    firstNode = ""
    matchNode = ""
    for node in r.json()['Components']:
        # grab first one as default
        if firstNode is "":
            firstNode = node['ID']
        
        # If nothing specified by user, we are done or keep looking until we find it
        if INPUT_COMPUTE_NODE is None:
            break
        elif INPUT_COMPUTE_NODE == node['ID']:
            matchNode = node['ID']
            break

    # bail with a failing error if there are no enabled compute nodes present
    if firstNode == "":
        logger.error("No enabled compute nodes present for barebones image boot test")
        raise BBException()

    # If we found one matching what the user wants, use that one
    if matchNode != "":
        return matchNode

    # report if user specified node was not found
    if not INPUT_COMPUTE_NODE is None:
        logger.warning(f"User specified node {INPUT_COMPUTE_NODE} not found, defaulting to node {firstNode}")
    return firstNode

def start_bos_session(template_name, compute_node):
    """
    Start the bos session that attempts to reboot a compute node.
    """
    logger.info(f"Creating bos session with template:{template_name}, on node:{compute_node}")

    # put together the session information
    # NOTE: templateUuid has been deprecated, but templateName (replacement) doesn't work at
    #       this time.  This will need to be fixed at some point in the future.
    bos_params = {
        "templateUuid": template_name,
        "limit": compute_node,
        "operation": "reboot"
        }

    # make the call to bos to create the session template
    url = API_GW_SECURE + "bos/v1/session"
    headers = {"Authorization": f"Bearer {API_GW_TOKEN}"}
    r = requests.post(url, headers = headers, json = bos_params)
    if r.status_code != 201:
        logger.error(f"BOS session creation incorrect return code: {r.status_code}: {r.text}")
        raise BBException()

    # pickle away information about what was created
    sessionUrl = ""
    statusUrl = ""
    rJson = r.json()
    jobId = rJson['job']
    for link in rJson['links']:
        if link['rel'] == "session":
            sessionUrl = link['href']
        elif link['rel'] == "status":
            statusUrl = link['href']

    # return all this information
    return jobId, sessionUrl, statusUrl

def parse_command_line():
    """
    Parse the command line arguments.
    """
    global INPUT_COMPUTE_NODE

    # get the command line arguments
    parser = argparse.ArgumentParser()
    parser.add_argument('-x', "--xname", nargs='?', help='Compute node to use for boot test')
    args = parser.parse_args()

    # get specified compute node if present
    if not args.xname == None:
        INPUT_COMPUTE_NODE = args.xname
        logger.debug(f"Input args.xname={args.xname}")
    else:
        logger.debug("Input arg xname not specified.")

def get_access_token(k8sClientApi):
    """
    Get the admin secret from k8s for the api gateway - command line equivalent is:
    #`kubectl get secrets admin-client-auth -o jsonpath='{.data.client-secret}' | base64 -d`
    """
    try:
        sec = k8sClientApi.read_namespaced_secret("admin-client-auth", "default").data
        adminSecret = base64.b64decode(sec['client-secret'])
    except Exception as err:
        logger.error(f"An unanticipated exception occurred while retrieving k8s secrets {err}")
        raise BBException from None

    # get an access token from keycloak
    payload = {"grant_type":"client_credentials",
               "client_id":"admin-client",
               "client_secret":adminSecret}
    url = "https://api-gw-service-nmn.local/keycloak/realms/shasta/protocol/openid-connect/token"
    r = requests.post(url, data = payload)

    # if the token was not provided, log the problem
    if r.status_code != 200:
        logger.error(f"Error retrieving gateway token: keycloak return code: {r.status_code}"
            f" text: {r.text}")
        raise BBException from None

    # pull the access token from the return data
    global API_GW_TOKEN
    API_GW_TOKEN = r.json()['access_token']

if __name__ == "__main__":
    # Format logs for stdout
    logger.info("Barebones image boot test starting")
    logger.info(f"  For complete logs look in the file {logFilePath}")

    # process any command line inputs
    parse_command_line()

    # initialize k8s
    config.load_kube_config()
    k8sClientApi = client.CoreV1Api()

    # secure an access token for the api gateway
    get_access_token(k8sClientApi)

    try:
        run(k8sClientApi)
    except BBException:
        logger.error("Failure of barebones image boot test.")
        logger.info(f"For troubleshooting information and manual steps, see {HELP_URL}")
        sys.exit(1)
    except Exception as err:
        logger.exception("An unanticipated exception occurred during during barebones image "
                         "boot test : %s; ", err)
        logger.info(f"For troubleshooting information and manual steps, see {HELP_URL}")
        sys.exit(1)

    # exit indicating success
    logger.info("Successfully completed barebones image boot test.")
    sys.exit(0)
