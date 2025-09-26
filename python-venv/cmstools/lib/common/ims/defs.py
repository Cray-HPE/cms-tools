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
IMS definitions
"""

from cmstools.lib.common.api import API_BASE_URL
from cmstools.lib.common.defs import ARM_ARCH, X86_ARCH

# IMS URLs
IMS_URL = f"{API_BASE_URL}/ims"
IMS_IMAGES_URL = f"{IMS_URL}/images"

# We use IMS v2 for deleting images because it gives the option to
# also delete the S3 resources associated with them.
IMS_V2_URL = f"{IMS_URL}/v2"
IMS_V2_IMAGES_URL = f"{IMS_V2_URL}/images"

# The strings IMS uses to identify image arch
IMS_ARM_ARCH = "aarch64"
IMS_X86_ARCH = "x86_64"
IMS_ARCH_STRINGS = { ARM_ARCH: IMS_ARM_ARCH, X86_ARCH: IMS_X86_ARCH }
