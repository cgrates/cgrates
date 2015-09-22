Software installation
=====================

As operating system we have choosen Debian Jessie, since all the software components we use provide packaging for it.

OpenSIPS_
---------

We got OpenSIPS_ installed via following commands:
::

 wget -O - http://apt.opensips.org/key.asc | apt-key add -
 echo "deb http://apt.opensips.org/debian/stable-2.1/jessie opensips-2.1-jessie main" > /etc/apt/sources.list.d/opensips.list
 apt-get update
 apt-get install opensips opensips-json-module opensips-restclient-module

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _OpenSIPS: http://www.opensips.org/