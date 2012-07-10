Data importing
=============

Rates profile
--------------

+------------+-----+-----------+-------------+----------------------+----------------+----------------------+
| Tenant     | TOR | Direction | Subject     | RatesFallbackSubject | RatesTimingTag | ActivationTime       |
+============+=====+===========+=============+======================+================+======================+
| CUSTOMER_1 | 0   | OUT       | rif:from:tm | danb                 | PREMIUM        | 2012-01-01T00:00:00Z |
+------------+-----+-----------+-------------+----------------------+----------------+----------------------+
| CUSTOMER_1 | 0   | OUT       | rif:from:tm | danb                 | STANDARD       | 2012-02-28T00:00:00Z |
+------------+-----+-----------+-------------+----------------------+----------------+----------------------+

+ Tenant
    Ceva text explicativ.
+ TOR
+ Direction
+ Subject
+ RatesFallbackSubject
+ RatesTimingTag
+ ActivationTime

Rates timings
-------------

+----------+----------------+--------------+
| Tag      | RatesTag       | TimingTag    |
+==========+================+==============+
| STANDARD | RT_STANDARD    | WORKDAYS_00  |
+----------+----------------+--------------+
| STANDARD | RT_STD_WEEKEND |  WORKDAYS_18 |
+----------+----------------+--------------+

+ Tag
    A string by witch this timing will be referenced in other places by.
+ RatesTag
+ TimingTag

Rates
-----

+---------------------+-----------------+------------+-------+-------------+--------+
| Tag                 | DestinationsTag | ConnectFee | Price | BillingUnit | Weight |
+=====================+=================+============+=======+=============+========+
| RT_STANDARD         | GERMANY         | 0          | 0.2   | 1           | 10     |
+---------------------+-----------------+------------+-------+-------------+--------+
| RT_STANDARD         | GERMANY_O2      | 0          | 0.1   | 1           | 10     |
+---------------------+-----------------+------------+-------+-------------+--------+


+ Tag
    A string by witch this rate will be referenced in other places by.
+ DestinationsTag
+ ConnectFee
+ Price
+ BillingUnit
+ Weight

Destinations
------------

The destinations are binding together various prefixes / caller ids to define a logical destination group. A prefix can appear in multiple destination groups.

+------------+-------+
| Tag        | Prefix|
+============+=======+
| GERMANY    | 49    |
+------------+-------+
| GERMANY_O2 | 49176 |
+------------+-------+

+ Tag
    A string by witch this destination will be referenced in other places by.
+ Prefix
    The prefix or caller id to be added to the specified destination.

Timings
-------

+-------------+--------+-----------+-----------+----------+
| Tag         | Months | MonthDays |  WeekDays | StartTime|
+=============+========+===========+===========+==========+
| WORKDAYS_00 | *all   | *all      | 1;2;3;4;5 | 00:00:00 |
+-------------+--------+-----------+-----------+----------+
| WORKDAYS_18 | *all   | *all      | 1;2;3;4;5 | 18:00:00 |
+-------------+--------+-----------+-----------+----------+

+ Tag
    A string by witch this timing will be referenced in other places by.
+ Months
+ MonthDays
+ WeekDays
+ StartTime

Account actions
---------------

+------------+---------+-----------+------------------+------------------+
|Tenant      | Account | Direction | ActionTimingsTag | ActionTriggersTag|
+============+=========+===========+==================+==================+
| CUSTOMER_1 | rif     | OUT       | STANDARD_ABO     | STANDARD_TRIGGER |
+------------+---------+-----------+------------------+------------------+
| CUSTOMER_1 | dan     | OUT       | STANDARD_ABO     | STANDARD_TRIGGER |
+------------+---------+-----------+------------------+------------------+

+ Tenant
+ Account
+ Direction 
+ ActionTimingsTag
+ ActionTriggersTag

Action triggers
---------------

+------------------+------------+----------------+----------------+------------+--------+
| Tag              | BalanceTag | ThresholdValue | DestinationTag | ActionsTag | Weight |
+==================+============+================+================+============+========+
| STANDARD_TRIGGER | MONETARY   | 30             | *all           | SOME_1     | 10     |
+------------------+------------+----------------+----------------+------------+--------+
| STANDARD_TRIGGER | SMS        | 30             | *all           |SOME_2      | 10     |
+------------------+------------+----------------+----------------+------------+--------+

+ Tag
    A string by witch this action trigger will be referenced in other places by.
+ BalanceTag
+ ThresholdValue
+ DestinationTag
+ ActionsTag 
+ Weight

Action timings
--------------

+--------------+------------+------------------+--------+
| Tag          | ActionsTag | TimingTag        | Weight |
+==============+============+==================+========+
| STANDARD_ABO | SOME       | WEEKLY_SAME_TIME | 10     |
+--------------+------------+------------------+--------+
| STANDARD_ABO | SOME       | WEEKLY_SAME_TIME | 10     |
+--------------+------------+------------------+--------+

+ Tag
    A string by witch this action timing will be referenced in other places by.
+ ActionsTag 
+ TimingTag
+ Weight

Actions
-------

+--------+-------------+------------+-------+----------------+-----------+------------+---------------+--------+
| Tag    | Action      | BalanceTag | Units | DestinationTag | PriceType | PriceValue | MinutesWeight | Weight |
+========+=============+============+=======+================+===========+============+===============+========+
| SOME   | TOPUP_RESET | MONETARY   | 10    | *all           |                                        | 10     |
+--------+-------------+------------+-------+----------------+-----------+------------+---------------+--------+
| SOME_1 | DEBIT       | MINUTES    | 10    | GERMANY_O2     | PERCENT   | 25         | 10            | 10     |
+--------+-------------+------------+-------+----------------+-----------+------------+---------------+--------+

+ Tag
    A string by witch this action will be referenced in other places by.
+ Action
+ BalanceTag
+ Units
+ DestinationTag
+ PriceType
+ PriceValue
+ MinutesWeight
+ Weight
