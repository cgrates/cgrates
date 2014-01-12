Tutorial fs_prepaid_csv
=======================

Scenario:
---------

* FreeSWITCH with default configuration. 

 * Modified following users: 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated.
 * Have added inside default dialplan CGR own extensions just before routing towards users.
 * FreeSWITCH configured to generate default .csv CDRs, modified example template to add cgr_reqtype from user variables.

* CGRateS with following components:

 * CGR-SM started as prepaid controller.
 * CGR-CDRC component importing CDRs into CGR.
 * CGR-CDRE exporting mediated CDRs

