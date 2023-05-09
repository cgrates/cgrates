Software installation
=====================

We have chosen Debian 11 (Bullseye) as the operating system, since all the software components we use provide packaging for it.

CGRateS
-------

CGRateS can be installed by following the instructions in this :ref:`installation guide<installation>`.

Kamailio
--------

Kamailio can be installed using the commands below, as documented in the `Kamailio Debian Installation Guide <https://kamailio.org/docs/tutorials/devel/kamailio-install-guide-deb/>`_.

::

 wget -O- http://deb.kamailio.org/kamailiodebkey.gpg | sudo apt-key add -
 echo "deb http://deb.kamailio.org/kamailio56 bullseye main" > /etc/apt/sources.list.d/kamailio.list
 apt-get update
 apt-get install kamailio kamailio-extra-modules kamailio-json-modules 

After the installation is complete, we will proceed to load the configuration based on the specific tutorial case provided in the subsequent section.

.. _Kamailio: https://www.kamailio.org/w/
.. _KamailioDebianInstallation: https://www.kamailio.org/wiki/packages/debs
