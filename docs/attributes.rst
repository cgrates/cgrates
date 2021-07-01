.. _attributes:

AttributeS
==========

**AttributeS** is a standalone subsystem within **CGRateS** and it is the equivalent of a key-value store. It is accessed via `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_.

As most of the other subsystems, it is performance oriented, stored inside *DataDB* but cached inside the *cgr-engine* process. 
Caching can be done dynamically/on-demand or at start-time/precached and it is configurable within *cache* section in the :ref:`JSON configuration <configuration>`.

Selection
---------

It is able to process generic events (hashmaps) and decision for matching it is outsourced to :ref:`FilterS`.

If there are multiple profiles (configurations) matching, the one with highest *Weight* will be the winner. There can be only one *AttributeProfile* processing the event per *process run*. If one configures multiple *process runs* either in  :ref:`JSON configuration <configuration>` or as parameter to the *.ProcessEvent* API call, the output event from one *process run* will be forwarded as input to the next selected profile. There will be independent *AttributeProfile* selection performed for each run, hence the event fields modified in one run can be applied as filters to the next *process run*, giving out the possibility to chain *AttributeProfiles* and have multiple replacements with a minimum of performance penalty (in-memory matching).


Parameters
----------


AttributeS
^^^^^^^^^^

**AttributeS** is the **CGRateS** component responsible of handling the *AttributeProfiles*.

It is configured within **attributes** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
  Will enable starting of the service. Possible values: <true|false>.

indexed_selects
  Enable profile matching exclusively on indexes. If not enabled, the *ResourceProfiles* are checked one by one which for a larger number can slow down the processing time. Possible values: <true|false>.

string_indexed_fields
  Query string indexes based only on these fields for faster processing. If commented out, each field from the event will be checked against indexes. If uncommented and defined as empty list, no fields will be checked.

prefix_indexed_fields
  Query prefix indexes based only on these fields for faster processing. If defined as empty list, no fields will be checked.

nested_fields
  Applied when all event fields are checked against indexes, and decides whether subfields are also checked.

process_runs
  Limit the number of loops when processing an Event. The event loop is however clever enough to stop when the same processing occurs or no more additional profiles are matching, so higher numbers are ignored if not needed.

.. _AttributeProfile:

AttributeProfile
^^^^^^^^^^^^^^^^

Represents the configuration for a group of attributes applied.

Tenant
 	The tenant on the platform (one can see the tenant as partition ID)
 
ID
 	Identifier for the *AttributeProfile*, unique within a *Tenant*
 
FilterIDs
	List of *FilterProfiles* which should match in order to consider the *AttributeProfile* matching the event.

Blocker
	In case of multiple *process runs* are allowed, this flag will break further processing.

Weight
	Used in case of multiple profiles matching an event. The higher, the better (0 has lowest possible priority).

Attributes
	List of :ref:`Attribute` objects part of this profile.


.. _Attribute:

Attribute
^^^^^^^^^

FilterIDs
	List of *FilterProfiles* which should match in order to consider the *Attribute* matching the event.

Path
	Identifying the targeted absolute path within the processed event.

Type
	Represents the type of substitution which will be performed on the Event. The following *Types* are available:

	**\*constant**
		The *Value* is a constant value, it will just set the *FieldName* to this value as it is.

  	**\*variable**
  		The *Value* is a *RSRParser* which will be able to capture the value out of one or more fields in the event (also combined with other constants) and write it to *Path*.

  	**\*composed** 
  		Same as *\*variable* but instead of overwriting *Path*, it will append to it.

  	**\*usageDifference**
  		Will calculate the duration difference between two field names defined in the *Value*. If the number of fields in the *Value* are different than 2, it will error.

  	**\*sum** 
  		Will sum up the values in the *Value*.

  	**\*valueExponent**
  		Will compute the exponent of the first field in the *Value*.

Value
	The value which will be set for *Path*. It can be a list of RSRParsers capturing even from multiple sources in the same event. If the *Value* is *\*remove* the field with *Path* will be removed from *Event*


Inline Attribute 
^^^^^^^^^^^^^^^^

In order to facilitate quick attribute definition (without the need of separate *AttributeProfile*), one can define attributes directly as *AttributeIDs* following the special format.

Inline filter format::
 
 attributeType:attributePath:attributeValue

Example::
 
 *constant:*req.RequestType:*prepaid


Use cases
---------

* Fields aliasing
  * Number portability (replacing a dialed number with it's translation)
  * Roaming (using *Category* to point out the zone where the user is roaming in so we can apply different rating or  consume out of restricted account bundles).

* Appending new fields
  * Adding separate header with location information
  * Adding additional rating information (ie: SMS only contains origin+destination, add *Tenant*, *Account*, *Subject*, *RequestType*)
  * Using as query language (ie: append user password for a given user so we can perform authorization on SIP Proxy side).


