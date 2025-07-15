.. _prometheus_agent:

PrometheusAgent
===============

**PrometheusAgent** is a CGRateS component that exposes metrics for Prometheus monitoring systems. It serves as a bridge between CGRateS and Prometheus by collecting and exposing metrics from:

1. **Core metrics** - collected from configured CGRateS engines via CoreSv1.Status API
2. **StatQueue metrics** - values from CGRateS :ref:`StatS <stats>` component, collected via StatSv1.GetQueueFloatMetrics API
3. **Cache statistics** - collected from configured :ref:`CacheS <caches>` components via CacheSv1.GetCacheStats API

For core metrics, the agent computes real-time values on each Prometheus scrape request. For StatQueue metrics, it retrieves the current state of the stored StatQueues without additional calculations. For cache statistics, it collects current cache utilization data from the configured cache partitions.

Configuration
-------------

Example configuration in the JSON file:

.. code-block:: json

    "prometheus_agent": {
        "enabled": true,
        "path": "/prometheus",
        "caches_conns": ["*internal"],
        "cache_ids": [
            "*attribute_filter_indexes",
            "*charger_profiles",
            "*rpc_connections"
        ],
        "cores_conns": ["*internal", "external"],
        "stats_conns": ["*internal", "external"],
        "stat_queue_ids": ["cgrates.org:SQ_1", "SQ_2"]
    }

The default configuration can be found in the :ref:`configuration` section.

Parameters
----------

enabled
    Enable the PrometheusAgent module. Possible values: <true|false>

path
    HTTP endpoint path where Prometheus metrics will be exposed, e.g., "/prometheus" or "/metrics"

caches_conns
    List of connection IDs to CacheS components for collecting cache statistics. Empty list disables cache metrics collection. Possible values: <""|*internal|$rpc_conns_id>

cache_ids
    List of cache partition IDs to collect statistics for. Available cache IDs can be found in the caches.partitions section of the default configuration. Empty list collects statistics for all available cache partitions.

cores_conns
    List of connection IDs to CoreS components for collecting core metrics. Empty list disables core metrics collection. Possible values: <""|*internal|$rpc_conns_id>

stats_conns
    List of connection IDs to StatS components for collecting StatQueue metrics. Empty list disables StatQueue metrics collection. Possible values: <""|*internal|$rpc_conns_id>

stat_queue_ids
    List of StatQueue IDs to collect metrics from. Can include tenant in format <[tenant]:ID>. If tenant is not specified, default tenant from general configuration is used.

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
    StatQueue metrics don't include node_id labels since StatQueues can be shared between CGRateS instances. Users should ensure StatQueue IDs are unique across their environment.

2. **Core Metrics** (when cores_conns is configured)
    - Standard Go runtime metrics (go_goroutines, go_memstats_*, etc.)
    - Standard process metrics (process_cpu_seconds_total, process_open_fds, etc.)
    - Node identification via "node_id" label, allowing multiple CGRateS engines to be monitored

    Example of core metrics output:

    .. code-block:: none

        # HELP go_goroutines Number of goroutines that currently exist.
        # TYPE go_goroutines gauge
        go_goroutines{node_id="e94160b"} 40

        # HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
        # TYPE process_cpu_seconds_total counter
        process_cpu_seconds_total{node_id="e94160b"} 0.34

        # HELP go_memstats_alloc_bytes Number of bytes allocated in heap and currently in use.
        # TYPE go_memstats_alloc_bytes gauge
        go_memstats_alloc_bytes{node_id="e94160b"} 1.1360808e+07

3. **Cache Metrics** (when caches_conns is configured)
    - Two separate metrics for cache statistics: ``cgrates_cache_groups_total`` and ``cgrates_cache_items_total`` with cache partition ID label
    - Obtained from CacheS services on each scrape request
    - Useful for identifying memory usage patterns and potential performance issues

    Example of cache metrics output:

    .. code-block:: none

        # HELP cgrates_cache_groups_total Total number of cache groups
        # TYPE cgrates_cache_groups_total gauge
        cgrates_cache_groups_total{cache="*attribute_filter_indexes"} 2
        cgrates_cache_groups_total{cache="*charger_profiles"} 0
        cgrates_cache_groups_total{cache="*rpc_connections"} 0

        # HELP cgrates_cache_items_total Total number of cache items
        # TYPE cgrates_cache_items_total gauge
        cgrates_cache_items_total{cache="*attribute_filter_indexes"} 6
        cgrates_cache_items_total{cache="*charger_profiles"} 2
        cgrates_cache_items_total{cache="*rpc_connections"} 1


How It Works
------------

The PrometheusAgent operates differently than other CGRateS components that use connection failover:

- When multiple connections are configured in stats_conns, the agent collects metrics from **all** connections, not just the first available one
- When multiple connections are configured in cores_conns, the agent attempts to collect metrics from **all** connections, labeling them with their respective node_id
- When multiple connections are configured in caches_conns, the agent collects cache statistics from **all** connections for the specified cache_ids
- The agent processes metrics requests only when Prometheus sends a scrape request to the configured HTTP endpoint

You can view all exported metrics and see what Prometheus would scrape by making a simple curl request to the HTTP endpoint:

.. code-block:: bash

    curl http://localhost:2080/prometheus
