TariffPlans
===========

Major concept within CGRateS architecture, implement mechanisms to load rating as well as account data into CGRateS.

Currently TariffPlans can be loaded using 2 different approaches:

Direct load out of TP-CSV files 
-------------------------------

This represents the fastest and easiest way to manage small set of TP definitions. It has the advantage of being simple to define and load but on the other hand as soon as the data set grows it becomes relatively hard to be maintaned.

Due to complex data definition we have split information necessary on each load process in more .csv files, identified by names close to their utility.

Each individual CSV file can have any number of rows starting with comment character (#) which will be ignored on processing.

Depending on CGRateS subsystem, following files are required:

RatingProfiles
~~~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   csv_tptimings

.. toctree::
   :maxdepth: 2

   csv_tpdestinations

.. toctree::
   :maxdepth: 2

   csv_tprates

.. toctree::
   :maxdepth: 2

   csv_tpdestinationrates

.. toctree::
   :maxdepth: 2

   csv_tpdestratetimings

.. toctree::
   :maxdepth: 2

   csv_tpratingprofiles

Accounts
~~~~~~~~

.. toctree::
   :maxdepth: 2

   csv_tptimings

.. toctree::
   :maxdepth: 2

   csv_tpdestinations

.. toctree::
   :maxdepth: 2

   csv_tpactions

.. toctree::
   :maxdepth: 2

   csv_tpactiontimings

.. toctree::
   :maxdepth: 2

   csv_tpactiontriggers

.. toctree::
   :maxdepth: 2

   csv_tpaccountactions


