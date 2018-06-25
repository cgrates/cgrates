FreeSWITCH_ generating *http-json* CDRs
=======================================

Scenario
--------

- FreeSWITCH with *vanilla* configuration adding *mod_json_cdr* for CDR generation. 

 - Modified following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1006-prepaid, 1007-rated.
 - Have added inside default dialplan CGR own extensions just before routing towards users (*etc/freeswitch/dialplan/default.xml*).
 - FreeSWITCH configured to generate default *http-json* CDRs.

- **CGRateS** with following components:

 - CGR-SM started as prepaid controller, with debits taking place at 5s intervals.
 - CGR-CDRS component receiving raw CDRs from FreeSWITCH, storing them and attaching costs inside CGR StorDB.
 - CGR-CDRE exporting processed CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).


Starting FreeSWITCH_ with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_evsock/freeswitch/etc/init.d/freeswitch start

To verify that FreeSWITCH_ is running we run the console command:

::

 fs_cli -x status


Starting **CGRateS** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_evsock/cgrates/etc/init.d/cgrates start

Check that cgrates is running

::

 cgr-console status


CDR processing
--------------

At the end of each call FreeSWITCH_ will issue a http post with the CDR. This will reach inside **CGRateS** through the *CDRS* component (close to real-time). Once in-there it will be instantly rated and it is ready to be exported: 

::

 cgr-console 'cdrs_export CdrFormat="csv" ExportPath="/tmp"'


**CGRateS** Usage
-----------------

Since it is common to most of the tutorials, the example for **CGRateS** usage is provided in a separate page `here <http://cgrates.readthedocs.org/en/latest/tut_cgrates_usage.html>`_


.. _FreeSWITCH: http://www.freeswitch.org/
.. _Jitsi: http://www.jitsi.org/
