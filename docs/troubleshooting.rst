Troubleshooting
===============

.. contents::
   :local:
   :depth: 3

Profiling
---------

This section covers how to set up and use profiling tools for ``cgrates`` to diagnose performance issues. You can enable profiling through configuration, runtime flags, or APIs.

For more information on profiling in Go, see the `Go Diagnostics: Profiling <https://go.dev/doc/diagnostics#profiling>`_ documentation.

Configuration
~~~~~~~~~~~~~

There are two ways to set up profiling:

Using JSON Configuration
^^^^^^^^^^^^^^^^^^^^^^^^

Enable profiling by adding the ``pprof_path`` under the ``http`` section in your JSON config file:

.. code-block:: json

   {
       "listen": {
           "http": ":2080",
           "http_tls": ":2280"
       },
       "http": {
           "pprof_path": "/debug/pprof/"
       }
   }

Profiling is enabled by default and exposes the ``/debug/pprof/`` endpoint. You can access it through the address set in the ``listen`` section (``http`` or ``http_tls``). To turn off profiling, set ``pprof_path`` to an empty string ``""``.

Using Runtime Flags
^^^^^^^^^^^^^^^^^^^

You can also control profiling with runtime flags when starting ``cgr-engine``:

.. code-block:: console

   $ cgr-engine -help
   Usage of cgr-engine:
    -cpuprof_dir string
          Directory for CPU profiles
    -memprof_dir string
          Directory for memory profiles
    -memprof_interval duration
          Interval between memory profile saves (default 15s)
    -memprof_maxfiles int
          Number of memory profiles to keep (most recent) (default 1)
    -memprof_timestamp
          Add timestamp to memory profile files

Generating Profile Data
~~~~~~~~~~~~~~~~~~~~~~~

Let's assume the profiling interface is available at ``http://localhost:2080/debug/pprof/``.

.. note::
   Profiling started with flags or APIs can be stopped using the corresponding API calls. If you start profiling on startup using flags and don't stop it manually, a profile will be automatically generated when the engine shuts down. The same applies if you start profiling via API and forget to stop it before shutting down the engine.

CPU Profiling
^^^^^^^^^^^^^

Here's how to generate CPU profile data:

- **Web Browser**: Go to ``http://localhost:2080/debug/pprof/`` in your browser. Click "profile" to start a 30-second CPU profile.
- **Custom Duration**: Add the ``seconds`` parameter to set a different duration: ``http://localhost:2080/debug/pprof/profile?seconds=5``.
- **Command Line**: Use ``curl`` to download the profile:

  .. code-block:: console

     curl -o cpu.prof http://localhost:2080/debug/pprof/profile?seconds=5

- **APIs**: Use ``CoreSv1.StartCPUProfiling`` and ``CoreSv1.StopCPUProfiling`` APIs:

  .. code-block:: json

     {
         "method": "CoreSv1.StartCPUProfiling",
         "params": [{
             "DirPath": "/tmp"
         }],
         "id": 1
     }

     {
         "method": "CoreSv1.StopCPUProfiling",
         "params": [],
         "id": 1
     }

- **Startup Profiling**: Profile the entire runtime by specifying a directory with the ``-cpuprof_dir`` flag:

  .. code-block:: console

     cgr-engine -cpuprof_dir=/tmp [other flags]

Memory Profiling
^^^^^^^^^^^^^^^^

Generate memory profile data like this:

- **Web Browser**: Visit ``http://localhost:2080/debug/pprof/`` to create a memory snapshot. Use ``?debug=2`` (or ``?debug=1``) for human-readable output. If the ``debug`` parameter is omitted or set to ``0``, the output will be in binary format.
- **Command Line**: Use ``curl`` to download the memory profile:

  .. code-block:: console

     curl -o mem.prof http://localhost:2080/debug/pprof/heap

- **Automated Profiling**: Use the ``CoreSv1.StartMemoryProfiling`` API for periodic memory snapshots:

  .. code-block:: json

     {
         "method": "CoreSv1.StartMemoryProfiling",
         "params": [{
             "DirPath": "/tmp",
             "Interval": 5000000000,
             "MaxFiles": 5,
             "UseTimestamp": true
         }],
         "id": 1
     }

  .. note::

     ``Interval`` is in nanoseconds. Future updates will allow using time strings (e.g., ``5s``, ``1h``) or seconds as an integer.

Other Useful Profiles
^^^^^^^^^^^^^^^^^^^^^

The ``/debug/pprof/`` endpoint offers more useful profiles:

- **Goroutine Profile** (``/debug/pprof/goroutine``): View or download goroutine stack dumps.
- **Mutex Profile** (``/debug/pprof/mutex``): Find bottlenecks where goroutines wait for locks.
- **Block Profile** (``/debug/pprof/block``): Identify where goroutines block waiting on synchronization primitives.
- **Thread Create Profile** (``/debug/pprof/threadcreate``): Show stack traces that led to the creation of new OS threads.
- **Execution Trace**: For information on generating and analyzing execution traces, see the `Tracing`_ section below.

