Data importing
=============

For importing the data into CGRateS database we are using cvs files. The import process can be started as many times it is desired with one ore more csv files and the existing values are overwritten. If the -flush option is used then the database is cleaned before importing.For more details see the cgr-loader tool from the tutorial chapter.

The rest of this section we will describe the content of every csv files.

Rates profile
--------------

The rates profile describes the prices to be applied for various calls to various destinations in various time frames. When a call is made the CGRateS system will locate the rates to be applied to the call using the rating profiles.

+------------+-----+-----------+-------------+----------------------+----------------+----------------------+
| Tenant     | TOR | Direction | Subject     | RatesFallbackSubject | RatesTimingTag | ActivationTime       |
+============+=====+===========+=============+======================+================+======================+
| CUSTOMER_1 | 0   | OUT       | rif:from:tm | danb                 | PREMIUM        | 2012-01-01T00:00:00Z |
+------------+-----+-----------+-------------+----------------------+----------------+----------------------+
| CUSTOMER_1 | 0   | OUT       | rif:from:tm | danb                 | STANDARD       | 2012-02-28T00:00:00Z |
+------------+-----+-----------+-------------+----------------------+----------------+----------------------+

+ Tenant
    Used to distinguish between carriers if more than one share the same database in the CGRates system.
+ TOR
    Type of record specifies the kind of transmission this rate profile applies to.
+ Direction
    Can be IN or OUT for the INBOUND and OUTBOUND calls.
+ Subject
    The client/user for who this profile is detailing the rates.
+ RatesFallbackSubject
    This specifies another profile to be used in case the call destination will not be found in the current profile. The same tenant, tor and direction will be used.
+ RatesTimingTag
    Forwards to a tag described in the rates timing file to be used for this profile.
+ ActivationTime
    Multiple rates timings/prices can be created for one profile with different activation times. When a call is made the appropriate profile(s) will be used to rate the call.

Rates timings
-------------

This file makes links between a ratings and timings so each of them can be described once and various combinations are made possible.

+----------+----------------+--------------+
| Tag      | RatesTag       | TimingTag    |
+==========+================+==============+
| STANDARD | RT_STANDARD    | WORKDAYS_00  |
+----------+----------------+--------------+
| STANDARD | RT_STD_WEEKEND |  WORKDAYS_18 |
+----------+----------------+--------------+

+ Tag
    A string by witch this rates timing will be referenced in other places by.
+ RatesTag
    The rating tag described in the rates file.
+ TimingTag
    The timing tag described in the timing file

Rates
-----

+---------------------+-----------------+------------+-------+-------------+
| Tag                 | DestinationsTag | ConnectFee | Price | BillingUnit |
+=====================+=================+============+=======+=============+
| RT_STANDARD         | GERMANY         | 0          | 0.2   | 1           |
+---------------------+-----------------+------------+-------+-------------+
| RT_STANDARD         | GERMANY_O2      | 0          | 0.1   | 1           |
+---------------------+-----------------+------------+-------+-------------+


+ Tag
    A string by witch this rate will be referenced in other places by.
+ DestinationsTag
    The destination tag witch these rates apply to.
+ ConnectFee
    The price to be charged once at the beginning of the call to the specified destination.
+ Price
    The price for the billing unit expressed in cents.    
+ BillingUnit
    The billing unit expressed in seconds

Timings
-------

+-------------+--------+-----------+-----------+----------+--------+
| Tag         | Months | MonthDays |  WeekDays | StartTime| Weight |
+=============+========+===========+===========+==========+========+
| WORKDAYS_00 | *all   | *all      | 1;2;3;4;5 | 00:00:00 | 10     |
+-------------+--------+-----------+-----------+----------+--------+
| WORKDAYS_18 | *all   | *all      | 1;2;3;4;5 | 18:00:00 | 10     |
+-------------+--------+-----------+-----------+----------+--------+

+ Tag
    A string by witch this timing will be referenced in other places by.
+ Months
+ MonthDays
+ WeekDays
+ StartTime
+ Weight
    If multiple timings cab be applied to a call the one with the lower weight wins.

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
