FreeSWITCH installation
=======================

As operating system we have chosen Debian Wheezy, since all the software components we use provide packaging for it.


FreeSWITCH_
-----------

More information regarding installing FreeSWITCH_ on Debian can be found on it's official `installation wiki <https://confluence.freeswitch.org/display/FREESWITCH/Debian#Debian-DebianPackage>`_.

To get FreeSWITCH_ installed and configured, we have choosen the simplest method, out of *vanilla* packages plus one individual module we need: *mod-json-cdr*.

We got FreeSWITCH_ installed via following commands:

::

 gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF
 gpg -a --export D76EDC7725E010CF | sudo apt-key add -
 cd /etc/apt/sources.list.d/
 wget http://apt.itsyscom.com/conf/freeswitch.apt.list
 apt-get update
 apt-get install freeswitch-meta-vanilla freeswitch-mod-json-cdr

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _FreeSWITCH: http://www.freeswitch.org/





