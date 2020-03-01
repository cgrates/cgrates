.. _CDRs:

CDRs
====


**CDRs** is a standalone subsystem within **CGRateS** responsible to process *CDR* events. It is accessed via :ref:`CGRateS RPC APIs<remote-management>` or separate *HTTP handlers* configured within *http* section inside :ref:`JSON configuration <configuration>`.

Due to multiple interfaces exposed, the **CDRs** is designed to function as centralized server for *CDRs* received from various sources. Examples of such sources are:
	*\*real-time events* from interfaces like *Diameter*, *Radius*, *Asterisk*, *FreeSWITCH*, *Kamailio*, *OpenSIPS*
	* \*files* like *.csv*, *.fwv*, *.xml*, *.json*.
	* \*database events* like *sql*, *kafka*, *rabbitmq*.


APIs logic
----------

ProcessEvent
^^^^^^^^^^^^

Receives the CDR in the form of *CGRateS Event* together with processing flags attached. Activating of the flags will trigger specific processing mechanisms for the CDR. Missing of the flags will be interpreted based on defaults. The following flags are available, based on the processing order:

\*attributes
	Will process the event with :ref:`AttributeS`. This allows modification of content in early stages of processing(ie: add new fields, modify or remove others). Defaults to *true* if there are connections towards :ref:`AttributeS` within :ref:`JSON configuration <configuration>`.

\*chargers
	Will process the event with :ref:`ChargerS`. This allows forking of the event into multiples. Defaults to *true* if there are connections towards :ref:`ChargerS` within :ref:`JSON configuration <configuration>`.

\*refund
	Will perform a refund for the *CostDetails* field in the event. Defaults to *false*.

\*rals
	Will calculate the *Cost* for the event using the :ref:`RALs`. If the event is *\*prepaid* the *Cost* will be attempted to be retrieved out of event or from *sessions_costs* table in the *StorDB* and if these two steps fail, :ref:`RALs` will be queried in the end. Defaults to *false*.

\*rerate
	Will re-rate the CDR as per the *\*rals* flag, doing also an automatic refund in case of *\*prepaid*, *\*postpaid* and *\*pseudoprepaid* request types. Defaults to *false*.

\*store
	Will store the *CDR* to *StorDB*. Defaults to *store_cdrs* parameter within :ref:`JSON configuration <configuration>`. If store process fails for one of the CDRs, an automated refund is performed for all derived.

\*export
	Will export the event matching export profiles. These profiles are defined within *cdre* section inside :ref:`JSON configuration <configuration>`. Defaults to *true* if there is at least one *online_cdr_exports* profile configured within :ref:`JSON configuration <configuration>`.

\*thresholds
	Will process the event with the :ref:`ThresholdS`, allowing us to execute actions based on filters set for matching profiles. Defaults to *true* if there are connections towards :ref:`ThresholdS` within :ref:`JSON configuration <configuration>`.

\*stats
	Will process the event with the :ref:`StatS`, allowing us to compute metrics based on the matching *StatQueues*. Defaults to *true* if there are connections towards :ref:`StatS` within :ref:`JSON configuration <configuration>`.


Use cases
---------

* Classic rating of your CDRs.
* Rating queues where one can receive the rated CDR few milliseconds after the *CommSwitch* has issued it. With custom export profiles there can be given the feeling that the *CommSwitch* itself sends rated CDRs.
* Rating with derived charging where we calculate automatically the cost for the same CDR multiple times (ie: supplier/customer, customer/distributor or local/premium/mobile charges).
* Fraud detection on CDR Costs with profiling.
* Improve network transparency based on monitoring Cost, ASR, ACD, PDD out of CDRs.