Analyzing Profiles
~~~~~~~~~~~~~~~~~~

The main tool for analyzing profiles is ``go tool pprof``. It helps visualize and analyze profiling data. You can use it with both downloaded profile files and directly with URLs.

Command-Line Analysis
^^^^^^^^^^^^^^^^^^^^^

For CPU profiles:

.. code-block:: console

   go tool pprof cpu.prof
   # or
   go tool pprof http://localhost:2080/debug/pprof/profile

For memory profiles:

.. code-block:: console

   go tool pprof mem.prof
   # or
   go tool pprof http://localhost:2080/debug/pprof/heap

This opens an interactive terminal. Use commands like ``top``, ``list``, ``web``, and ``svg`` to explore the profile.

.. hint::
   Run ``go tool pprof -h`` for more information on available commands and options.

Visual Analysis
^^^^^^^^^^^^^^^

Create visual representations of your profiling data:

- **SVG**: Generate an SVG graph:

  .. code-block:: console

     go tool pprof -svg cpu.prof > cpu.svg
     # or
     go tool pprof -svg mem.prof > mem.svg

- **Web Interface**: Use ``-http`` for an interactive visualization in your browser:

  .. code-block:: console

     go tool pprof -http=:8080 cpu.prof
     # or
     go tool pprof -http=:8080 mem.prof

  .. note:: 

     You might need to install the ``graphviz`` package.

Tracing
-------

Execution tracing provides a detailed view of runtime behavior of your Go program.

For detailed information on tracing in Go, see the `Go Diagnostics: Execution Tracing <https://go.dev/doc/diagnostics#tracing>`_ documentation.

To generate and analyze trace data:

.. code-block:: console

   # Generate trace data
   curl -o trace.out http://localhost:2080/debug/pprof/trace?seconds=5

   # Analyze trace data
   go tool trace trace.out

This opens a browser interface for detailed execution analysis.

Debugging
---------

This section covers how to set up and use Delve, a Go debugger, with ``cgrates``.

For detailed information on debugging Go programs, see the `Go Diagnostics: Debugging <https://go.dev/doc/diagnostics#debugging>`_ documentation.

Installation
~~~~~~~~~~~~

To install Delve, run:

.. code-block:: console

   go install github.com/go-delve/delve/cmd/dlv@latest

Basic Usage
~~~~~~~~~~~

There are several ways to use Delve with ``cgrates``:

1. Start ``cgr-engine`` in debug mode:

   .. code-block:: console

      dlv exec /path/to/cgr-engine -- --config_path=/etc/cgrates --logger=*stdout

2. Attach to a running instance:

   .. code-block:: console

      ENGINE_PID=$(pidof cgr-engine)
      dlv attach $ENGINE_PID

3. Debug tests:

   .. code-block:: console

      dlv test github.com/cgrates/cgrates/apier/v1 -- -test.run=TestName


.. hint::
   For better debugging, disable optimizations (``-N``) and inlining (``-l``) when building ``cgr-engine``:

   .. code-block:: console

      go install -gcflags="all=-N -l" -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-engine

Handling Crashes
~~~~~~~~~~~~~~~~

To capture more information when ``cgrates`` crashes:

1. Enable core dump generation:

   .. code-block:: console

      ulimit -c unlimited
      GOTRACEBACK=crash cgr-engine -config_path=/etc/cgrates

2. Analyze core dumps with Delve:

   .. code-block:: console

      dlv core /path/to/cgr-engine core

Common Debugging Commands
~~~~~~~~~~~~~~~~~~~~~~~~~

Once in a Delve debug session, you can use these common commands:

- ``break`` or ``b``: Set a breakpoint
- ``continue`` or ``c``: Run until breakpoint or program termination
- ``next`` or ``n``: Step over to next line
- ``step`` or ``s``: Step into function call
- ``print`` or ``p``: Evaluate an expression
- ``goroutines``: List current goroutines
- ``help``: Show help for commands

For more information on using Delve, refer to the `Delve Documentation <https://github.com/go-delve/delve/tree/master/Documentation>`_.

Further Reading
---------------

For more comprehensive information on Go diagnostics, profiling, and debugging, check out these resources:

- `Go Diagnostics <https://go.dev/doc/diagnostics>`_: Official documentation on diagnostics in Go.
- `Profiling Go Programs <https://go.dev/blog/pprof>`_: In-depth blog post on profiling in Go.
- `net/http/pprof godoc <https://pkg.go.dev/net/http/pprof>`_: Documentation for the pprof package.
- `Delve Debugger <https://github.com/go-delve/delve>`_: GitHub repository for the Delve debugger.
