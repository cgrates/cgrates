.. _cgr-engine:

cgr-engine
==========

Groups various services and components.

Customisable through the use of *json* :ref:`configuration <engine_configuration>` or command line arguments (higher prio).
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


Agents
------

**Agents** are interfaces towards external systems, implementing protocols enforced by the communication channels opened. 
These can be standard or privately defined.

All of the **Agents** implemented within CGRateS are flexible to be configured with generic parameters configurable for both *request* and *replies*.

Following *Agents* are implemented within CGRateS:

.. toctree::
   :maxdepth: 2

   diamagent
   radagent
   httpagent
   dnsagent
   astagent
   fsagent
   kamagent
   ers

