7.1. FreeSWITCH Integration Tutorials
=====================================
 
7.1.1. Automated prepaid with CSV CDRs
--------------------------------------

In this tutorial we are going to focus on simplicity of integrating CGRateS as carrier grade real-time charging engine for FreeSWITCH_. CGRateS will serve as both prepaid controller and rater for FreeSWITCH_ default generated CSV CDRs. A screencast version of this tutorial can be found on `YouTube <http://youtu.be/qTQZZpb-m7Q>`_.


7.1.1.1. Prerequisites
~~~~~~~~~~~~~~~~~~~~~~

OS: Debian Wheezy. Default options selected for installer.

Install Redis_ to serve as DataDB for CGRateS.

::

 apt-get install redis-server


Install MongoDB_ to serve as LogDB for CGRateS.

::

  apt-get install mongodb


7.1.1.2. FreeSWITCH_
~~~~~~~~~~~~~~~~~~~~

More information regarding installing FreeSWITCH_ on Debian can be found on it's official `installation wiki <http://wiki.freeswitch.org/wiki/Installation_Guide#Debian_packages>`_.

To get FreeSWITCH_ installed and configured, we have choosen the simplest method, out of *vanilla* packages. 
Bellow are the commands we have used to get FreeSWITCH_ up.

::

 gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF
 gpg -a --export D76EDC7725E010CF | sudo apt-key add -
 cd /etc/apt/sources.list.d/
 wget http://apt.itsyscom.com/repos/apt/conf/freeswitch.apt.list
 apt-get update
 apt-get install freeswitch-meta-vanilla
 cp -r /usr/share/freeswitch/conf/vanilla /etc/freeswitch
 chown -R freeswitch:freeswitch /etc/freeswitch/
 /etc/init.d/freeswitch start
 fs_cli


7.1.1.2. CGRateS
~~~~~~~~~~~~~~~~

Installation steps are provided on CGRateS `install documentation <https://cgrates.readthedocs.org/en/latest/installation.html>`_.

To get CGRateS installed and configured, we have executed the following commands over console:

::

 cd /etc/apt/sources.list.d/
 wget -O - http://apt.itsyscom.com/repos/apt/conf/cgrates.gpg.key|apt-key add -
 wget http://apt.itsyscom.com/repos/apt/conf/cgrates.apt.list
 apt-get update
 apt-get install cgrates
 cd /usr/share/cgrates/data/tariffplans/prepaid1centpsec/
 cgr-loader
 cd /etc/cgrates/
 cp /usr/share/cgrates/data/conf/cgr_fs_prep_csv.cfg cgrates.cfg
 svc -d /etc/service/cgrates/
 svc -u /etc/service/cgrates/
 tail -f /var/log/syslog


7.1.1.3. Final integration tests - Jitsi
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ 

On our ubuntu desktop host, we have installed Jitsi_ to be used as SIP UA, out of stable provided packages on `Jitsi download <https://jitsi.org/Main/Download>`_ and had Jitsi_ configured with 2 accounts out of default FreeSWITCH_ provided ones: 1001/1234 and 1002/1234.

Calling between 1001 and 1002 should generate prepaid debits which are to be monitored in */var/log/syslog*.

To check rating simply rotate the cdr files via fs_cli, checking the CDR prices in the location where CGRateS moves them, */var/log/cgrates/cdr_out*.

::

 fs_cli -x "cdr_csv rotate"

.. _Redis: http://redis.io/
.. _MongoDB: http://www.mongodb.org/
.. _FreeSWITCH: http://www.freeswitch.org/
.. _Jitsi: http://www.jitsi.org/





 

