.. _chargers:

ChargerS
========

**ChargerS** is a **CGRateS** subsystem designed to produce billing runs via *DerivedCharging* mechanism. 

It works as standalone component of **CGRateS**, accessible via `CGRateS RPC <https://godoc.org/github.com/cgrates/cgrates/apier/>`_ via a rich set of *APIs*. As input **ChargerS** is capable of receiving generic events (hashmaps) with dynamic types for fields.

**ChargerS** is an **important** part of the charging process within **CGRateS** since with no *ChargingProfile* matching, there will be no billing run performed.


DerivedCharging
---------------

Is a process of receiving an event as input and *deriving* that into multiples (unlimited) out. The *derived* event will be a standalone clone of original with possible modifications of individual event fields. In case of billing, this will translate into multiple Events or CDRs being billed simultaneously for the same input.


Processing logic
----------------

For the received *Event* we will retrieve the list of matching *ChargingProfiles' via :ref:`FilterS`. These profiles will be then ordered based on their *Weight* - higher *Weight* will have more priority. If no profile will match due to *Filter*, *NOT_FOUND* will be returned back to the RPC client.

Each *ChargingProfile* matching the *Event*  will produce a standalone event based on configured *RunID*. These events will each have a special field added (or overwritten), the *RunID*, which is taken from the applied *ChargingProfile*. 

If *AttributeIDs* are different than *\*none*, the newly created *Event* will be sent to [AttributeS](AttributeS) and fields replacement will be performed based on the logic there. If the *AttributeIDs* is populated, these profile IDs will be selected directly for faster processing, otherwise (if empty) the *AttributeProfiles* will be selected using :ref:`FilterS`.


Parameters
----------

ChargerProfile
^^^^^^^^^^^^^^

A *ChargerProfile* is the configuration producing the *DerivedCharging* for the Event received. It's made of the following fields:

Tenant
	Is the tenant on the platform (one can see the tenant as partition ID)

ID
	Identifier for the ChargerProfile, unique within a *Tenant*.

FilterIDs
	List of *FilterProfiles* which should match in order to consider the ChargerProfile matching the event.

RunID
	The identifier for a single bill run / charged output *Event*.

AttributeIDs
	List of *AttributeProfileIDs* which will be applied for the output *Event* in order to change some of it's fields. If empty, the list is discovered via [FilterS](FilterS) (*AttributeProfiles* matching the event). If *\*none, no AttributeProfile will be applied, event will be a simple clone of the one at input with just *RunID* being different.

Weight
	Used in case of multiple profiles matching an event. The higher, the better (0 has lowest possible priority).


Use cases
---------

* Calculating standard charges for the *Customer* calling as well as for the *Reseller*/*Distributor*. One can build chains of charging rules if multiple *Resellers* are involved.
* Calculating revenue based on *Customer* vs *Supplier* pricing.
* Calculating pricing for multiple *RouteS* for revenue protection.
* Adding *local* vs *mobile* charges for *premium numbers* when accessed from mobile headsets.
* etc.