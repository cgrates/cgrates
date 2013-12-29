3.Installation
==============

CGRateS can be installed via packages as well as Go automated source install.
We recommend using source installs for advanced users familiar with Go programming and packages for users not willing to be involved in the code building process.

3.1. Using packages
-------------------

3.1.2. Debian Wheezy (Squeeze is also backwards compatible)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This is for the moment the only packaged and the most recommended to use method to install CGRateS.

On the server you want to install CGRateS, simply execute the following commands:
::

   cd /etc/apt/sources.list.d/
   wget -O - http://apt.itsyscom.com/repos/apt/conf/cgrates.gpg.key|apt-key add -
   wget http://apt.itsyscom.com/repos/apt/conf/cgrates.apt.list
   apt-get update
   apt-get install cgrates

These commands should set you up with a running version of CGRateS on your machine. 
Details regarding running and errors should be checked in the syslog.
Since on Debian we use Daemontools_ to control the CGRateS another way to check the daemon status is to issue following command over your console:
::
   svstat /etc/service/cgrates/

.. _Daemontools: http://cr.yp.to/daemontools.html

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
CGRateS needs at minimum one external database where to keep it's main data as well as logs of it's operation.

At present we support the following databases:
    As DataDB:
     - Redis_
     - mongoDB_
    As LogDB:
     - mongoDB_
     - Redis_
     - PostgreSQL_ (partially supported)

When using PostgreSQL_ as your LogDB, the following data table needs to be created and accessible to CGRateS LogDB user::

        CREATE TABLE callcosts (
            uuid varchar(80) primary key,
            source varchar(32),
            direction varchar(32),
            tenant varchar(32),
            tor varchar(32),
            subject varchar(32),
            account varchar(32),
            destination varchar(32),
            cost real,
            conect_fee real,
            timespans text
        );


.. _Redis: http://redis.io/
.. _PostgreSQL: http://www.postgresql.org/
.. _mongoDB: http://www.mongodb.org/

