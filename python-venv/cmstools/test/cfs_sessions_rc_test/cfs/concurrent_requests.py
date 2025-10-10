#
# MIT License
#
# (C) Copyright 2025 Hewlett Packard Enterprise Development LP
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
Class for creating and managing concurrent API requests using threading.
"""

import threading
import random
from typing import Callable, List, TypeVar
from dataclasses import dataclass

from cmstools.test.cfs_sessions_rc_test.log import logger

T = TypeVar('T')


@dataclass
class RequestBatch:
    """Configuration for a batch of requests."""
    max_parallel: int
    request_func: Callable[[], None]


class ConcurrentRequestManager:
    """Manage parallel API requests using threading."""

    def __init__(self):
       pass

    def create_batch(self, batch: RequestBatch) -> List[threading.Thread]:
        """
        Create and return a list of worker threads for batch requests.

        Args:
            batch: RequestBatch configuration containing items and request function

        Returns:
            List of Thread objects ready to be started
        """
        threads = []

        logger.info(f"Creating {batch.max_parallel} worker threads for batch execution")

        for _ in range(batch.max_parallel):
            thread = threading.Thread(
                target=batch.request_func
            )
            threads.append(thread)

        logger.debug(f"Created {len(threads)} threads")
        return threads

    def execute_batch(self, threads: List[threading.Thread], shuffle: bool = False) -> None:
        """
        Start and join all threads in the provided list.

        Args:
            threads: List of Thread objects to execute
            shuffle: If True, randomize thread start order to simulate race conditions
        """
        thread_count = len(threads)
        logger.info(f"Starting execution of {thread_count} threads")

        if shuffle:
            logger.debug("Shuffling thread execution order")
            random.shuffle(threads)

        # Start all threads
        for idx, thread in enumerate(threads):
            thread.start()
            logger.debug(f"Started thread {idx + 1}/{thread_count}")

        # Wait for all threads to complete
        for idx, thread in enumerate(threads):
            thread.join()
            logger.debug(f"Joined thread {idx + 1}/{thread_count}")

        logger.info(f"Batch execution complete: all {thread_count} threads finished")