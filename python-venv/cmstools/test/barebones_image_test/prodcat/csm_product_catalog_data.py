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
CsmProductCatalogData class
"""

# To support forward references in type hinting
from __future__ import annotations
from dataclasses import dataclass
import re
import yaml

import packaging.version

from cmstools.lib.defs import CmstoolsException as BBException, JsonDict
from cmstools.lib.k8s import get_k8s_configmap_data
from cmstools.test.barebones_image_test.log import logger
from cmstools.lib.prodcat.defs import (
                                       PRODCAT_ARCH_STRINGS,
                                       PRODUCT_CATALOG_CONFIG_MAP_NAME,
                                       PRODUCT_CATALOG_CONFIG_MAP_NS,
                                       VCS_URL
                                      )


@dataclass(frozen=True)
class CsmProductCatalogData:
    """
    Product Catalog data for latest CSM version
    """
    version: str
    data: JsonDict

    @property
    def configuration(self) -> JsonDict:
        """
        Return the 'configuration' stanza from this product catalog data
        """
        try:
            return self.data["configuration"]
        except KeyError as exc:
            logger.exception("No configuration stanza found in Cray Product Catalog for CSM %s",
                         self.version)
            raise BBException() from exc

    @property
    def clone_url(self) -> str:
        """
        Return the 'clone_url' field from the 'configuration' stanza of this product catalog data
        """
        try:
            prodcat_clone_url = self.configuration["clone_url"]
        except KeyError as exc:
            logger.exception("No 'clone_url' key found in configuration stanza of Cray Product "
                         "Catalog entry for CSM %s", self.version)
            raise BBException() from exc

        # Take the clone_url from the product catalog and modify it to use the
        # API gateway instead
        if "/vcs/" not in prodcat_clone_url:
            logger.error("/vcs/ not found in clone_url found in Cray Product Catalog: %s",
                         prodcat_clone_url)
            raise BBException()
        vcs_index = prodcat_clone_url.index("/vcs/")
        return f"{VCS_URL}{prodcat_clone_url[vcs_index+4:]}"

    @property
    def commit(self) -> str:
        """
        Return the 'commit' field from the 'configuration' stanza of this product catalog data
        """
        try:
            return self.configuration["commit"]
        except KeyError as exc:
            logger.exception("No 'commit' key found in configuration stanza of Cray Product "
                             "Catalog entry for CSM %s", self.version)
            raise BBException() from exc

    @property
    def major_minor(self) -> str:
        """
        Return the CSM version of this product catalog data in <major>.<minor> format
        """
        regex_find = r"^([1-9][0-9]*[.][0-9][1-9]*)(?:[.].*)?$"
        return re.sub(regex_find, r"\1", self.version)

    def barebones_image_id(self, arch: str) -> str:
        """
        Find and return the IMS ID of the metal barebones image for this CSM version
        The image name will be of the form "compute-*-<arch>
        """
        try:
            csm_images = self.data["images"]
        except KeyError as exc:
            logger.exception("No images found in Cray Product Catalog for CSM %s", self.version)
            raise BBException() from exc

        compute_name_re = re.compile(f'^compute-.+-{PRODCAT_ARCH_STRINGS[arch]}$')

        for image_name, image_data in csm_images.items():
            if not compute_name_re.match(image_name):
                logger.debug("Skipping image '%s'", image_name)
                continue
            logger.info("Found barebones image in product catalog: '%s': %s", image_name,
                        image_data)
            return image_data["id"]

        logger.error("No barebones %s compute image found in Cray Product Catalog for CSM %s",
                     arch, self.version)
        raise BBException()

    @classmethod
    def __load_csm_data(cls) -> JsonDict:
        """
        Returns the csm product entries in the Cray Product Catalog.
        """
        cpc_data = get_k8s_configmap_data(cm_name=PRODUCT_CATALOG_CONFIG_MAP_NAME,
                                          cm_namespace=PRODUCT_CATALOG_CONFIG_MAP_NS)
        try:
            csm_yaml = cpc_data["csm"]
        except KeyError as exc:
            logger.exception("No entry found for 'csm' in Cray Product Catalog")
            raise BBException() from exc
        return yaml.safe_load(csm_yaml)

    @classmethod
    def get_latest(cls) -> CsmProductCatalogData:
        """
        Looks up the csm product in the Cray Product Catalog.
        Finds the latest version.
        """
        csm_data = cls.__load_csm_data()

        # csm_data is a mapping from version strings to associated data about that csm version
        # Identify the latest version
        sorted_versions = sorted(csm_data, key=packaging.version.parse)
        latest_version = sorted_versions[-1]
        logger.info("Latest CSM version found in Product Catalog: %s", latest_version)
        return cls(data=csm_data[latest_version], version=latest_version)

    @classmethod
    def get_version(cls, version_string: str) -> CsmProductCatalogData:
        """
        Looks up the csm product in the Cray Product Catalog.
        Finds the specified version.
        """
        csm_data = cls.__load_csm_data()

        logger.info("Retrieving entry for CSM version '%s' in Product Catalog", version_string)
        try:
            return cls(data=csm_data[version_string], version=version_string)
        except KeyError as exc:
            logger.exception("No entry found for csm version '%s' in Cray Product Catalog",
                             version_string)
            raise BBException() from exc
