#!/usr/bin/env python3
#
# MIT License
#
# (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP
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
This is a fairly straightforward process for testing the barebones boot image.
The steps are run from main and if the boot succeeds will return '0', any other
return code indicates failure of the boot test. Any problem encountered will be
logged with as much information as possible.

The default test procedure is:
1. For the latest CSM version in the Cray Product Catalog, identify the following:
  a. The IMS ID of the X86 barebones compute image
  b. The VCS clone URL and commit ID
2. Pick an enabled X86 compute node.
3. Create a CFS configuration using the VCS data.
4. Create a customized IMS image using the CFS configuration and the barebones compute image.
5. Create a BOS session template for that image.
6. Create a BOS session to boot the chosen compute node using that template.
7. Wait for the session to complete.
8. If the test is successful, delete the resources created during its execution (CFS
   configuration, CFS session, customized IMS image, BOS session template, and BOS session)

Command line options to the test allow the user to modify the above procedure in various ways.
For example:
- Instead of using the latest CSM version in the product catalog, specify the version to use.
- Bypass the entire image customization process by specifying the IMS ID of a customized image.
- Still do the image customization, but specify a different base image, or different VCS options.
- Specify a specific target compute node to use.
- Specify that the test should find Arm architecture images and nodes and run against those.
- Specify that even on success, the test should not delete the resources that it creates.
"""

import argparse
import sys

from typing import NamedTuple

from cmstools.test.barebones_image_test.bos import BosSession, BosTemplate
from cmstools.test.barebones_image_test.cfs import CfsConfigLayerData, CfsConfig, CfsSession, CfsComponents, CfsComponentUpdateData
from cmstools.test.barebones_image_test.hsm import ComputeNode, find_compute_node, get_compute_node
from cmstools.test.barebones_image_test.prodcat import CsmProductCatalogData
from cmstools.lib.common.defs import ARCH_LIST, TestException as BBException
from cmstools.test.barebones_image_test.ims import ImsImage
from cmstools.lib.common.log import LOG_FILE_PATH, get_test_logger
from cmstools.test.barebones_image_test.test_resource import TestResource

logger = get_test_logger("barebones_image_test")

class HelpUrl:
    """
    URL for additional troubleshooting help
    """
    URL_BASE = "https://github.com/Cray-HPE/docs-csm/blob"
    URL = "troubleshooting/cms_barebones_image_boot.md"

    def __init__(self):
        self.url = HelpUrl.URL

    def update(self, csm_prodcat_data: CsmProductCatalogData) -> None:
        """
        Update HELP_URL now that we know the CSM version
        """
        # Convert latest_csm_version to 'release/a.b' format:
        release_branch = "release/" + csm_prodcat_data.major_minor
        # Update HELP_URL
        self.url = f"{HelpUrl.URL_BASE}/{release_branch}/{HelpUrl.URL}"
        logger.debug("New HELP_URL value: %s", self.url)


class ScriptArgs(NamedTuple):
    """
    Encapsulates the arguments to the script
    """
    cleanup_on_success: bool
    playbook: str
    csm_version: str | None
    base_ims_image: ImsImage | None
    customized_ims_image: ImsImage | None
    target_node: str | None
    cfs_config: CfsConfig | None
    vcs_url: str | None
    git_commit: str | None
    arch: str

help_url = HelpUrl()
created_resources: list[TestResource] = []

DEFAULT_ARCH = ARCH_LIST[0]
DEFAULT_PLAYBOOK = "compute_nodes.yml"

def record_resource_creation(resource: TestResource) -> None:
    """
    Record that this resource has been created
    """
    created_resources.append(resource)
    logger.info("Created %s", resource.label_and_name)


def cleanup_resources(xname: str=None) -> None:
    """
    Delete each created resource
    """
    if not created_resources:
        return
    logger.info("Cleaning up resources created during the test execution")
    while created_resources:
        if isinstance(created_resources[-1], CfsConfig):
            # CFS config should be unset from component before deletion
            CfsComponents.update_cfs_component(cfs_component_name=xname, data=CfsComponentUpdateData(desired_config=""))
        created_resources.pop().delete()

def get_cfs_config(script_args:ScriptArgs, csm_prodcat_data: CsmProductCatalogData=None) -> CfsConfig:
    """
    Get the CFS configuration to use for the test
    """
    if script_args.cfs_config is not None:
        logger.debug("Using user-specified %s", script_args.cfs_config.label_and_name)
        return script_args.cfs_config

    for resource in created_resources:
        if isinstance(resource, CfsConfig):
            logger.debug("Using created %s", resource.label_and_name)
            return resource

    # This means we are creating a new CFS config
    cfs_config_layer_data = get_cfs_config_layer_data(script_args, csm_prodcat_data)
    cfs_config = CfsConfig.create_in_cfs(cfs_config_layer_data)
    record_resource_creation(cfs_config)
    return cfs_config


def run(script_args: ScriptArgs) -> None:
    """
    Main workflow of the barebones image boot test process.
    """
    # Look up latest CSM product entry in Cray Product Catalog
    if script_args.csm_version is None:
        csm_prodcat_data = CsmProductCatalogData.get_latest()
    else:
        csm_prodcat_data = CsmProductCatalogData.get_version(script_args.csm_version)

    help_url.update(csm_prodcat_data)

    compute_node = script_args.target_node
    if compute_node is None:
        compute_node = find_compute_node(script_args.arch)

    ims_image = script_args.customized_ims_image
    if ims_image is None:
        ims_image = get_customized_image(csm_prodcat_data, script_args)
    logger.debug("Using customized %s", ims_image.label_and_name)

    # create BOS session template for this image
    cfs_config = get_cfs_config(script_args=script_args)
    bos_st = BosTemplate(ims_image=ims_image, cfs_config_name=cfs_config.name)
    record_resource_creation(bos_st)

    # Create a BOS session to reboot the node using the created BOS template,
    # and wait for it to complete
    bos_session = BosSession(bos_st, compute_node)
    record_resource_creation(bos_session)
    bos_session.wait_for_session_to_complete()
    logger.info("BOS session completed with no errors - success!!!")
    if script_args.cleanup_on_success:
        cleanup_resources(xname=compute_node.xname)


def get_customized_image(csm_prodcat_data: CsmProductCatalogData,
                         script_args: ScriptArgs) -> ImsImage:
    """
    Customize an IMS image and return it
    """
    # If the base image has been specified, then this is the simplest case to cover.
    if script_args.base_ims_image is not None:
        return customize_base_image(script_args.base_ims_image, csm_prodcat_data, script_args)

    # No base image has been specified. We must turn to the product catalog to
    # get the base image ID for the right architecture

    logger.debug("Getting IMS ID of %s barebones image from Product Catalog", script_args.arch)
    base_ims_image_id = csm_prodcat_data.barebones_image_id(script_args.arch)
    base_ims_image = ImsImage(base_ims_image_id)
    # And let's not take the product catalog's word about the image architecture
    base_ims_image.load_from_ims()
    logger.info("Base %s from the product catalog has arch '%s'", base_ims_image.label_and_name,
                base_ims_image.arch)
    if base_ims_image.arch != script_args.arch:
        logger.error("%s found in the product catalog should have arch '%s' but actually "
                     "has arch '%s'", script_args.arch, base_ims_image.label_and_name,
                     base_ims_image.arch)
        raise BBException()
    return customize_base_image(base_ims_image, csm_prodcat_data, script_args)


def customize_base_image(base_ims_image: ImsImage, csm_prodcat_data: CsmProductCatalogData,
                         script_args: ScriptArgs) -> ImsImage:
    """
    Create a customized image from the specified base image
    """
    cfs_config = get_cfs_config(script_args=script_args, csm_prodcat_data=csm_prodcat_data)
    cfs_session = CfsSession(base_ims_image=base_ims_image, cfs_config=cfs_config)
    record_resource_creation(cfs_session)
    cfs_session.wait_for_session_to_complete()
    ims_image_id = cfs_session.result_id
    ims_image = ImsImage(ims_image_id)
    record_resource_creation(ims_image)
    logger.info("%s created customized %s", cfs_session.label_and_name, ims_image.label_and_name)
    ims_image.load_from_ims()
    return ims_image


def get_cfs_config_layer_data(script_args: ScriptArgs,
                              csm_prodcat_data: CsmProductCatalogData) -> CfsConfigLayerData:
    """
    Parse the script arguments and CSM product catalog entry to collect the CFS
    configuration layer data
    """
    if script_args.vcs_url is not None:
        vcs_url = script_args.vcs_url
        logger.debug("Using user-specified VCS clone URL: '%s'", vcs_url)
    else:
        # Read it from the Product Catalog
        vcs_url = csm_prodcat_data.clone_url
        logger.debug("Using VCS clone URL from product catalog: '%s'", vcs_url)

    if script_args.git_commit is not None:
        git_commit = script_args.git_commit
        logger.debug("Using user-specified git commit: '%s'", git_commit)
    else:
        # Read it from the Product Catalog
        git_commit = csm_prodcat_data.commit
        logger.debug("Using git commit from product catalog: '%s'", git_commit)

    return CfsConfigLayerData(playbook=script_args.playbook, vcs_url=vcs_url,
                              git_commit=git_commit)


def check_for_mutually_exclusive_arguments(args: argparse.Namespace,
                                           arg_mutex_map: dict[str, list[str]]) -> None:
    """
    Log an error and raise an exception if any mutually exclusive arguments were specified
    """

    def arg_specified(arg_name: str, args: argparse.Namespace) -> bool:
        """
        Returns True if arg_name was specified. False otherwise.
        """
        # Strip the leading -- and then convert - to _ to go from the name of the arguments
        # to the name of the attr in the argparse.Namespace
        return getattr(args, arg_name.lstrip('-').replace('-', '_')) is not None

    conflicting_args = False
    for arg_name, mutex_args in arg_mutex_map.items():
        if not arg_specified(arg_name, args):
            # Nothing to check if this argument was not specified
            continue

        for mutex_arg in mutex_args:
            if arg_specified(mutex_arg, args):
                conflicting_args = True
                logger.error("Arguments '%s' and '%s' are mutually exclusive", arg_name, mutex_arg)

    if conflicting_args:
        raise BBException()


def load_user_specified_data(arch: str|None,
                             base_image_id: str|None,
                             cfs_config_name: str|None,
                             customized_image_id: str|None,
                             node_xname: str|None) -> tuple[str, ImsImage|None, CfsConfig|None,
                                                            ImsImage|None, ComputeNode|None]:
    """
    Make sure user-specified IMS images, CFS configs, and node xnames (if any) exist
    on the system and do not conflict with each other. Also determines the arch value.
    """
    base_ims_image, cfs_config, customized_ims_image, target_node = [None]*4

    if cfs_config_name is not None:
        cfs_config = CfsConfig(cfs_config_name)
        if not cfs_config.exists:
            logger.error("Specified %s does not exist", cfs_config.label_and_name)
            raise BBException()
        logger.info("Specified %s", cfs_config.label_and_name)

    if arch is not None:
        # arch is mutually exclusive with the xname and IMS image arguments, so
        # we can stop here.
        return arch, base_ims_image, cfs_config, customized_ims_image, target_node

    if base_image_id is not None:
        base_ims_image=ImsImage(base_image_id)
        base_ims_image.load_from_ims()
        logger.info("Specified base %s has arch '%s'", base_ims_image.label_and_name,
                    base_ims_image.arch)
        arch = base_ims_image.arch

    if customized_image_id is not None:
        customized_ims_image=ImsImage(customized_image_id)
        customized_ims_image.load_from_ims()
        logger.info("Specified customized %s has arch '%s'", customized_ims_image.label_and_name,
                    customized_ims_image.arch)
        arch = customized_ims_image.arch

    if node_xname is not None:
        target_node=get_compute_node(node_xname)
        logger.info("Specified compute node '%s' has arch '%s'", target_node.xname,
                    target_node.arch)

        # Check for any possible architecture conflicts between just the
        # user-specified arguments. Since --arch is mutually exclusive with xname/image arguments,
        # the only thing to check is that if an xname was specified and an image was specified,
        # that they do not conflict.

        if arch is None:
            # This means that no IMS images were specified, thus there can be no conflicts.
            arch = target_node.arch
            return arch, base_ims_image, cfs_config, customized_ims_image, target_node

        if base_ims_image is not None:
            if base_ims_image.arch != target_node.arch:
                logger.error("Conflicting architectures: Specified target node '%s' is %s, but "
                             "specified base %s is %s", target_node.xname, target_node.arch,
                             base_ims_image.label_and_name, base_ims_image.arch)
        elif customized_ims_image is not None:
            if customized_ims_image.arch != target_node.arch:
                logger.error("Conflicting architectures: Specified target node '%s' is %s, but "
                             "specified customized %s is %s", target_node.xname, target_node.arch,
                             customized_ims_image.label_and_name, customized_ims_image.arch)

    if arch is None:
        # This means that no arch, xname, or IMS images were specified. In this case, we use the
        # default arch
        arch = DEFAULT_ARCH
        logger.info("No architecture, node, or image specified; using default arch: %s", arch)

    return arch, base_ims_image, cfs_config, customized_ims_image, target_node

def parse_command_line() -> ScriptArgs:
    """
    Parse the command line arguments and returns them.
    """
    # get the command line arguments
    parser = argparse.ArgumentParser()
    # By default, the test will find the entry for the latest CSM version in the product
    # catalog, find the barebones x86 compute image ID in that entry, and customize that image
    # to use for the test.
    # The '--csm-version' flag can be used to instead specify the CSM version to use in the
    # Product Catalog
    # or
    # The '--id' flag can be used to specify the IMS ID of the already-customized image to use.
    parser.add_argument('--arch', choices=ARCH_LIST, help="Node/image architecture to use")
    parser.add_argument('--csm-version', help='Use barebones image and VCS info for '
                                              'this CSM version in the Product Catalog')

    parser.add_argument('--base-id', help='IMS ID of image to customize and boot')
    parser.add_argument('--id', help='IMS ID of customized image to boot')
    parser.add_argument('--cfs-config',
                        help='CFS configuration to use when customizing the base barebones image')
    parser.add_argument('--vcs-url', help='VCS URL to use when customizing image')
    parser.add_argument('--git-commit', help='git commit to use when customizing image')
    parser.add_argument('--playbook', help='Playbook to use when customizing image')
    parser.add_argument('--xname', help='Xname of compute node to use for boot test')
    parser.add_argument('--no-cleanup', action='store_true',
                        help='Do not delete created resources on test success')
    args = parser.parse_args()

    logger.debug("Input args %s", args)

    # --arch is mutually exclusive with --id, --base-id, and --xname
    #
    # --id is mutually exclusive with --base-id, --cfs-config, --vcs-url, --git-commit, and
    # --playbook
    #
    # --cfs-config is mutually exclusive with --vcs-url, --git-commit, and --playbook
    arg_mutex_map = {
        '--arch': [ '--base-id', '--id', '--xname' ],
        '--id': [ '--base-id', '--cfs-config', '--vcs-url', '--git-commit', '--playbook' ],
        '--cfs-config': [ "--vcs-url", "--git-commit", "--playbook" ] }
    check_for_mutually_exclusive_arguments(args, arg_mutex_map)

    arch, base_ims_image, cfs_config, \
    customized_ims_image, \
    target_node = load_user_specified_data(arch=args.arch, base_image_id=args.base_id,
                                           cfs_config_name=args.cfs_config,
                                           customized_image_id=args.id,
                                           node_xname=args.xname)

    return ScriptArgs(csm_version=args.csm_version,
                      base_ims_image=base_ims_image,
                      cfs_config=cfs_config,
                      vcs_url=args.vcs_url,
                      git_commit=args.git_commit,
                      playbook=args.playbook if args.playbook is not None else DEFAULT_PLAYBOOK,
                      customized_ims_image=customized_ims_image,
                      target_node=target_node,
                      arch=arch,
                      cleanup_on_success=not args.no_cleanup)


def main():
    # Format logs for stdout
    logger.info("Barebones image boot test starting")
    logger.info("  For complete logs look in the file %s", LOG_FILE_PATH)

    try:
        # process any command line inputs
        script_args = parse_command_line()

        run(script_args)
    except BBException:
        logger.error("Failure of barebones image boot test.")
        logger.info("For troubleshooting information and manual steps, see %s", help_url.url)
        sys.exit(1)
    except Exception as err:
        logger.exception("An unanticipated exception occurred during during barebones image "
                         "boot test : %s; ", err)
        logger.info("For troubleshooting information and manual steps, see %s", help_url.url)
        sys.exit(1)

    # exit indicating success
    logger.info("Successfully completed barebones image boot test.")
    sys.exit(0)

if __name__ == "__main__":
    main()
