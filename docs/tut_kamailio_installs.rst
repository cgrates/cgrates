Software installation
=====================

As operating system we have choosen Debian Jessie, since all the software components we use provide packaging for it.

Kamailio_
---------

We got Kamailio_ installed via following commands:
::

 apt-key adv --recv-keys --keyserver keyserver.ubuntu.com 0xfb40d3e6508ea4c8
 echo "deb http://deb.kamailio.org/kamailio43 jessie main" > /etc/apt/sources.list.d/kamailio.list
 apt-get update
 apt-get install kamailio kamailio-extra-modules kamailio-json-modules

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _Kamailio: http://www.kamailio.org/