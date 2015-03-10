Tutorial FS_JSON
================

Scenario:
---------

- FreeSWITCH with minimal custom configuration. 

 - Added following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1006-prepaid, 1007-prepaid.
 - Have added inside default dialplan CGR own extensions just before routing towards users (*etc/freeswitch/dialplan/default.xml*).

- **CGRateS** with following components:

 - CGR-SM started as prepaid controller, with debits taking place at 5s intervals.
 - CGR-Mediator component attaching costs to the raw CDRs from FreeSWITCH_ inside CGR StorDB.
 - CGR-CDRE exporting mediated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-CDRStats component building up stats in 5 different queues.
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).
