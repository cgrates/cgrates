Software installation
=====================

We have chosen Debian Jessie as operating system, since all the software components we use provide packaging for it.

Kamailio_
---------

We got Kamailio_ installed via following commands:
::

 wget -O- http://deb.kamailio.org/kamailiodebkey.gpg | sudo apt-key add -
 echo "deb http://deb.kamailio.org/kamailio52 stretch main" > /etc/apt/sources.list.d/kamailio.list
 apt-get update
 apt-get install kamailio kamailio-extra-modules kamailio-json-modules

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _Kamailio: http://www.kamailio.org/
