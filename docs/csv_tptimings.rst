mings.csv
+++++++++++

Holds time related definitions.

CSV fields examples as tabular representations:

+-----------------+--------+--------+-----------+-----------+----------+
| Tag             | Years  | Months | MonthDays |  WeekDays | Time     |
+=================+========+========+===========+===========+==========+
| WORKDAYS        | \*any  | \*any  | \*any     | 1;2;3;4;5 | 00:00:00 |
+-----------------+--------+--------+-----------+-----------+----------+
| WEEKENDS        | \*any  | \*any  | \*any     | 6;7       | 00:00:00 |
+-----------------+--------+--------+-----------+-----------+----------+
| ALWAYS          | \*any  | \*any  | \*any     | \*any     | 00:00:00 |
+-----------------+--------+--------+-----------+-----------+----------+
| ASAP            | \*any  | \*all  | \*all     | \*all     | \*asap   |
+-----------------+--------+--------+-----------+-----------+----------+

**Fields**

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

Index 4 - *Time*
  The start time for this time period.

  Possible values:
   * String representation of time (hh:mm:ss).
   * "\*asap" metatag used to represent time converted at runtime.


