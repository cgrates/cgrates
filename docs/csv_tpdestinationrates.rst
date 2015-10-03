DestinationRates.csv
++++++++++++++++++++

Attach rates to destinations.

CSV fields as tabular representation:

+--------------------+------------------+---------------------+
| Tag                | DestinationsTag  | RatesTag            |
+====================+==================+=====================+
| DR_RETAIL_PEAK     | GERMANY          | LANDLINE_PEAK       |
+--------------------+------------------+---------------------+
| DR_RETAIL_OFFPEAK  | GERMANY          | LANDLINE_OFFPEAK    |
+--------------------+------------------+---------------------+

**Fields**

Index 0 - *Tag*
  Free-text field used to reference the entry from other files.

Index 1 - *DestinationsTag*
  References profile in Destinations.csv_.

Index 2 - *RatesTag*
  References profile defined in Rates.csv_.


.. _Destinations.csv: csv_tpdestinations.rst
.. _Rates.csv: csv_tprates.rst



