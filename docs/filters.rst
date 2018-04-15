FilterS 
=======

**FilterS** are code blocks applied to generic events (hashmaps) in order to allow/deny further processing.

A Tenant will define multiple Filter profiles via .csv or API calls. The Filter profile ID is unique within a tenant but it can be repeated over multiple Tenants.

In order to be used in event processing a Filter profile will be attached inside another subsystem profile definition, otherwise Filter profile will have no effect on it's own. 
A Filter profile can be shared between multiple subsystem profile definitions.


Filter profile 
--------------

Definition::

 type Filter struct {
	Tenant             string
	ID                 string
	Rules              []*FilterRule
	ActivationInterval *utils.ActivationInterval
 }


A Filter profile can contain any number of Filter rules and all of them must pass in order for the filter profile to pass.

A Filter profile can be activated on specific interval, if multiple filters are used within a subsystem profile at least one needs to be active and passing in order for the subsystem profile to pass the event.


Filter rule 
-----------

Definition::

 type FilterRule struct {
	Type            string              // Filter type (*string, *timing, *rsr, *stats, *lt, *lte, *gt, *gte)
	FieldName       string              // Name of the field providing us the Values to check (used in case of some )
	Values          []string            // Filter definition
 }


The matching logic of each FilterRule is given by it's type.

The following types are implemented:

- *\*string* will match in full the *FieldName* with at least one value defined inside *Values*. Any of them matching will match the FilterRule. It is indexed for performance and, in order to be enabled, the subsystem configuration where the Filter profile is used needs to have the parameter *string_indexed_fields* nil or contain the Filter profile ID inside.

- *\*prefix* will match at beginning of *FieldName* one of the values defined inside *Values*. It is indexed for performance and, in order to be enabled, the subsystem configuration where the Filter profile is used needs to have the parameter *prefix_indexed_fields* nil or contain the Filter profile ID inside.

- *\*timings* will compare the time contained in *FieldName* with one of the TimingIDs defined in Values.

- *\*destinations* will make sure that the *FieldName* is a prefix contained inside one of the destination IDs as *Values*.

- *\*rsr* will match the *RSRRules* defined in Values. The field name is taken out of *RSRRule.ID* and matching logic is done against *RSRRule.Filters*

- *\*lt* (less than), *\*lte* (less than or equal), *\*gt* (greather than), *\*gte* (greather than or equal) are comparison operators and they pass if at least one of the values defined in *Values* are passing for the *FieldName* of event. The operators are able to compare string, float, int, time.Time, time.Duration, however both types need to be the same, otherwise the filter will raise *incomparable* as error.