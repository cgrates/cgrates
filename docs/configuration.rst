.. _configuration:

Configuration
=============

Has a *JSON* format with commented lines starting with *//*.

Organized into configuration sections which offers the advantage of being easily splitable.

.. hint:: You can split the configuration into any number of *.json* files/directories since the :ref:cgr-engine loads recursively the complete configuration folder, alphabetically ordered


All configuration options come with defaults and we have tried our best to choose the best ones for a minimum of efforts necessary when running.

Can be loaded from local folders or remotely using HTTP transport.

The configuration can be loaded at start and reloaded at run time using APIs designed for that. This can be done either as *config pull* (reload from path) or as *config push* (the *JSON BLOB* is sent via API to the engine). 

.. hint:: You can reload from remote HTTP server as well.

Below is the default configuration file which comes hardcoded into :ref:`cgr-engine`:

.. literalinclude:: ../data/conf/cgrates/cgrates.json
   :language: javascript
   :linenos:



