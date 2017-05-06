**CGRateS** Installation
========================

As operating system we have choosen Debian Wheezy, since all the software components we use provide packaging for it.

Prerequisites
-------------

Some components of **CGRateS** (whether enabled or not is up to the administrator) depend on external software like:

- Git_ used by **CGRateS** History Server as archiver.
- Redis_ to serve as Rating and Accounting DB for **CGRateS**.
- MySQL_ to serve as StorDB for **CGRateS**.

We will install them in one shot using the command bellow.

::

 apt-get install git redis-server mysql-server

*Note*: For simplicity sake we have used as MySQL_ root password when asked: *CGRateS.org*.


Installation
------------

Installation steps are provided within **CGRateS** `install documentation <https://cgrates.readthedocs.org/en/latest/installation.html>`_.

Since this tutorial is for master version of **CGRateS** we will install CGRateS out of temporary .deb packages built out of master code:

::

 wget http://www.cgrates.org/tmp_pkg/cgrates_0.9.1~rc8_amd64.deb
 dpkg -i cgrates_0.9.1~rc8_amd64.deb

As described in post-install section, we will need to set up the MySQL_ database (using *CGRateS.org* as our root password):

::

 cd /usr/share/cgrates/storage/mysql/
 ./setup_cgr_db.sh root CGRateS.org localhost


At this point we have **CGRateS** installed but not yet configured. To facilitate the understanding and speed up the process, **CGRateS** comes already with the configurations used in these tutorials, available in the */usr/share/cgrates/tutorials* folder, so we will load them custom on each tutorial case.

.. _Redis: http://redis.io/
.. _MySQL: http://www.mysql.org/
.. _Git: http://git-scm.com/ 
