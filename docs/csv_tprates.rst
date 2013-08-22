Rates.csv
+++++++++

Defines the rates on the system. 
Each entry is a part of a rate group, each group having at least one entry. Group entries share *Tag*, *ConnectFee*, *Weight* parameters.

CSV fields example as tabular representation:

+---------------------+------------+------+----------+---------------+--------------------+----------------+------------------+---------+
| Tag                 | ConnectFee | Rate | RateUnit | RateIncrement | GroupIntervalStart | RoundingMethod | RoundingDecimals | Weight  |
+=====================+============+======+==========+===============+====================+================+==================+=========+
| LANDLINE_PEAK       | 0.02       | 0.02 | 60s      | 60s           | 0s                 | \*up           | 4                | 10      |
+---------------------+------------+------+----------+---------------+--------------------+----------------+------------------+---------+
| MOBILE_PEAK         | 1          | 2    | 60s      | 10s           | 0s                 | \*middle       | 4                | 10      |
+---------------------+------------+------+----------+---------------+--------------------+----------------+------------------+---------+
| MOBILE_PEAK         | 1          | 1    | 60s      | 20s           | 40s                | \*middle       | 4                | 10      |
+---------------------+------------+------+----------+---------------+--------------------+----------------+------------------+---------+
| MOBILE_PEAK         | 1          | 0    | 60s      | 10s           | 60s                | \*middle       | 4                | 10      |
+---------------------+------------+------+----------+---------------+--------------------+----------------+------------------+---------+



Index 0 - *Tag*
  Free-text field used to reference the entry from other files.

Index 1 - *ConnectFee*
  Connect fee charged at start of each call. Should be the same for all members of a group interval.

  Possible values:
   * Float or integer value, granularity given by rates administrator and not predefined (eg: cent vs euro).

Index 2 - *Rate*
  The rate which will be charged.

  Possible values:
   * Float or integer value, granularity given by rates administrator and not predefined (eg: cent vs euro).

Index 3 - *RateUnit*
  The duration unit which is rated by *Rate* field.

  Possible values:
   * Duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

Index 4 - *RateIncrement*
  The total duration will be split and rounded into smaller intervals based on this (eg: for *RateIncrement*  of 60s, total duration of 1m2s will be charged as 2 minutes).

  Possible values:
   * Duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

Index 5 - *GroupIntervalStart*
  The position in the rate group. 

  Possible values:
   * Duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

Index 6 - *RoundingMethod*
  The routine which will round the cost on each timespan.

  Possible values:
   * A *MetaTag* referring the internal routine doing the rounding (eg: \*up, \*down, \*middle)

Index 7 - *RoundingDecimals*
  Round the number of decimals of each timespan based on this setting.

  Possible values:
   * An integer value.

Index 8 - *Weight*
  Solve possible conflicts between different rates matching same interval based on this parameter. 
  Higher *Weight* has bigger priority.
  Should be the same for all members of a group interval.

  Possible values:
   * Float/Integer representing the rate weight on collisions.
  


