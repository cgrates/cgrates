.. _filters:

Filters
=======

Filters define conditions for matching events. Components such as AttributeS use them to determine which profiles match an event. Conditions can compare values, match text or check whether a field is present.

Profiles specify filters in ``FilterIDs``, using IDs for stored definitions or inline expressions for individual rules. Stored filters belong to a tenant and can be shared across profiles. All rules in a filter, and all filters listed in ``FilterIDs``, must match.

Stored filters
--------------

A stored filter is identified by its tenant and ID and contains a list of rules:

.. code-block:: json
   :caption: filter.json

   {
     "Tenant": "cgrates.org",
     "ID": "FLTR_ACCOUNTS",
     "Rules": [
       {
         "Type": "*string",
         "Element": "~*req.Account",
         "Values": ["1001", "1002"]
       }
     ]
   }

This filter matches events with ``Account`` set to ``1001`` or ``1002``, but not ``1003``.

``Type`` and ``Element`` are strings. ``Values`` is an array of strings, including for numeric and time comparisons.

* ``Type`` specifies the comparison or check.
* ``Element`` specifies what to check, here the ``Account`` field referenced by ``~*req.Account``.
* ``Values`` supplies the values or parameters required by ``Type``. For ``*string``, the rule matches if the element equals any listed value.

Stored filters can be loaded or managed through the APIs. An attribute profile refers to stored filters through ``FilterIDs``, for example ``"FilterIDs": ["FLTR_ACCOUNTS"]``. The profile and the filters it references must belong to the same tenant.

Filter rules
------------

Field references
~~~~~~~~~~~~~~~~

In the event APIs used below, ``*req`` contains the fields in ``Event`` and ``*opts`` contains those in ``APIOpts``. Thus ``~*req.Account`` reads ``Event.Account``.

Unlike the literal value ``1001``, ``~*req.Account`` is resolved from the request when the filter is evaluated. Both ``Element`` and ``Values`` in a ``*string`` rule can use field references. For example, ``Element: "~*req.Account"`` with ``Values: ["~*req.Subject"]`` compares the two request fields.

Combining conditions
~~~~~~~~~~~~~~~~~~~~

All rules in a filter must match. Combining the account rule above with this rule also requires a destination starting with ``20``:

.. code-block:: json

   {
     "Type": "*prefix",
     "Element": "~*req.Destination",
     "Values": ["20"]
   }

An event with ``Account: "1001"`` and ``Destination: "2001"`` matches, but changing the destination to ``3001`` makes it fail.

All filters listed in ``FilterIDs`` must also match. An empty ``FilterIDs`` list matches every event. An unknown stored filter ID returns an error.

To match events with ``Account`` set to ``1001`` **or** ``1002``, keep both values in one ``*string`` rule. Two separate account rules would require the account to equal both values. Rules with identical ``Type`` and ``Element`` remain separate conditions that must all match.

Missing fields and negative checks
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

When ``Element`` references a missing field, a ``*string`` rule does not match. Field references in ``Values`` are skipped when the referenced field is missing. ``*notstring`` reverses the match result: it passes when none of the values match, including when the checked field is missing. To require a field to exist as well, add a ``*exists`` rule.

Other negative types work the same way: ``*prefix`` becomes ``*notprefix``, ``*eq`` becomes ``*noteq``, and so on. If evaluation fails, negative checks return the error.

``*exists`` checks that a field is present. ``*empty`` matches missing fields, null values, empty strings or collections, and zero values such as ``0`` and ``false``. Neither type needs ``Values``.

Inline filters
--------------

An inline filter defines a single rule directly in ``FilterIDs``. Its syntax is:

.. code-block:: text

   Type:Element:Value1|Value2

For example, the account rule above can be written as:

.. code-block:: json

   {
     "FilterIDs": ["*string:~*req.Account:1001|1002"]
   }

To require the destination prefix too, add another inline filter:

.. code-block:: json

   {
     "FilterIDs": [
       "*string:~*req.Account:1001|1002",
       "*prefix:~*req.Destination:20"
     ]
   }

Inline filters require the final colon, including when ``Values`` is unused: ``*exists:~*req.Account:``. IDs beginning with ``*`` are interpreted as inline filters, so do not use that prefix for stored filter IDs.

Use stored filters when several records should share a definition that can be changed in one place.

