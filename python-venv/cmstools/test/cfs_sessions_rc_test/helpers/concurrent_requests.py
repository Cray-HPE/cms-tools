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
from collections.abc import Callable
from dataclasses import dataclass
from typing import TypeVar

from cmstools.test.cfs_sessions_rc_test.log import logger

T = TypeVar('T')


@dataclass
class RequestBatch:
    """Configuration for a batch of requests."""
    max_parallel: int
    request_func: Callable[..., None]


class ConcurrentRequestManager:
    """Manage parallel API requests using threading."""

    def create_batch(self, batch: RequestBatch) -> list[threading.Thread]:
        """
        Create and return a list of worker threads for batch requests.

        Args:
            batch: RequestBatch configuration containing items and request function

        Returns:
            List of Thread objects ready to be started
        """
        threads = []

        logger.info("Creating %d worker threads for batch execution", batch.max_parallel)

        for _ in range(batch.max_parallel):
            thread = threading.Thread(
                target=batch.request_func
            )
            threads.append(thread)

        logger.debug("Created %d threads", len(threads))
        return threads

    def execute_batch(self, threads: list[threading.Thread], shuffle: bool = False) -> None:
        """
        Start and join all threads in the provided list.

        Args:
            threads: List of Thread objects to execute
            shuffle: If True, randomize thread start order to simulate race conditions
        """
        thread_count = len(threads)
        logger.info("Starting execution of %d threads", thread_count)

        if shuffle:
            logger.debug("Shuffling thread start order")
            random.shuffle(threads)

        # Start all threads
        for thread in threads:
            thread.start()
        logger.debug("Started all %d threads", thread_count)

        # Wait for all threads to complete
        for idx, thread in enumerate(threads):
            thread.join()
            logger.debug("Joined thread %d/%d", (idx + 1), thread_count)

        logger.info("Batch execution complete: all %d threads finished", thread_count)

    def create_batch_with_items(self, items: list, batch: RequestBatch) -> list[threading.Thread]:
        """
        Create and return a list of worker threads for batch requests with items.
        Args:
            items: List of items to process
            batch: RequestBatch configuration containing max_parallel and request function
        Returns:
            List of Thread objects ready to be started
        """
        threads = []

        logger.info("Creating %d worker threads (one per item) for batch execution", len(items))

        for item in items:
            thread = threading.Thread(
                target=batch.request_func,
                args=(item,)
            )
            threads.append(thread)
        return threads

    def create_batch_with_pool(self, items: list, batch: RequestBatch) -> list[threading.Thread]:
        """
        Create and return a list of worker threads for batch requests with items using a thread pool.
        Args:
            items: List of items to process
            batch: RequestBatch configuration containing max_parallel and request function
        Returns:
            List of Thread objects ready to be started
        """
        threads = []
        item_count = len(items)
        pool_size = min(batch.max_parallel, item_count)

        logger.info("Creating thread pool of size %d for batch execution with items", pool_size)

        for i in range(pool_size):
            thread = threading.Thread(
                target=self._thread_pool_worker,
                args=(items[i::pool_size], batch.request_func)
            )
            threads.append(thread)
        return threads

    def _thread_pool_worker(self, items: list[T], request_func: Callable[[T], None]) -> None:
        """
        Worker function that processes a subset of items from the pool.
        Args:
            items: List of items assigned to this worker thread
            request_func: Function to call for each item
        """
        logger.info("Thread pool worker starting with %d items", len(items))
        exceptions = []
        for item in items:
            try:
                request_func(item)
            except Exception as e:
                exceptions.append(e)
        if exceptions:
            if len(exceptions) == 1:
                raise exceptions[0]
            raise ExceptionGroup("Thread pool worker encountered exceptions", exceptions)

        logger.info("Thread pool worker completed processing %d items", len(items))
