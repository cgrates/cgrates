.. _CDRs:

CDRs
====


**CDRs** is a standalone subsystem within **CGRateS** responsible to process *CDR* events. It is accessed via :ref:`CGRateS RPC APIs<remote-management>` or separate *HTTP handlers* configured within *http* section inside :ref:`JSON configuration <configuration>`.

Due to multiple interfaces exposed, the **CDRs** is designed to function as centralized server for *CDRs* received from various sources. Examples of such sources are:
	*\*real-time events* from interfaces like *Diameter*, *Radius*, *Asterisk*, *FreeSWITCH*, *Kamailio*, *OpenSIPS*
	* \*files* like *.csv*, *.fwv*, *.xml*, *.json*.
	* \*database events* like *sql*, *kafka*, *rabbitmq*.

Parameters
----------


CDRs
^^^^

**CDRs** is configured within **cdrs** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

extra_fields
	Select extra fields from the request, other than the primary ones used by CGRateS (see storage schemas for listing those). Used in particular applications where the received fields are not selectable at the source(ie. FreeSWITCH JSON).

store_cdrs
	Controls storing of the received CDR within the *StorDB*. Possible values: <true|false>.

session_cost_retries
	In case of decoupling the events charging from CDRs, the charges done by :ref:`SessionS` will be stored in *sessions_costs* *StorDB* table. When receiving the CDR, these costs will be retrieved and attached to the CDR. To avoid concurrency between events and CDRs, it is possible to configure a multiple number of retries from *StorDB* table.

chargers_conns
	Connections towards :ref:`ChargerS` component to query charges for CDR events. Empty to disable the functionality.

rals_conns
	Connections towards :ref:`RALs` component to query costs for CDR events. Empty to disable the functionality.

attributes_conns
	Connections towards :ref:`AttributeS` component to alter information within CDR events. Empty to disable the functionality.

thresholds_conns
	Connections towards :ref:`ThresholdS` component to monitor and react to information within CDR events. Empty to disable the functionality.

stats_conns
	Connections towards :ref:`StatS` component to compute stat metrics for CDR events. Empty to disable the functionality.

online_cdr_exports
	List of :ref:`CDRe` profiles which will be processed for each CDR event. Empty to disable online CDR exports.



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

