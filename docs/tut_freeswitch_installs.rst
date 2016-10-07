Software installation
=====================

As operating system we have chosen Debian Jessie, since all the software components we use provide packaging for it.


FreeSWITCH_
-----------

More information regarding installing FreeSWITCH_ on Debian can be found on it's official `installation wiki <https://freeswitch.org/confluence/display/FREESWITCH/FreeSWITCH+1.6+Video>`_.

To get FreeSWITCH_ installed and configured, we have choosen the simplest method, out of *vanilla* packages plus one individual module we need: *mod-json-cdr*.

We got FreeSWITCH_ installed via following commands:

::

 wget -O - http://files.freeswitch.org/repo/deb/freeswitch-1.6/key.gpg |apt-key add -
 echo "deb http://files.freeswitch.org/repo/deb/freeswitch-1.6/ jessie main" > /etc/apt/sources.list.d/freeswitch.list
 apt-get update
 apt-get install freeswitch-meta-vanilla freeswitch-mod-json-cdr libyuv-dev

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _FreeSWITCH: http://www.freeswitch.org/
