.. _FilterS:

FilterS 
=======

**FilterS** are code blocks applied to generic events (hashmaps) in order to allow/deny further processing.

A Tenant will define multiple Filter profiles via .csv or API calls. The Filter profile ID is unique within a tenant but it can be repeated over multiple Tenants.

In order to be used in event processing, a Filter profile will be attached inside another subsystem profile definition, otherwise Filter profile will have no effect on it's own. 

A subsystem can use a *Filter* via *FilterProfile* or in-line (ad-hock in the same place where subsystem profile is defined).


Filter profile 
--------------

Definition::

 type Filter struct {
	Tenant             string
	ID                 string
	Rules              []*FilterRule
 }

A Filter profile can be shared between multiple subsystem profile definitions.

A Filter profile can contain any number of Filter rules and each of them must pass in order for the filter profile to pass.

A Filter profile can be activated on specific interval, if multiple filters are used within a subsystem profile at least one needs to be active and passing in order for the subsystem profile to pass the event.


Filter rule 
-----------

Definition::

 type FilterRule struct {
	Type            string              // Filter type
	Element       	string              // Name of the field providing us the Values to check (used in case of some )
	Values          []string            // Filter definition
 }


The matching logic of each FilterRule is given by it's type.

The following types are implemented:

\*string*
	Will match in full the *Element* with at least one value defined inside *Values*.
	Any of the values matching will have the FilterRule as *matched*. 

\*notstring 
	Is the negation of *\*string*.

\*prefix
	Will match at beginning of *Element* one of the values defined inside *Values*.

\*notprefix 
	Is the negation of *\*prefix*.

\*suffix
	Will match at end of *Element* one of the values defined inside *Values*.

\*notsuffix* 
	Is the negation of *\*suffix*.

\*empty
	Will make sure that *Element* is empty or it does not exist in the event.

\*notempty 
	Is the negation of *\*empty*.

\*exists
	Will make sure that *Element* exists in the event.

\*notexists
	Is the negation of *\*exists*.

\*nottimings
	Is the negation of *\*timings*.

\*destinations
	Will make sure that the *Element* is a prefix contained inside one of the destination IDs as *Values*.

\*notdestinations
	Is the negation of *\*destinations*.

\*rsr
	Will match the *RSRFilters* defined in Values on the Element.

\*notrsr*
	Is the negation of *\*rsr*.

*\*lt* (less than), *\*lte* (less than or equal), *\*gt* (greather than), *\*gte* (greather than or equal) 
	Are comparison operators and they pass if at least one of the values defined in *Values* are passing for the *Element* of event. The operators are able to compare string, float, int, time.Time, time.Duration, however both types need to be the same, otherwise the filter will raise *incomparable* as error.


Inline Filter 
--------------

In order to facilitate quick filter definition (without the need of separate FilterProfile), one can define filters directly as FilterIDs following the special format.

Inline filter format::
 
 filterType:fieldName:fieldValue

Example::
 
 *string:WebsiteName:CGRateS.org


Subsystem profiles selection based on Filters
---------------------------------------------

When a subsystem will process an event it will need to find fast enough (close to real-time and most preferably with constant speed) all the profiles having filters matching the event. For low number of profiles (tens of) we can go through all available profiles and check their filters but as soon as the number of profiles is growing, processing time will exponentially grow also. As an example, the *AttributeS* need to deal with 20 mil+ profiles in case of number portability implementation.

In order to guarantee constant processing time - **O(1)** - *CGRateS* will use internally a profile selection mechanism based on indexed filters which can be enabled within *.json* configuration file via *indexed_selects*. When *indexed_selects* is disabled, the indexes will not be used at all and profiles will be checked one by one. On  the other hand, if *indexed_selects* is enabled, each FilterProfile needs to have at least one *\*string* or *\*prefix* type in order to be visible to the indexes (otherwise being completely ignored).

The following settings are further applied once *indexed_selects* is enabled:

string_indexed_fields
	list of field names in the event which will be checked against string indexes (defaults to nil which means check all fields)

prefix_indexed_fields
	list of field names in the event which will be checked against prefix indexes (default is empty, hence prefix matching is disabled inside indexes - small optimization since for prefixes there are multiple queries done for one field)

 
