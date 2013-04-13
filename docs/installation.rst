3.Installation
==============

CGRateS can be installed via packages as well as Go automated source install.
We recommend using source installs for advanced users familiar with Go programming and packages for users not willing to be involved in the code building process.

3.1. Using packages
-------------------
Details will come here.

3.2. Using source
-----------------

After the go environment is installed_ and configured_ issue the following commands:
::

        go get github.com/cgrates/cgrates

This command will install the trunk version of CGRateS together with all the necessary dependencies.

.. _installed: http://golang.org/doc/install
.. _configured: http://golang.org/doc/code.html

Post-install
--------------
CGRateS needs at minimum one external database where to keep it's main data as well as logs of it's operation.

At present we support the following databases:
    As DataDB:
     - Redis_
    As LogDB:
     - PostgreSQL_
     - mongoDB_
     - Redis_

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

