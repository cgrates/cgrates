Software installation
=====================

We have chosen Debian Bullseye as operating system, since all the software components we use provide packaging for it.

CGRateS
--------

**CGRateS** can be installed using the instructions found :ref:`here<installation>`. 



Kamailio_
---------

We got Kamailio_ installed via following commands, documented in KamailioDebianInstallation_ section:
::

 wget -O- https://deb.kamailio.org/kamailiodebkey.gpg | sudo apt-key add -
 echo "deb http://deb.kamailio.org/kamailio56 bullseye main" > /etc/apt/sources.list.d/kamailio.list
 apt-get update
 apt-get install kamailio kamailio-extra-modules kamailio-json-modules 

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _Kamailio: https://www.kamailio.org/w/
.. _KamailioDebianInstallation: https://www.kamailio.org/wiki/packages/debs
