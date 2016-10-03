Asterisk_ interaction via  *ARI*
===========================================

Scenario
--------

- Asterisk out of *basic-pbx* configuration samples. 

 - Considering the following users: 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1007-rated.

- **CGRateS** with following components:

 - CGR-SM started as translator between Asterisk_ and **CGR-RALs** for both authorization events (prepaid/pseudoprepaid) as well as postpaid ones.
 - CGR-CDRS component processing raw CDRs from CGR-SM component and storing them inside CGR StorDB.
 - CGR-CDRE exporting rated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).


Starting Asterisk_ with custom configuration
----------------------------------------------

::

 asterisk -r -s /tmp/cgr_asterisk_ari/asterisk/run/asterisk.ctl

To verify that Asterisk_ is running we run the console command:

::

 ari show status


Starting **CGRateS** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/asterisk_ari/cgrates/etc/init.d/cgrates start

Make sure that cgrates is running

::

 cgr-console status


CDR processing
--------------

At the end of each call Asterisk_ will generate an CDR event and due to automatic handler registration built in **CGRateS-SM** component, this will be directed towards the port configured inside *cgrates.json*. This event will reach inside **CGRateS** through the *SM* component (close to real-time). Once in-there it will be instantly rated and be ready for export. 


**CGRateS** Usage
-----------------

Since it is common to most of the tutorials, the example for **CGRateS** usage is provided in a separate page `here <http://cgrates.readthedocs.org/en/latest/tut_cgrates_usage.html>`_


.. _Asterisk: http://www.asterisk.org/
