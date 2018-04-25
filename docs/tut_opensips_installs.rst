Software installation
=====================

We have chosen Debian Jessie as operating system, since all the software components we use provide packaging for it.

OpenSIPS_
---------

We got OpenSIPS_ installed via following commands:
::

 apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 049AD65B
 echo "deb http://apt.opensips.org jessie 2.4-nightly" >/etc/apt/sources.list.d/opensips.list
 apt-get update
 apt-get install opensips opensips-cgrates-module

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _OpenSIPS: http://www.opensips.org/
