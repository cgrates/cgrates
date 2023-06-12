Tutorial
========

.. contents::
   :local:
   :depth: 3

Introduction
------------

This tutorial provides detailed instructions for setting up a SIP Server and managing communication between the server and the CGRateS instance.

.. note::

   The development and testing of the instructions in this tutorial has been done on a Debian 11 (Bullseye) virtual machine.


Scenario Overview
-----------------

The tutorial comprises the following steps:

1. **SIP Server Setup**:
   Select and install a SIP Server. The tutorial supports the following options:

   -  FreeSWITCH_
   -  Asterisk_
   -  Kamailio_
   -  OpenSIPS_

2. **CGRateS Initialization**:
   Launch a CGRateS instance with the corresponding agent configured. In this context, an "agent" refers to a component within CGRateS that manages communication between CGRateS and the SIP Servers.

3. **Account Configuration**:
   Establish user accounts for different request types.

4. **Balance Addition**:
   Allocate suitable balances to the user accounts.

5. **Call Simulation**:
   Use Zoiper_ (or any other SIP UA of your choice) to register the user accounts and simulate calls between the configured accounts, and then verify the balance updates post-calls.

6. **Fraud Detection Setup**:
   Implement a fraud detection mechanism to secure and maintain the integrity of the service.

As we progress through the tutorial, each step will be elaborated in detail. Let's embark on this journey with the SIP Server Setup.



Software Installation
---------------------

*CGRateS* already has a section within this documentation regarding installation. It can be found :ref:`here<installation>`.

Regarding the SIP Servers, click on the tab corresponding to the choice you made and follow the steps in order to set up:

.. tabs::

   .. group-tab:: FreeSWITCH

      For detailed information on installing FreeSWITCH_ on Debian, please refer to its official `installation wiki <https://developer.signalwire.com/freeswitch/FreeSWITCH-Explained/Installation/Linux/Debian_67240088/>`_.

      Before installing FreeSWITCH_, you need to authenticate by creating a SignalWire Personal Access Token. To generate your personal token, follow the instructions in the `SignalWire official wiki on creating a personal token <https://developer.signalwire.com/freeswitch/freeswitch-explained/installation/howto-create-a-signalwire-personal-access-token_67240087/>`_.

      To install FreeSWITCH_ and configure it, we have chosen the simplest method using *vanilla* packages.

      .. code-block:: bash

         TOKEN=YOURSIGNALWIRETOKEN # Insert your SignalWire Personal Access Token here
         apt-get update && apt-get install -y gnupg2 wget lsb-release
         wget --http-user=signalwire --http-password=$TOKEN -O /usr/share/keyrings/signalwire-freeswitch-repo.gpg https://freeswitch.signalwire.com/repo/deb/debian-release/signalwire-freeswitch-repo.gpg
         echo "machine freeswitch.signalwire.com login signalwire password $TOKEN" > /etc/apt/auth.conf
         chmod 600 /etc/apt/auth.conf
         echo "deb [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ `lsb_release -sc` main" > /etc/apt/sources.list.d/freeswitch.list
         echo "deb-src [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ `lsb_release -sc` main" >> /etc/apt/sources.list.d/freeswitch.list

         # If /etc/freeswitch does not exist, the standard vanilla configuration is deployed
         apt-get update && apt-get install -y freeswitch-meta-all

   .. group-tab:: Asterisk

      To install Asterisk_, follow these steps:

      .. code-block:: bash

         # Install the necessary dependencies
         sudo apt-get install -y build-essential libasound2-dev autoconf \
                              openssl libssl-dev libxml2-dev \
                              libncurses5-dev uuid-dev sqlite3 \
                              libsqlite3-dev pkg-config libedit-dev \
                              libjansson-dev

         # Download Asterisk
         wget https://downloads.asterisk.org/pub/telephony/asterisk/asterisk-20-current.tar.gz -P /tmp

         # Extract the downloaded archive
         sudo tar -xzvf /tmp/asterisk-20-current.tar.gz -C /usr/src

         # Change the working directory to the extracted Asterisk source
         cd /usr/src/asterisk-20*/

         # Compile and install Asterisk
         sudo ./configure --with-jansson-bundled
         sudo make menuselect.makeopts
         sudo make
         sudo make install
         sudo make samples
         sudo make config
         sudo ldconfig

         # Create the Asterisk system user
         sudo adduser --quiet --system --group --disabled-password --shell /bin/false --gecos "Asterisk" asterisk

   .. group-tab:: Kamailio

      Kamailio_ can be installed using the commands below, as documented in the `Kamailio Debian Installation Guide <https://kamailio.org/docs/tutorials/devel/kamailio-install-guide-deb/>`_.

      .. code-block:: bash

         wget -O- http://deb.kamailio.org/kamailiodebkey.gpg | sudo apt-key add -
         echo "deb http://deb.kamailio.org/kamailio56 bullseye main" > /etc/apt/sources.list.d/kamailio.list
         apt-get update
         apt-get install kamailio kamailio-extra-modules kamailio-json-modules 

   .. group-tab:: OpenSIPS

      We got OpenSIPS_ installed via following commands:

      .. code-block:: bash

         apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 049AD65B
         echo "deb https://apt.opensips.org buster 3.3-releases" >/etc/apt/sources.list.d/opensips.list
         echo "deb https://apt.opensips.org buster cli-nightly" >/etc/apt/sources.list.d/opensips-cli.list
         apt-get update
         sudo apt-get install opensips opensips-mysql-module opensips-cgrates-module opensips-cli