In inline syntax, ``|`` separates values. Use a stored ``Values`` array when a value needs a literal pipe, such as a regular expression containing ``|``. Backslash is not a general escape for inline separators.

Filter types
------------

For positive text and comparison checks, matching any entry in ``Values`` is enough. For ``*ai``, the entries define time bounds.

.. list-table::
   :header-rows: 1
   :widths: 20 20 60

   * - Type
     - Negative form
     - Check
   * - ``*string``
     - ``*notstring``
     - Exact text equality.
   * - ``*contains``
     - ``*notcontains``
     - The text contains one of the values.
   * - ``*prefix``
     - ``*notprefix``
     - The text starts with one of the values.
   * - ``*suffix``
     - ``*notsuffix``
     - The text ends with one of the values.
   * - ``*regex``
     - ``*notregex``
     - A Go regular expression matches the text. Patterns are used directly, not resolved as field references. Use ``^`` and ``$`` to match the whole text rather than a substring.
   * - ``*eq``
     - ``*noteq``
     - Converts the operands to numbers, times or durations where possible, then checks equality. Strings are compared as text. Booleans are not supported. Use ``*string`` to compare boolean values as text.
   * - ``*lt``, ``*lte``, ``*gt``, ``*gte``
     - None
     - Less than, less than or equal, greater than, or greater than or equal. These compare numbers, times or durations, not arbitrary text.
   * - ``*exists``, ``*empty``
     - ``*notexists``, ``*notempty``
     - Presence and emptiness checks. See `Missing fields and negative checks`_.
   * - ``*ipnet``
     - ``*notipnet``
     - An IP address belongs to one of the CIDR networks in ``Values``.
   * - ``*ai``
     - ``*notai``
     - With two bounds, the time must fall strictly between them. With only the first bound, it must be later than that bound. Leave the first value empty to require a time strictly before the second bound.
   * - ``*cronexp``
     - ``*notcronexp``
     - A time matches one of the cron expressions, at minute precision.
   * - ``*rsr``
     - ``*notrsr``
     - A literal such as ``1001`` matches anywhere in the element. Use ``^10`` for a prefix, ``01$`` for a suffix, and ``^1001$`` for an exact match. A leading ``~`` introduces a regular expression. A leading ``!`` negates an expression. Expressions also support ``>``, ``>=``, ``<`` and ``<=`` comparisons.
   * - ``*never``
     - None
     - Does not match. No element or values are needed.
   * - ``*apiban``
     - ``*notapiban``
     - Checks whether an IP address is banned by APIBan. See `APIBan filters`_.
   * - ``*sentrypeer``
     - ``*notsentrypeer``
     - Checks whether an IP address or number is reported by SentryPeer. See `SentryPeer filters`_.
   * - ``*http#[URL]``
     - None
     - Uses an HTTP response body as the match result. See `HTTP filters`_.

APIBan filters
~~~~~~~~~~~~~~

``*apiban`` requires the corresponding service configuration. Its first value in ``Values`` is ``*single`` to query the IP address or ``*all`` to obtain the banned list.

SentryPeer filters
~~~~~~~~~~~~~~~~~~

``*sentrypeer`` requires the corresponding service configuration. Its first value in ``Values`` is ``*ip`` or ``*number``.

HTTP filters
~~~~~~~~~~~~

``*http#[URL]`` converts the HTTP response body to a boolean match result. It does not use ``Values``, although the filter APIs still require a nonempty values array.

Managing stored filters
-----------------------

The AdminS APIs manage filter definitions and update their indexes and caches according to the configured cache policy. Inline filters are changed by editing the record containing them.

``AdminSv1.SetFilter``
    Creates a filter if it does not exist. When updating, it replaces all stored rules that have the same ``Type`` and ``Element`` as a submitted rule. Rules with other ``Type`` and ``Element`` combinations remain.

    For example, submitting a ``*string`` rule for ``~*req.Account`` replaces the account rule's values rather than adding to them. A stored ``*prefix`` rule for ``~*req.Destination`` would remain. Submitted duplicate rules are kept as separate rules.

    ``Overwrite: true`` instead replaces the whole filter: rules absent from the request are removed. An empty ``Rules`` list is rejected in either mode.

``AdminSv1.GetFilter``
    Reads the current definition by tenant and ID.

