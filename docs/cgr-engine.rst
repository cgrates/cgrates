.. _cgr-engine:

cgr-engine
==========

Groups most of functionality from services and components.

Customisable through the use of *json* :ref:`JSON configuration <configuration>` or command line arguments (higher prio).

Able to read the configuration from either a local directory  of *.json* files with an unlimited number of subfolders (ordered alphabetically) or a list of http paths (separated by ";").

::

 $ cgr-engine -help
 Usage of cgr-engine:
  -config_path string
      Configuration directory path. (default "/etc/cgrates/")
  -cpuprof_dir string
      write cpu profile to files
  -httprof_path string
      http address used for program profiling
  -log_level int
      Log level (0-emergency to 7-debug) (default -1)
  -logger string
      logger <*syslog|*stdout>
  -memprof_dir string
      write memory profile to file
  -memprof_interval duration
      Time betwen memory profile saves (default 5s)
  -memprof_nrfiles int
      Number of memory profile to write (default 1)
  -node_id string
      The node ID of the engine
  -pid string
      Write pid file
  -scheduled_shutdown string
      shutdown the engine after this duration
  -singlecpu
      Run on single CPU core
  -version
      Prints the application version.


.. hint:: $ cgr-engine -config_path=/etc/cgrates

.. figure::  images/CGRateSInternalArchitecture.png
   :alt: CGRateS Internal Architecture
   :align: Center
   :scale: 75 %


   Internal Architecture of **cgr-engine**


The components from the diagram can be found documented in the links bellow:

.. toctree::
   :maxdepth: 1

   agents
   sessions
   rals
   cdrs
   attributes
   chargers
   resources
   suppliers
   stats
   thresholds
   filters
   dispatchers
   schedulers
   cdre
   apiers
   loaders
   datadb
   stordb
   caches
   

