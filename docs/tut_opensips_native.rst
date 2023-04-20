OpenSIPS_ interaction via  own *cgrates* module
===============================================

Scenario
--------

- OpenSIPS out of *residential* configuration generated. 

 - The users are all defined within CGRateS.
 - For simplicity we configure no authentication (WARNING: Not for production usage).

- **CGRateS** with following subsystems:

 - **SM**: (SessionManager) started as gateway between OpenSIPS_ and rest of CGRateS subsystems.
 - **ChargerS**: used to decide the number of billing runs for customer/supplier charging.
 - **AttributeS**: used to populate extra data to requests (ie: prepaid/postpaid, passwords, paypal account, LCR profile).
 - **RALs**: used to calculate costs as well as account bundle management.
 - **SupplierS**: selection of suppliers for each session. This will work in tandem with OpenSIPS_'s DRouting module.
 - **StatS**: computing statistics in real-time regarding sessions and their charging.
 - **ThresholdS**: monitoring and reacting to events coming from above subsystems.
 - **CDRe**: exporting rated CDRs from CGR StorDB (export path: */tmp*).


Creating OpenSIPS_ database for DRouting module
-----------------------------------------------

::

 opensips-cli -x database create


Starting OpenSIPS_ with custom configuration
--------------------------------------------

::

 /usr/share/cgrates/tutorials/osips_native/opensips/etc/init.d/opensips start

To verify that OpenSIPS_ is running we run the console command:

::

 opensipsctl moni


Starting **CGRateS** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/osips_native/cgrates/etc/init.d/cgrates start

Make sure that cgrates is running

::

 cgr-console status


CDR processing
--------------

At the end of each call OpenSIPS_ will generate an CDR event and due to automatic handler registration built in **CGRateS-SM** component, this will be directed towards the port configured inside *cgrates.json*. This event will reach inside **CGRateS** through the *SM* component (close to real-time). Once in-there it will be instantly rated and be ready for export. 


**CGRateS** Usage
-----------------

Since it is common to most of the tutorials, the example for **CGRateS** usage is provided in a separate page `here <http://cgrates.readthedocs.org/en/latest/tut_cgrates_usage.html>`_


.. _OpenSIPS: https://opensips.org/
