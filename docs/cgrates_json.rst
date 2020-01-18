
.. _engine_configuration:

4.1. cgr-engine configuration file
==================================

Has a *JSON* format with commented lines starting with *//*.
Organized into configuration sections which offers the advantage of being easily splitable. 
All configuration options come with defaults and we have tried our best to choose the best ones for a minimum of efforts necessary when running.
Can be loaded from local folders or remotely using http transport.

.. hint:: You can split the configuration into any number of *.json* files/directories since the :ref:cgr-engine loads recursively the complete configuration folder, alphabetically ordered



Below is the default configuration file which comes hardcoded into :ref:cgr-engine:

.. literalinclude:: ../data/conf/cgrates/cgrates.json
   :language: javascript
   :linenos:
