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
BosTemplate class
"""

from typing import ClassVar

from barebones_image_test.api import request_and_check_status
from barebones_image_test.ims import ImsImage
from barebones_image_test.log import logger
from barebones_image_test.test_resource import TestResource

from .defs import BOS_TEMPLATES_URL

class BosTemplate(TestResource):
    """
    BOS session template
    """
    base_url: ClassVar[str] = BOS_TEMPLATES_URL
    label: ClassVar[str] = "BOS session template"

    def __init__(self, ims_image: ImsImage, template_name: str|None=None):
        """
        Create the specified BOS session template
        """
        if template_name is None:
            super().__init__()
        else:
            super().__init__(template_name)
        logger.debug("Creating %s with %s (etag:%s, path:%s)", self.label_and_name,
                     ims_image.label_and_name, ims_image.s3_etag, ims_image.s3_path)
        logger.info("Creating %s", self.label_and_name)

        # put together the session template information
        kernel_parameters = (
            "console=ttyS0,115200 bad_page=panic crashkernel=512M hugepagelist=2m-2g "
            "intel_iommu=off intel_pstate=disable iommu.passthrough=on "
            "modprobe.blacklist=amdgpu numa_interleave_omit=headless oops=panic pageblock_order=14 "
            "rd.neednet=1 rd.retry=10 rd.shell split_lock_detect=off "
            "systemd.unified_cgroup_hierarchy=1 ip=dhcp quiet spire_join_token=${SPIRE_JOIN_TOKEN} "
            f"root=live:s3://boot-images/{ims_image.name}/rootfs "
            f"nmd_data=url=s3://boot-images/{ims_image.name}/rootfs,etag={ims_image.s3_etag}")

        compute_set = {
            "etag": ims_image.s3_etag,
            "kernel_parameters": kernel_parameters,
            "node_roles_groups": ["Compute"],
            "path": ims_image.s3_path,
            "rootfs_provider": "",
            "rootfs_provider_passthrough": "",
            "type": "s3" }

        bos_params = {
            "enable_cfs": False,
            "boot_sets": {"compute": compute_set} }

        # make the call to bos to create the session template
        resp_data = request_and_check_status("put", self.url, expected_status=200, parse_json=True,
                                             json=bos_params)
        logger.debug("Created %s: %s", self.label_and_name, resp_data)
