Software installation
=====================

As operating system we have choosen Debian Wheezy, since all the software components we use provide packaging for it.

Prerequisites
-------------

Some components of **CGRateS** (whether enabled or not is up to the administrator) depend on external software like:

- Git_ used by **CGRateS** History Server as archiver.
- Redis_ to serve as Rating and Accounting DB for **CGRateS**.
- MySQL_ to serve as StorDB for **CGRateS**.

We will install them in one shoot using the command bellow.

::

 apt-get install git redis-server mysql-server

*Note*: For simplicity sake we have used as MySQL_ root password when asked: ***CGRateS**.org*.


FreeSWITCH_
-----------

More information regarding installing FreeSWITCH_ on Debian can be found on it's official `installation wiki <http://wiki.freeswitch.org/wiki/Installation_Guide#Debian_packages>`_.

To get FreeSWITCH_ installed and configured, we have choosen the simplest method, out of *vanilla* packages.

We got FreeSWITCH_ installed via following commands:

::

 gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF
 gpg -a --export D76EDC7725E010CF | sudo apt-key add -
 cd /etc/apt/sources.list.d/
 wget http://apt.itsyscom.com/repos/apt/conf/freeswitch.apt.list
 apt-get update
 apt-get install freeswitch-meta-vanilla

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.


**CGRateS**
-----------

Installation steps are provided on **CGRateS** `install documentation <https://cgrates.readthedocs.org/en/latest/installation.html>`_.

To get **CGRateS** installed execute the following commands over ssh console:

::

 cd /etc/apt/sources.list.d/
 wget -O - http://apt.itsyscom.com/repos/apt/conf/cgrates.gpg.key|apt-key add -
 wget http://apt.itsyscom.com/repos/apt/conf/cgrates.apt.list
 apt-get update
 apt-get install cgrates

As described in post-install section, we will need to set up the MySQL_ database (using **CGRateS**.org as our root password):

::

 cd /usr/share/cgrates/storage/mysql/
 ./setup_cgr_db.sh root **CGRateS**.org localhost


Since by default FreeSWITCH_ restricts access to *.csv* CDRs to it's own user, we will add the *cgrates* user to freeswitch group.

::

 usermod -a -G freeswitch cgrates


At this point we have **CGRateS** installed but not yet configured. To facilitate the understanding and speed up the process, **CGRateS** comes already with the configurations used in this tutorial, available in the */usr/share/cgrates/tutorials* folder, so we will load them custom on each tutorial case.


SIP UA - Jitsi_
---------------

On our ubuntu desktop host, we have installed Jitsi_ to be used as SIP UA, out of stable provided packages on `Jitsi download <https://jitsi.org/Main/Download>`_ and had Jitsi_ configured with 4 accounts out of default FreeSWITCH_ provided ones: 1001/**CGRateS**.org and 1002/**CGRateS**.org, 1003/**CGRateS**.org and 1004/**CGRateS**.org.


.. _Redis: http://redis.io/
.. _FreeSWITCH: http://www.freeswitch.org/
.. _MySQL: http://www.mysql.org/
.. _Jitsi: http://www.jitsi.org/
.. _Git: http://git-scm.com/ 





