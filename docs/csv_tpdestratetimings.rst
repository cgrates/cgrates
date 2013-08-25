DestinationRateTimings.csv
++++++++++++++++++++++++++

Enable DestinationRates at specific times.

CSV fields examples as tabular representations:

+-----------------+----------------------+-----------+--------+
| Tag             | DestinationRatesTag  | TimingTag | Weight |
+=================+======================+===========+========+
| RETAIL1         | DR_RETAIL_PEAK       | PEAK      | 10     |
+-----------------+----------------------+-----------+--------+
| RETAIL1         | DR_FREESWITCH_USERS  | ALWAYS    | 10     |
+-----------------+----------------------+-----------+--------+


**Fields**

Index 0 - *Tag*
  Free-text field used to reference the entry from other files.

Index 1 - *DestinationRatesTag*
  References profile in DestinationRates.csv_.

Index 2 - *TimingTag*
  References profile defined in Timings.csv_.

Index 3 - *Weight*
  Solves possible conflicts between different DestinationRateTimings profiles matching on same interval. 
  Higher *Weight* has higher priority.


.. _DestinationRates.csv: csv_tpdestinationrates.html
.. _Timings.csv: csv_tptimings.html

