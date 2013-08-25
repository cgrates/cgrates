RatingProfiles.csv
++++++++++++++++++

Main definitions file for the Rating subsystem.

+------------------+------+-----------+---------+-----------------------+----------------------------+----------------------+
| Tenant           | TOR  | Direction | Subject | ActivationTime        | DestinationRateTimingTag   | RatesFallbackSubject |
+==================+======+===========+=========+=======================+============================+======================+
| cgrates.org      | call | \*out     | \*any   | 2012-01-01T00:00:00Z  | RETAIL1                    |                      |
+------------------+------+-----------+---------+-----------------------+----------------------------+----------------------+

**Fields**

Index 0 - *Tenant*
  Free-text field used to identify the tenant the entries are valid for.
Index 1 - *TOR*
  Free-text field used to identify the type of record the entries are valid for.
Index 2 - *Direction*
  *Metatag* identifying the traffic direction the entries are valid for. Outbound direction is the only one supported for now.
Index 3 - *Subject*
  Rating subject definition.

  Possible values:
   * Free-text rating subject (flexible defition for example out of concatenating various cdr fields)
   * "\*any" metatag matching any rating subject in the eventuality of no explicit subject string matching.
Index 4 - *ActivationTime*
  Time this rating profile gets active at.

  Possible values:
   * RFC3339 time as string
   * Unix timestamp
   * String starting with "+" to represent duration dynamically calculated at runtime (eg: +1h to specify ActivationTime one hour after runtime).
   * "\*monthly" metatag for ActivationTime dynamically calculated one month after runtime.
Index 4 - *DestinationRateTimingTag*
  References profile in DestinationRateTimings.csv_.
Index 5 - *RatesFallbackSubject*
  Name of the fallback subject to be considered if existing subject has no destination matching the one searched. *Tenant*, *TOR*, *Direction*, *Subject* are kept when matching the fallback profile.

.. _DestinationRateTimings.csv: csv_tpdestinationrates.html
