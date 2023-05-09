Software installation
=====================

We have chosen Debian Jessie as operating system, since all the software components we use provide packaging for it.

CGRateS
--------

**CGRateS** can be installed using the instructions found :ref:`here<installation>`. 


OpenSIPS_
---------

We got OpenSIPS_ installed via following commands:
::

 apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 049AD65B
 echo "deb http://apt.opensips.org jessie 2.4-nightly" >/etc/apt/sources.list.d/opensips.list
 apt-get update
 apt-get install opensips opensips-cgrates-module

After the installation is complete, we will proceed to load the configuration based on the specific tutorial case provided in the subsequent section.

.. _OpenSIPS: https://www.opensips.org/