``AdminSv1.SetFilterRule``
    Adds missing values to the rule selected by ``Type`` and ``Element``. If no rule matches, it creates that rule inside the existing filter. The filter must already exist.

    There is no ``Overwrite`` option. To replace values instead of adding, use ``SetFilter``.

``AdminSv1.RemoveFilterRule``
    Selects a rule by ``Type`` and ``Element``. If ``Values`` is nonempty, it removes those values from the rule. If this removes the rule's last value, the rule is deleted. Omitting ``Values`` or supplying an empty list deletes the selected rule itself.

    Removing the filter's final rule is rejected, leaving the stored filter unchanged. Use ``RemoveFilter`` to delete the whole filter.

``SetFilterRule`` and ``RemoveFilterRule`` return an error if several rules match the requested ``Type`` and ``Element``. Use ``SetFilter`` to replace those matching rules. ``RemoveFilterRule`` also returns an error if no rule matches.

``AdminSv1.RemoveFilter`` deletes the filter by tenant and ID. Records that refer to its ID must be changed separately.

Usage examples
--------------

These examples require a test engine with AdminS enabled. The attribute example also requires AttributeS. ``cgr-console`` connects to ``127.0.0.1:2012`` by default. Use ``cgr-console -server host:port`` to connect to another engine.

Checking stored and inline filters
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

``AdminSv1.FiltersMatch`` checks an event against the stored and inline filters in ``FilterIDs`` and returns whether they all match.

These examples use the ``FLTR_ACCOUNTS`` definition from `Stored filters`_. Save it as ``filter.json`` in the console's working directory. The following commands create the filter and check two events against it:

.. code-block:: console

   $ cgr-console
   cgr> admins.setFilter @filter.json
   "OK"
   cgr> admins.filtersMatch Tenant=cgrates.org FilterIDs=["FLTR_ACCOUNTS"] Event={"Account":"1001"}
   true
   cgr> admins.filtersMatch Tenant=cgrates.org FilterIDs=["FLTR_ACCOUNTS"] Event={"Account":"1003"}
   false

The same check can use an inline filter:

.. code-block:: console

   cgr> admins.filtersMatch Tenant=cgrates.org FilterIDs=["*string:~*req.Account:1001|1002"] Event={"Account":"1001"}
   true

Selecting an attribute profile
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This example requires ``FLTR_ACCOUNTS`` to be stored with the definition in `Stored filters`_. The expected results assume there are no other matching attribute profiles in the tenant. Save this attribute profile as ``attribute.json``:

.. code-block:: json
   :caption: attribute.json

   {
     "Tenant": "cgrates.org",
     "ID": "ATTR_ACCOUNTS",
     "FilterIDs": ["FLTR_ACCOUNTS"],
     "Weights": [
       {
         "Weight": 10
       }
     ],
     "Attributes": [
       {
         "Path": "*req.Category",
         "Type": "*constant",
         "Value": "call"
       }
     ]
   }

The profile defines an attribute that sets ``Category`` to ``call``, but ``GetAttributeForEvent`` only selects the profile without applying that change:

.. code-block:: console

   cgr> admins.setAttributeProfile @attribute.json
   "OK"
   cgr> attributes.getAttributeForEvent Tenant=cgrates.org Event={"Account":"1001"}

The returned profile has ID ``ATTR_ACCOUNTS``. An event with ``Account: "1003"`` does not match this profile:

.. code-block:: console

   cgr> attributes.getAttributeForEvent Tenant=cgrates.org Event={"Account":"1003"}
   error: NOT_FOUND

Adding and removing rule values
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This example requires ``FLTR_ACCOUNTS`` to be stored with the definition in `Stored filters`_. Adding ``1003`` allows events with that account value to match the stored filter. Removing it restores the original condition:

.. code-block:: console

   cgr> admins.setFilterRule Tenant=cgrates.org ID=FLTR_ACCOUNTS Type=*string Element=~*req.Account Values=["1003"]
   "OK"
   cgr> admins.removeFilterRule Tenant=cgrates.org ID=FLTR_ACCOUNTS Type=*string Element=~*req.Account Values=["1003"]
   "OK"

Here, ``RemoveFilterRule`` rejects a request without ``Values`` because it would delete the filter's only rule.

Use ``help admins.setFilter``, ``help admins.setFilterRule`` and ``help admins.removeFilterRule`` in the console to inspect the request fields.
