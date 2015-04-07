LCR System
==========

In voice telecommunications, least-cost routing (LCR) is the process of selecting the path of outbound communications traffic based on cost. Within a telecoms carrier, an LCR team might periodically (monthly, weekly or even daily) choose between routes from several or even hundreds of carriers for destinations across the world. This function might also be automated by a device or software program known as a "Least Cost Router." [WIKI2015]_

Data structures
---------------
The LCR rule parameters are: Direction, Tenant, Category, Account, Subject, DestinationId, RPCategory, Strategy, StrategyParameters, ActivationTime, Weight.

The first five are used to match the rule for a specific call descriptor. They can have a value or marked as \*any.

The DestinationId can be used to filter the LCR rules entries or to make the rule more specific.

RPCategory is used to indicate the rating profile category.

Strategy indicates supplier selection algorithm and StrategyParams will be specific to each strategy. Strategy can be one of the following:

\*static (filter)
  Will use the suppliers provided as params.
  StrategyParams: suppier1;supplier2;etc
  
\*lowest_cost (sorting)
  Matching suppliers will be sorted by ascending cost.
  StrategyParams: None

\*highest_cost (sorting)
  Matching suppliers will be sorted by descending cost.
  StrategyParams: None

\*qos_with_threshold (filter)
  The system will reject the suppliers that have out of bounds average success ratio or average call duration.
  StrategyParams: min_asr;max_asr;min_acd;max_acd 

\*qos (sorting)
  The system will sort by metrics in the order of appearance.
  StrategyParams: metric1;metric2;etc

ActivationTime is the date/time when the LCR entry starts to be active.

Weight is used to further sort the rules. It passed untouched and unused by the system for the session manager use only.

Example
+++++++

::
   
     *in, cgrates.org,call,*any,*any,EU_LANDLINE,LCR_STANDARD,*static,ivo;dan;rif,2012-01-01T00:00:00Z,10

Code implementation
-------------------
The general process of getting LCRs is this.

The LCR rules for a specific call descriptor are searched using direction, tenant, category, account and subject of the call descriptor matched as strictly as possible with LCR rules.

Because a rule can have several entries they will be sorted by activation time.

Next the system will find out if one ore more LCR entries apply to this call considering entry's  activation time. If more than one applies the call will be split into LCR time-spans attaching the corresponding entry to each timespan.

Each timespan is than iterated and acted upon according to it's entry strategy. For static strategy the cost is calculated for each supplier found in the parameters and the suppliers are listed as they are found.

For the QOS strategies the suppliers are searched using call descriptor parameters (direction, tenant, category, account, subject), than the cdrstats module is queried for the QOS values and the suppliers are filtered or sorted according to the StrategyParameters field.

For the lowest/highest cost strategies the matched suppliers are sorted ascending/descending on cost.

::
   
   [{
    "Entry": {
        "DestinationId": "*any",
        "RPCategory": "LCR_STANDARD",
        "Strategy": "*lowest_cost",
        "StrategyParams": "",
        "Weight": 20
    },
    "StartTime": "2015-04-06T17:40:00Z",
    "SupplierCosts": [{"Supplier":"rif", Cost:"2.0"},{"Supplier":"dan", Cost:"1.0"}]
    }]

.. [WIKI2015] http://en.wikipedia.org/wiki/Least-cost_routing
