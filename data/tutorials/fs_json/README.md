Tutorial fs_prepaid_csv
=======================

Scenario:
---------

- FreeSWITCH with *vanilla* configuration, replacing *mod_cdr_csv* with *mod_json_cdr*. 

 - Modified following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated.
 - Have added inside default dialplan CGR own extensions just before routing towards users (*etc/freeswitch/dialplan/default.xml*).
 - FreeSWITCH configured to generate default *http-json* CDRs.

- **CGRateS** with following components:

 - CGR-SM started as prepaid controller, with debits taking place at 5s intervals.
 - CGR-Mediator compoenent attaching costs to the raw CDRs from FreeSWITCH_ inside CGR StorDB.
 - CGR-CDRE exporting mediated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).
