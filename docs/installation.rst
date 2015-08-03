3.Installation
==============

CGRateS can be installed via packages as well as Go automated source install.
We recommend using source installs for advanced users familiar with Go programming and packages for users not willing to be involved in the code building process.

3.1. Using packages
-------------------

3.1.2. Debian Jessie/Wheezy
~~~~~~~~~~~~~~~~~~~~~~~~~~~

This is for the moment the only packaged and the most recommended to use method to install CGRateS.

On the server you want to install CGRateS, simply execute the following commands:
::

   cd /etc/apt/sources.list.d/
   wget -O - http://apt.itsyscom.com/conf/cgrates.gpg.key|apt-key add -
   wget http://apt.itsyscom.com/conf/cgrates.apt.list
   apt-get update
   apt-get install cgrates

Once the installation is completed, one should perform the post-install section in order to have the CGRateS properly set and ready to run.
After post-install actions are performed, CGRateS will be configured in */etc/cgrates/cgrates.json* and enabled in */etc/default/cgrates*.

3.2. Using source
-----------------

After the go environment is installed_ (at least go1.2) and configured_ issue the following commands:
::

    go get github.com/cgrates/cgrates

This command will install the trunk version of CGRateS together with all the necessary dependencies.

.. _installed: http://golang.org/doc/install
.. _configured: http://golang.org/doc/code.html


3.3. Post-install
-----------------

Database setup
~~~~~~~~~~~~~~

For it's operation CGRateS uses more database types, depending on it's nature, install and configuration being further necessary. 
At present we support the following databases:

As DataDB types (rating and accounting subsystems):

- Redis_

As StorDB (persistent storage for CDRs and tariff plan versions).

Once installed there should be no special requirements in terms of setup since no schema is necessary.

- MySQL_

Once database is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package)

 ::
   
  cd /usr/share/cgrates/storage/mysql/
  ./setup_cgr_db.sh root CGRateS.org localhost

- PostgreSQL_

Once database is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package)

 ::
   
  cd /usr/share/cgrates/storage/postgres/
  ./setup_cgr_db.sh root CGRateS.org localhost

.. _Redis: http://redis.io/
.. _MySQL: http://www.mysql.org/
.. _PostgreSQL: http://www.postgresql.org/


Git
~~~

The CGR-History component will use Git_ to archive tariff plan changes, hence it's installation is necessary before using CGR-History.

.. _Git: http://git-scm.com/

