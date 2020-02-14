AttributeS
==========

**AttributeS** is a standalone subsystem within **CGRateS** and it is the equivalent of a key-value store. It is accessed via :ref:`CGRateS RPC APIs<remote-management>`.

As most of the other subsystems, it is performance oriented, stored inside *DataDB* but cached inside the *cgr-engine* process. 
Caching can be done dynamically/on-demand or at start-time/precached and it is configurable within "cache" section in the .json configuration file.

Selection
---------

It is able to process generic events (hashmaps) and decision for matching it is outsourced to :ref:`FilterS`.

If there are multiple profiles (configurations) matching, the one with highest *Weight* will be the winner. There can be only one *AttributeProfile* processing the event per *process run*. If one configures multiple *process runs* either in  :ref:`JSON configuration <configuration>` or as parameter to the *.ProcessEvent* API call, the output event from one *process run* will be forwarded as input to the next selected profile. There will be independent *AttributeProfile* selection performed for each run, hence the event fields modified in one run can be applied as filters to the next *process run*, giving out the possibility to chain *AttributeProfiles* and have multiple replacements with a minimum of performance penalty (in-memory matching).


Parameters
----------


AttributeProfile
^^^^^^^^^^^^^^^^

Tenant
 	The tenant on the platform (one can see the tenant as partition ID)
 
ID
 	Identifier for the *AttributeProfile*, unique within a *Tenant*
 
Context
	A list of *contexts* applying to this profile. A *context* is usually associated with a logical phase during event processing (ie: *\*sessions* or *\*cdrs* for events parsed by :ref:`SessionS` or :ref:`CDRs`)

FilterIDs
	List of *FilterProfiles* which should match in order to consider the *AttributeProfile* matching the event.

ActivationInterval
	The time interval when this profile becomes active. If undefined, the profile is always active. Other options are start time, end time or both.

Blocker
	In case of multiple *process runs* are allowed, this flag will break further processing.

Weight
	Used in case of multiple profiles matching an event. The higher, the better (0 has lowest possible priority).

Attributes
	List of [Attribute](#attribute) objects part of this profile


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

  	**\*usage_difference**
  		Will calculate the duration difference between two field names defined in the *Value*. If the number of fields in the *Value* are different than 2, it will error.

  	**\*sum** 
  		Will sum up the values in the *Value*.

  	**\*value_exponent**
  		Will compute the exponent of the first field in the *Value*.

Value
	The value which will be set for *Path*. It can be a list of :ref:`RSRParsers` capturing even from multiple sources in the same event. If the *Value* is *\*none* the field with *Path* will be removed from *Event*


Use cases
---------

* Fields aliasing
  * Number portability (replacing a dialed number with it's translation)
  * Roaming (using *Category* to point out the zone where the user is roaming in so we can apply different rating or  consume out of restricted account bundles).

* Appending new fields
  * Adding separate header with location information
  * Adding additional rating information (ie: SMS only contains origin+destination, add *Tenant*, *Account*, *Subject*, *RequestType*)
  * Using as query language (ie: append user password for a given user so we can perform authorization on SIP Proxy side).


