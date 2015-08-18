Software installation
=====================

As operating system we have choosen Debian Wheezy, since all the software components we use provide packaging for it.

OpenSIPS_
---------

We got OpenSIPS_ installed via following commands:
::

 wget http://apt.opensips.org/key.asc
 apt-key add key.asc
 cd /etc/apt/sources.list.d/
 wget http://apt.itsyscom.com/conf/opensips.wheezy.apt.list
 apt-get update
 apt-get install opensips opensips-json-module opensips-restclient-module

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _OpenSIPS: http://www.opensips.org/