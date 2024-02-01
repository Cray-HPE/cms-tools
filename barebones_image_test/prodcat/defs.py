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
CsmProductCatalogData definitions
"""

from barebones_image_test.api import API_GW_SECURE
from barebones_image_test.defs import ARM_ARCH, X86_ARCH

# VCS URLs:
VCS_URL = f"{API_GW_SECURE}/vcs"

PRODUCT_CATALOG_CONFIG_MAP_NAME = "cray-product-catalog"
PRODUCT_CATALOG_CONFIG_MAP_NS = "services"

# The strings used in the product catalog to identify node arch
PRODCAT_ARM_ARCH = "aarch64"
PRODCAT_X86_ARCH = "x86_64"
PRODCAT_ARCH_STRINGS = { ARM_ARCH: PRODCAT_ARM_ARCH, X86_ARCH: PRODCAT_X86_ARCH }
