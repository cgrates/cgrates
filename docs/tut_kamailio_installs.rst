Software installation
=====================

As operating system we have choosen Debian Wheezy, since all the software components we use provide packaging for it.

Kamailio_
---------

We got Kamailio_ installed via following commands:
::

 apt-key adv --recv-keys --keyserver keyserver.ubuntu.com 0xfb40d3e6508ea4c8
 cd /etc/apt/sources.list.d/
 wget http://apt.itsyscom.com/conf/kamailio.apt.list .
 apt-get update
 apt-get install kamailio kamailio-extra-modules kamailio-json-modules

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _Kamailio: http://www.kamailio.org/