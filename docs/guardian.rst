.. _guardian:

Guardian
========

Guardian is CGRateS' internal locking mechanism that ensures data consistency during concurrent operations.

What Guardian Does
------------------

Guardian prevents race conditions when multiple processes try to access or modify the same data. It uses string-based locks, typically created using some variation of the tenant and ID of the resource being protected, often with a type prefix. Guardian can use either explicit IDs or generate UUIDs internally for reference-based locking when no specific ID is provided.

When CGRateS Uses Guardian
--------------------------

Guardian protects:

* Account balance operations (debits/topups) - the most critical use case
* ResourceProfiles, Resources, StatQueueProfiles, StatQueues, ThresholdProfiles, and Thresholds while they're being used or loaded into cache
* Filter index updates
* There are other cases, but the ones listed above are the most frequent applications

Performance Implications
------------------------

Guardian affects system performance in these ways:

* Operations on the same resource are processed one after another, not simultaneously
* Under heavy load on the same resources, operations may queue up and wait
* System throughput is better when operations are distributed across different resources

Configuration
-------------

Guardian has a single configuration option:

The `locking_timeout` setting in the general configuration determines how long Guardian will hold a lock before forcing it to release. Zero timeout (no timeout) is the default and recommended setting. However, setting a reasonable timeout can help prevent system hangs if a process fails to release a lock.

When a timeout occurs, Guardian logs a warning and forces the lock to release. This keeps the system running, but the operation that timed out may fail.
