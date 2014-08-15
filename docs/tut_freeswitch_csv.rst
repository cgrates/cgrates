FreeSWITCH_ generating *.csv* CDRs
==================================

Scenario
--------

- FreeSWITCH with *vanilla* configuration, minimal modifications to fit our needs. 

 - Modified following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated.
 - Have added inside default dialplan CGR own extensions just before routing towards users (*etc/freeswitch/dialplan/default.xml*).
 - FreeSWITCH configured to generate default *.csv* CDRs, modified example template to add cgr_reqtype from user variables (*etc/freeswitch/autoload_configs/cdr_csv.conf.xml*).

- **CGRateS** with following components:

 - CGR-SM started as prepaid controller, with debits taking place at 5s intervals.
 - CGR-CDRC component importing FreeSWITCH_ generated *.csv* CDRs into CGR and moving the processed *.csv* files to */tmp* folder.
 - CGR-Mediator compoenent attaching costs to the raw CDRs from CGR-CDRC inside CGR StorDB.
 - CGR-CDRE exporting mediated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).


Starting FreeSWITCH_ with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch start

To verify that FreeSWITCH_ is running we run the console command:

::

 fs_cli -x status


Starting **CGRateS** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_csv/cgrates/etc/init.d/cgrates start

Check that cgrates is running

::

 cgr-console status

CDR processing
--------------

For every call FreeSWITCH_ will generate CDR records within the *Master.csv* file. 
In order to avoid double-processing we will use the rotate mechanism built in FreeSWITCH_. 
Once rotated, we will move the resulted files inside the path considered by **CGRateS** *CDRC* component as inbound.

These steps are automated in a script provided in the */usr/share/cgrates/scripts* location:

::

 /usr/share/cgrates/scripts/freeswitch_cdr_csv_rotate.sh


On each rotate CGR-CDRC component will be informed via *inotify* subsystem and will instantly process the CDR file. The records end up in **CGRateS**/StorDB inside *cdrs_primary* table via CGR-CDRS. As soon as the CDR will hit CDRS component, mediation will occur, either considering the costs calculated in case of prepaid and postpaid calls out of *cost_details* table or query it's own one from rater in case of *pseudoprepaid* and *rated* CDRs.

Once the CDRs are mediated, can be exported as *.csv* format again via remote command offered by *cgr-console* tool:

::

 cgr-console 'cdrs_export CdrFormat="csv" ExportDir="/tmp"'


**CGRateS** Usage
-------------
Since it is common to most of the tutorials, the example for **CGRateS** usage is provided in a separate page `here <http://cgrates.readthedocs.org/en/latest/tut_cgrates_usage.html>`_


.. _FreeSWITCH: http://www.freeswitch.org/
.. _Jitsi: http://www.jitsi.org/
