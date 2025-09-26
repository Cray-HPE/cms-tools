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
ImsImage class
"""

from typing import ClassVar

from cmstools.lib.api import request_and_check_status
from cmstools.lib.defs import CmstoolsException as BBException
from cmstools.test.barebones_image_test.log import logger
from cmstools.lib.s3 import S3Url, get_s3_artifact_etag
from cmstools.test.barebones_image_test.test_resource import TestResource
from cmstools.lib.ims.defs import IMS_ARCH_STRINGS, IMS_IMAGES_URL, IMS_V2_IMAGES_URL

# Used to indicate an uninitialized values
UNSET = object()

class ImsImage(TestResource):
    """
    Data about an image in IMS
    """
    base_url: ClassVar[str] = IMS_IMAGES_URL
    label: ClassVar[str] = "IMS image"

    def __init__(self, image_id: str):
        super().__init__(name=image_id)
        self.__s3_etag = UNSET
        self.__s3_path = UNSET
        self.__arch = UNSET

    @property
    def arch(self) -> str:
        """
        Return the architecture of the image (using the global test string format, not the
        IMS format)
        """
        if self.__arch is UNSET:
            logger.error("Programming error: attempted to view architecture of %s before loading "
                         "it from IMS", self.label_and_name)
            raise BBException()
        return self.__arch

    @property
    def s3_etag(self) -> str:
        """
        Return the S3 etag of the image
        """
        if self.__s3_etag is UNSET:
            logger.error("Programming error: attempted to view S3 etag of %s before loading "
                         "it from IMS", self.label_and_name)
            raise BBException()
        return self.__s3_etag

    @property
    def s3_path(self) -> S3Url:
        """
        Return the S3 path of the image
        """
        if self.__s3_path is UNSET:
            logger.error("Programming error: attempted to view S3 path of %s before loading "
                         "it from IMS", self.label_and_name)
            raise BBException()
        return self.__s3_path

    @property
    def v2_url(self) -> str:
        """
        Return the IMS v2 URL to this image
        """
        return f"{IMS_V2_IMAGES_URL}/{self.name}"

    def delete(self) -> None:
        """
        Calls the API to delete this IMS image and its associated S3 artifacts.
        Raises an exception if there are problems.
        """
        # make the call to IMS to do the delete
        logger.info("Deleting %s and associated S3 artifacts", self.label_and_name)
        request_and_check_status("delete", self.v2_url, expected_status=204, parse_json=False,
                                 params={"cascade": True})
        logger.info("Deleted %s and associated S3 artifacts", self.label_and_name)

    def load_from_ims(self) -> None:
        """
        Load data about the specified image from IMS and return an ImsImage based on it.
        """
        image = self.get()

        try:
            image_link = image['link']
        except KeyError as exc:
            logger.exception("%s does not have a 'link' field: %s", self.label_and_name, image)
            raise BBException() from exc

        try:
            s3_etag = image_link['etag']
            s3_path = image_link['path']
        except KeyError as exc:
            logger.exception("%s does not have a 'link.%s' field: %s", self.label_and_name, exc,
                             image)
            raise BBException() from exc
        try:
            ims_arch=image['arch']
        except KeyError as exc:
            logger.exception("%s does not have an 'arch' field: %s", self.label_and_name, image)
            raise BBException() from exc

        if not s3_path:
            logger.error("%s has an empty or null 'link.path' field: %s", self.label_and_name,
                         image)
            raise BBException()
        s3_path = S3Url(s3_path)
        if not ims_arch:
            logger.error("%s has an empty or null 'arch' field: %s", self.label_and_name, image)
            raise BBException()
        if not s3_etag:
            logger.warning("%s has an empty or null 'link.etag' field: %s", self.label_and_name,
                         image)
            logger.info("Will attempt to get the etag directly from S3")
            s3_etag = get_s3_artifact_etag(s3_path)
            if not s3_etag:
                logger.error("Even in S3 (%s), %s has an empty or null 'link.etag' field", s3_path,
                             self.label_and_name)
                raise BBException()

        for arch_string, ims_arch_string in IMS_ARCH_STRINGS.items():
            if ims_arch == ims_arch_string:
                self.__arch = arch_string
                self.__s3_etag = s3_etag
                self.__s3_path = s3_path
                return
        logger.error("%s has unsupported 'arch' in IMS: %s", self.label_and_name, image)
        raise BBException()
