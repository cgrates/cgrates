Kamailio_ interaction via  *evapi* module
=========================================

Scenario
--------

 - Kamailio default configuration modified for **CGRateS** interaction. For script maintainability and simplicity we have separated CGRateS specific routes in *kamailio-cgrates.cfg* file which is included in main *kamailio.cfg* via include directive.

 - Considering the following users (with configs hardcoded in the *kamailio.cfg* configuration script and loaded in htable): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1005-rated, 1006-prepaid, 1007-prepaid.

- **CGRateS** with following components:

 - CGR-SM started as translator between Kamailio_ and CGR-Rater for both authorization events as well as accounting ones.
 - CGR-CDRS component processing raw CDRs from CGR-SM component and storing them inside CGR StorDB.
 - CGR-CDRE exporting rated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).


Starting Kamailio_ with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/kamevapi/kamailio/etc/init.d/kamailio start

To verify that Kamailio_ is running we run the console command:

::

 kamctl moni


Starting **CGRateS** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/kamevapi/cgrates/etc/init.d/cgrates start

Make sure that cgrates is running

::

 cgr-console status


CDR processing
--------------

At the end of each call Kamailio_ will generate an CDR event via *evapi* and this will be directed towards the port configured inside *cgrates.json*. This event will reach inside **CGRateS** through the *SM* component (close to real-time). Once in-there it will be instantly rated and be ready for export. 


**CGRateS** Usage
-----------------

Since it is common to most of the tutorials, the example for **CGRateS** usage is provided in a separate page `here <http://cgrates.readthedocs.org/en/latest/tut_cgrates_usage.html>`_


.. _Kamailio: http://www.kamailio.org/
