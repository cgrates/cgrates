Software installation
=====================

As operating system we have choosen Debian Jessie, since all the software components we use provide packaging for it.

OpenSIPS_
---------

We got OpenSIPS_ installed via following commands:
::

 apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 049AD65B
 echo "deb http://apt.opensips.org jessie 2.2-releases" >>/etc/apt/sources.list
 apt-get update
 apt-get install opensips opensips-json-module opensips-restclient-module

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _OpenSIPS: http://www.opensips.org/
