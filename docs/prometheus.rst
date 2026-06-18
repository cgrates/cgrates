.. _prometheusAgent:

PrometheusAgent
===============

**PrometheusAgent** is a CGRateS component that exposes metrics for Prometheus monitoring systems. It serves as a bridge between CGRateS and Prometheus by collecting and exposing metrics from:

1. **Core metrics** - collected from configured CGRateS engines via CoreSv1.Status API
2. **StatQueue metrics** - values from CGRateS :ref:`StatS <stats>` component, collected via StatSv1.GetQueueFloatMetrics API
3. **Cache statistics** - collected from configured :ref:`CacheS <caches>` components via CacheSv1.GetStats API

For core metrics, the agent computes real-time values on each Prometheus scrape request. For StatQueue metrics, it retrieves the current state of the stored StatQueues without additional calculations. For cache statistics, it collects current cache utilization data from the configured cache partitions.

Configuration
-------------

Example configuration in the JSON file:

.. code-block:: json

    "prometheusAgent": {
        "enabled": true,
        "path": "/prometheus",
        "conns": {
		    "*caches": [{"connIDs": ["*internal"]}],
            "*cores": [{"connIDs": ["*internal", "external"]}],
            "*stats": [{"connIDs": ["*internal", "external"]}]
	    },
        "cacheIDs": [
            "*attributeFilterIndexes",
            "*chargerProfiles",
            "*rpcConnections"
        ],
        "statQueueIDs": ["cgrates.org:SQ_1", "SQ_2"]
    }

The default configuration can be found in the :ref:`configuration` section.

Parameters
----------

enabled
    Enable the PrometheusAgent module. Possible values: <true|false>

path
    HTTP endpoint path where Prometheus metrics will be exposed, e.g., "/prometheus" or "/metrics"

/*admins
    List of connection IDs to AdminS components. Required when statQueueIDs is empty to fetch all available StatQueue profile IDs. Must match the length of *stats when auto-fetching is used. Possible values: <""|*internal|$rpc_conns_id>

/*caches
    List of connection IDs to CacheS components for collecting cache statistics. Empty list disables cache metrics collection. Possible values: <""|*internal|$rpc_conns_id>

cacheIDs
    List of cache partition IDs to collect statistics for. Available cache IDs can be found in the caches.partitions section of the default configuration. Empty list collects statistics for all available cache partitions.

/*cores
    List of connection IDs to CoreS components for collecting core metrics. Empty list disables core metrics collection. Possible values: <""|*internal|$rpc_conns_id>

/*stats
    List of connection IDs to StatS components for collecting StatQueue metrics. Empty list disables StatQueue metrics collection. Possible values: <""|*internal|$rpc_conns_id>

statQueueIDs
    List of StatQueue IDs to collect metrics from. Can include tenant in format <[tenant]:ID>. If tenant is not specified, default tenant from general configuration is used. Leave empty to automatically collect metrics from all available StatQueues (requires /*admins).

Available Metrics
-----------------

The PrometheusAgent exposes the following metrics:

1. **StatQueue Metrics**
    - Uses the naming format ``cgrates_stats_metrics`` with labels for tenant, queue, and metric type
    - Obtained from StatS services on each scrape request

    Example of StatQueue metrics output:

    .. code-block:: none

        # HELP cgrates_stats_metrics Current values for StatQueue metrics
        # TYPE cgrates_stats_metrics gauge
        cgrates_stats_metrics{metric="*acc",queue="SQ_1",tenant="cgrates.org"} 7.73779
        cgrates_stats_metrics{metric="*tcc",queue="SQ_1",tenant="cgrates.org"} 23.21337
        cgrates_stats_metrics{metric="*acc",queue="SQ_2",tenant="cgrates.org"} 11.34716
        cgrates_stats_metrics{metric="*tcc",queue="SQ_2",tenant="cgrates.org"} 34.04147

.. note::
    StatQueue metrics don't include nodeID labels since StatQueues can be shared between CGRateS instances. Users should ensure StatQueue IDs are unique across their environment.

2. **Core Metrics** (when /*cores is configured)
    - Standard Go runtime metrics (goGoroutines, go_memstats_*, etc.)
    - Standard process metrics (processCPUSecondsTotal, process_open_fds, etc.)
    - Node identification via "nodeID" label, allowing multiple CGRateS engines to be monitored

    Example of core metrics output:

    .. code-block:: none

        # HELP goGoroutines Number of goroutines that currently exist.
        # TYPE goGoroutines gauge
        goGoroutines{nodeID="e94160b"} 40

        # HELP processCPUSecondsTotal Total user and system CPU time spent in seconds.
        # TYPE processCPUSecondsTotal counter
        processCPUSecondsTotal{nodeID="e94160b"} 0.34

        # HELP goMemstatsAllocBytes Number of bytes allocated in heap and currently in use.
        # TYPE goMemstatsAllocBytes gauge
        goMemstatsAllocBytes{nodeID="e94160b"} 1.1360808e+07

3. **Cache Metrics** (when /*caches is configured)
    - Two separate metrics for cache statistics: ``cgrates_cache_groups_total`` and ``cgrates_cache_items_total`` with cache partition ID and nodeID labels
    - Obtained from CacheS services on each scrape request
    - Useful for identifying memory usage patterns and potential performance issues
    - Includes nodeID labels for multi-engine environments, allowing collection from multiple CGRateS engines

    Example of cache metrics output:

    .. code-block:: none

        # HELP cgrates_cache_groups_total Total number of cache groups
        # TYPE cgrates_cache_groups_total gauge
        cgrates_cache_groups_total{cache="*attributeFilterIndexes",nodeID="dc2cb63"} 2
        cgrates_cache_groups_total{cache="*chargerProfiles",nodeID="dc2cb63"} 0
        cgrates_cache_groups_total{cache="*rpcConnections",nodeID="dc2cb63"} 0

        # HELP cgrates_cache_items_total Total number of cache items
        # TYPE cgrates_cache_items_total gauge
        cgrates_cache_items_total{cache="*attributeFilterIndexes",nodeID="dc2cb63"} 6
        cgrates_cache_items_total{cache="*chargerProfiles",nodeID="dc2cb63"} 2
        cgrates_cache_items_total{cache="*rpcConnections",nodeID="dc2cb63"} 1


How It Works
------------

The PrometheusAgent operates differently than other CGRateS components that use connection failover:

- When multiple connections are configured in /*stats, the agent collects metrics from **all** connections, not just the first available one
- When multiple connections are configured in /*cores, the agent attempts to collect metrics from **all** connections, labeling them with their respective nodeID
- When multiple connections are configured in /*caches, the agent collects cache statistics from **all** connections for the specified cacheIDs
- The agent processes metrics requests only when Prometheus sends a scrape request to the configured HTTP endpoint

StatQueue metrics are collected based on the ``statQueueIDs`` configuration. When specific StatQueue IDs are provided, only those StatQueues are monitored. When ``statQueueIDs`` is left empty, all available StatQueues are monitored by fetching StatQueue profile IDs from the configured ``*admins``.

.. note::
    When fetching all StatQueues (empty statQueueIDs), each AdminS connection in ``*admins`` corresponds to its StatS counterpart at the same index position in ``*stats``.

You can view all exported metrics and see what Prometheus would scrape by making a simple curl request to the HTTP endpoint:

.. code-block:: bash

    curl http://localhost:2080/prometheus
