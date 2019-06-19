**CGRateS** Installation
========================

We have chosen Debian Jessie as operating system, since all the software components we use provide packaging for it.

Prerequisites
-------------

Some components of **CGRateS** (whether enabled or not, is up to the administrator) depend on external software like:

- Git_ used by **CGRateS** History Server as archiver.
- Redis_ to serve as Rating and Accounting DB for **CGRateS**.
- MySQL_ to serve as StorDB for **CGRateS**.

We will install them in one shot using the command bellow.

::

 apt-get install git redis-server mysql-server

*Note*: We will use this MySQL_ root password when asked: *CGRateS.org*.


Installation
------------

Installation steps are provided within the **CGRateS** `install documentation <https://cgrates.readthedocs.org/en/latest/installation.html>`_.

Since this tutorial is for master version of **CGRateS**, we will install CGRateS out of temporary .deb packages built out of master code:

::

 wget http://www.cgrates.org/tmp_pkg/cgrates_0.9.1~rc8_amd64.deb
 dpkg -i cgrates_0.9.1~rc8_amd64.deb

As described in post-install section, we will need to set up the MySQL_ database (using *CGRateS.org* as our root password):

::

 cd /usr/share/cgrates/storage/mysql/
 ./setup_cgr_db.sh root CGRateS.org localhost

Once the database is in place, we can now set versions:
::

   cgr-migrator -stordb_passwd="CGRateS.org" -exec="*set_versions"

At this point we have **CGRateS** installed but not yet configured. To facilitate understanding and speed up the process, **CGRateS** has the configurations used in these tutorials available in the */usr/share/cgrates/tutorials* folder.

.. _Redis: http://redis.io/
.. _MySQL: http://www.mysql.org/
.. _Git: http://git-scm.com/