Configuration and initialization
--------------------------------

This section will be dedicated to configuring both the chosen SIP Server, as well as CGRateS and then get them running.

Regarding the SIP Servers, we have prepared custom configurations in advance, as well as an init scripts that can be used to start the services using said configurations. It can also be used to stop/restart/check on the status of the services. Another way to do that would be to copy the configuration in the default folder, where the Server will be searching for the configuration before starting, with it usually being /etc/<software name>.

.. tabs::

   .. group-tab:: FreeSWITCH


      The FreeSWITCH_ setup consists of:

         - *vanilla* configuration + "mod_json_cdr" for CDR generation;
         - configurations for the following users (found in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1006-prepaid, 1007-rated;
         - addition of CGRateS' own extensions befoure routing towards users in the dialplan (found in *etc/freeswitch/dialplan/default.xml*).


      To start FreeSWITCH_ with the prepared custom configuration, run:

      .. code-block:: bash

         /usr/share/cgrates/tutorials/fs_evsock/freeswitch/etc/init.d/freeswitch start

      To verify that FreeSWITCH_ is running, run the following command:

      .. code-block:: bash

         fs_cli -x status


   .. group-tab:: Asterisk


      The Asterisk_ setup consists of:

         - *basic-pbx* configuration sample;
         - configurations for the following users: 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1007-rated.


      To start Asterisk_ with the prepared custom configuration, run:

      .. code-block:: bash

         /usr/share/cgrates/tutorials/asterisk_ari/asterisk/etc/init.d/asterisk start
      

      To verify that Asterisk_ is running, run the following commands:

      .. code-block:: bash

         asterisk -r -s /tmp/cgr_asterisk_ari/asterisk/run/asterisk.ctl
         ari show status

   .. group-tab:: Kamailio

      The Kamailio_ setup consists of:

         - default configuration with small modifications to add **CGRateS** interaction;
         - for script maintainability and simplicity, we have separated **CGRateS** specific routes in *kamailio-cgrates.cfg* file which is included in main *kamailio.cfg* via include directive;
         - configurations for the following users: 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, stored using the CGRateS AttributeS subsystem.


      To start Kamailio_ with the prepared custom configuration, run:

      .. code-block:: bash

         /usr/share/cgrates/tutorials/kamevapi/kamailio/etc/init.d/kamailio start

      To verify that Kamailio_ is running, run the following command:

      .. code-block:: bash

         kamctl moni

   .. group-tab:: OpenSIPS

      The OpenSIPS_ setup consists of:
         - *residential* configuration;
         - user accounts configuration not needed since it's enough for them to only be defined within CGRateS;
         - for simplicity, no authentication was configured (WARNING: Not suitable for production).
         - creating database for the DRouting module, using the following command:
            .. code-block:: bash

               opensips-cli -x database create
               

      To start OpenSIPS_ with the prepared custom configuration, run:

      .. code-block:: bash

         /usr/share/cgrates/tutorials/osips_native/opensips/etc/init.d/opensips start

      To verify that OpenSIPS_ is running, run the following command:

      .. code-block:: bash

         opensipsctl moni



**CGRateS** will be configured with the following subsystems enabled:

 - **SessionS**: started as gateway between the SIP Server and rest of CGRateS subsystems;
 - **ChargerS**: used to decide the number of billing runs for customer/supplier charging;
 - **AttributeS**: used to populate extra data to requests (ie: prepaid/postpaid, passwords, paypal account, LCR profile);
 - **RALs**: used to calculate costs as well as account bundle management;
 - **SupplierS**: selection of suppliers for each session (in case of OpenSIPS_, it will work in tandem with their DRouting module);
 - **StatS**: computing statistics in real-time regarding sessions and their charging;
 - **ThresholdS**: monitoring and reacting to events coming from above subsystems;
 - **CDRe**: exporting rated CDRs from CGR StorDB (export path: */tmp*).

Just as with the SIP Servers, we have also prepared configurations and init scripts for CGRateS. And just as well, you can manage the CGRateS service using systemctl if you prefer. You can even start it using the cgr-engine binary, like so:

 .. code-block:: bash

         cgr-engine -config_path=<path_to_config> -logger=*stdout

.. note::
   The logger flag from the command above is optional, it's usually more convenient for us to check for logs in the terminal that cgrates was started in rather than checking the syslog.


.. tabs::

   .. group-tab:: FreeSWITCH

      .. code-block:: bash

         /usr/share/cgrates/tutorials/fs_evsock/cgrates/etc/init.d/cgrates start

   .. group-tab:: Asterisk

      .. code-block:: bash

         /usr/share/cgrates/tutorials/asterisk_ari/cgrates/etc/init.d/cgrates start

   .. group-tab:: Kamailio

      .. code-block:: bash

         /usr/share/cgrates/tutorials/kamevapi/cgrates/etc/init.d/cgrates start

   .. group-tab:: OpenSIPS

      .. code-block:: bash

         /usr/share/cgrates/tutorials/osips_native/cgrates/etc/init.d/cgrates start

.. note::
   If you have chosen OpenSIPS_, CGRateS has to be started first since the dependency is reversed.


Loading **CGRateS** Tariff Plans
--------------------------------

Now that we have **CGRateS** installed and started with one of the custom configurations, we can load the prepared data out of the shared folder, containing the following rules:

- Create the necessary timings (always, asap, peak, offpeak).
- Configure 3 destinations (1002, 1003 and 10 used as catch all rule).
- As rating we configure the following:

 - Rate id: *RT_10CNT* with connect fee of 20cents, 10cents per minute for the first 60s in 60s increments followed by 5cents per minute in 1s increments.
 - Rate id: *RT_20CNT* with connect fee of 40cents, 20cents per minute for the first 60s in 60s increments, followed by 10 cents per minute charged in 1s increments.
 - Rate id: *RT_40CNT* with connect fee of 80cents, 40cents per minute for the first 60s in 60s increments, follwed by 20cents per minute charged in 10s increments.
 - Rate id: *RT_1CNT* having no connect fee and a rate of 1 cent per minute, chargeable in 1 minute increments.
 - Rate id: *RT_1CNT_PER_SEC* having no connect fee and a rate of 1 cent per second, chargeable in 1 second increments.

- Accounting part will have following configured:

  - Create 3 accounts: 1001, 1002, 1003.
  - 1001, 1002 will receive 10units of **\*monetary** balance.


.. code-block:: bash

 cgr-loader -verbose -path=/usr/share/cgrates/tariffplans/tutorial

To verify that all actions successfully performed, we use following *cgr-console* commands:

- Make sure all our balances were topped-up:

 .. code-block:: bash

  cgr-console 'accounts Tenant="cgrates.org" AccountIds=["1001"]'
  cgr-console 'accounts Tenant="cgrates.org" AccountIds=["1002"]'

- Query call costs so we can see our calls will have expected costs (final cost will result as sum of *ConnectFee* and *Cost* fields):

 .. code-block:: bash
 
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1002" AnswerTime="2014-08-04T13:00:00Z" Usage="20s"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1002" AnswerTime="2014-08-04T13:00:00Z" Usage="1m25s"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1003" AnswerTime="2014-08-04T13:00:00Z" Usage="20s"'


Test calls
----------


1001 -> 1002
~~~~~~~~~~~~

Since the user 1001 is marked as *prepaid* inside the telecom switch, calling between 1001 and 1002 should generate pre-auth and prepaid debits which can be checked with *get_account* command integrated within *cgr-console* tool. Charging will be done based on time of day as described in the tariff plan definition above.

.. note::

   An important particularity to  note here is the ability of **CGRateS** SessionManager to refund units booked in advance (eg: if debit occurs every 10s and rate increments are set to 1s, the SessionManager will be smart enough to refund pre-booked credits for calls stoped in the middle of debit interval).

Check that 1001 balance is properly deducted, during the call, and moreover considering that general balance has priority over the shared one debits for this call should take place at first out of general balance.

.. code-block:: bash

 cgr-console 'accounts Tenant="cgrates.org" AccountIds=["1001"]'


1002 -> 1001
~~~~~~~~~~~~

The user 1002 is marked as *postpaid* inside the telecom switch hence his calls will be debited at the end of the call instead of during a call and his balance will be able to go on negative without influencing his new calls (no pre-auth).

To check that we had debits we use again console command, this time not during the call but at the end of it:

.. code-block:: bash

 cgr-console 'accounts Tenant="cgrates.org" AccountIds=["1002"]'


1001 -> 1003
~~~~~~~~~~~~
The user 1001 call user 1003 and after 12 seconds the call will be disconnected.

CDR Processing
--------------

  - The SIP Server generates a CDR event at the end of each call (i.e., FreeSWITCH_ via HTTP Post and Kamailio_ via evapi)
  - The event is directed towards the port configured inside cgrates.json due to the automatic handler registration built into the SessionS subsystem.
  - The event reaches CGRateS through the SessionS subsystem in close to real-time.
  - Once inside CGRateS, the event is instantly rated and ready for export.


CDR Exporting
-------------

Once the CDRs are mediated, they are available to be exported. One can use available RPC APIs for that or directly call exports from console:

.. code-block:: bash

 cgr-console 'cdrs_export ExportArgs={"ExportFormat":"*file_csv", "ExportPath":"/tmp"}'


Fraud detection
---------------

Since we have configured some action triggers (more than 20 units of balance topped-up or less than 2 and more than 5 units spent on *FS_USERS* we should be notified over syslog when things like unexpected events happen, e.g.: fraud with more than 20 units topped-up). Most important is the monitor for 100 units topped-up which will also trigger an account disable together with killing it's calls if prepaid debits are used.

To verify this mechanism simply add some random units into one account's balance:

.. code-block:: bash

 cgr-console 'balance_set Tenant="cgrates.org" Account="1003" Value=23 BalanceType="*monetary" Balance={"ID":"MonetaryBalance"}'
 tail -f /var/log/syslog -n 20

 cgr-console 'balance_set Tenant="cgrates.org" Account="1001" Value=101 BalanceType="*monetary" Balance={"ID":"MonetaryBalance"}'
 tail -f /var/log/syslog -n 20

On the CDRs side we will be able to integrate CdrStats monitors as part of our Fraud Detection system (eg: the increase of average cost for 1001 and 1002 accounts will signal us abnormalities, hence we will be notified via syslog).


.. _Zoiper: https://www.zoiper.com/
.. _Asterisk: http://www.asterisk.org/
.. _FreeSWITCH: https://freeswitch.com/
.. _Kamailio: https://www.kamailio.org/w/
.. _OpenSIPS: https://opensips.org/
.. _CGRateS: http://www.cgrates.org/