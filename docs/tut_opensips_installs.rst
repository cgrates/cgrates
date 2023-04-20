Software installation
=====================

We have chosen Debian Buster as operating system, since all the software components we use provide packaging for it.

CGRateS
--------

**CGRateS** can be installed using the instructions found :ref:`here<installation>`. 


OpenSIPS_
---------

We got OpenSIPS_ installed via following commands:
::

 apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 049AD65B
 echo "deb https://apt.opensips.org buster 3.3-releases" >/etc/apt/sources.list.d/opensips.list
 echo "deb https://apt.opensips.org buster cli-nightly" >/etc/apt/sources.list.d/opensips-cli.list
 apt-get update
 sudo apt-get install opensips opensips-mysql-module opensips-cgrates-module opensips-cli

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _OpenSIPS: https://opensips.org/
