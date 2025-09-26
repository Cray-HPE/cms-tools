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
S3Url class for cmstools test
"""

from urllib.parse import urlparse

class S3Url(str):
    """
    A string class whose value is standardized through URLparser, and with extra properties
    to display S3 bucket, key, etc

    https://stackoverflow.com/questions/42641315/s3-urls-get-bucket-name-and-path/42641363
    """

    def __new__(cls, url):
        return super().__new__(cls, urlparse(url, allow_fragments=False).geturl())

    def __init__(self, url):
        parsed = urlparse(url, allow_fragments=False)
        if parsed.query:
            self.__key = parsed.path.lstrip('/') + '?' + parsed.query
        else:
            self.__key = parsed.path.lstrip('/')
        self.__bucket = parsed.netloc

    @property
    def key(self) -> str:
        """
        Return the S3 key portion of this S3 URL
        """
        return self.__key

    @property
    def bucket(self) -> str:
        """
        Return the S3 bucket portion of this S3 URL
        """
        return self.__bucket
