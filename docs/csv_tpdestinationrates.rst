DestinationRates.csv
++++++++++++++++++++

Bind destination group from Destinations.csv_ with rates defined in Rates.csv_ files.

CSV fields as tabular representation:

+--------------------+------------------+---------------------+
| Tag                | DestinationsTag  | RatesTag            |
+====================+==================+=====================+
| DR_RETAIL_PEAK     | GERMANY          | LANDLINE_PEAK       |
+--------------------+------------------+---------------------+
| DR_RETAIL_OFFPEAK  | GERMANY          | LANDLINE_OFFPEAK    |
+--------------------+------------------+---------------------+

Index 0 - *Tag*
    Free-text field used to reference the entry from other files.

Index 1 - *Years*
    Years this timing is valid on.

    Possibile values:
     * Semicolon (;) separated list of years as descriptive filter.
     * "\*any" metatag used as match-any filter.

Index 2 - *Months*
    Months this timing is valid on.

    Possibile values:
     * Semicolon (;) separated list of months as descriptive filter.
     * "\*any" metatag used as match-any filter.

Index 3 - *MonthDays*
    Days of a month this timing is valid on.

    Possibile values:
     * Semicolon (;) separated list of month days as descriptive filter.
     * "\*any" metatag used as match-any filter.

Index 4 - *WeekDays*
    Days of a week this timing is valid on. Week days represented as integers where 1=Monday and 7=Sunday

    Possibile values:
     * Semicolon (;) separated list of week days as descriptive filter.
     * "\*any" metatag used as match-any filter.

Index 5 - *Time*
    The start time for this time period.

    Possibile values:
     * String representation of time (hh:mm:ss).
     * "\*asap" metatag used to represent time converted at runtime.


.. _Destinations.csv: csv_tpdestinations
.. _Rates.csv: csv_tprates



