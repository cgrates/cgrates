3. Installation
===============

CGRateS can be installed via packages as well as Go automated source install.
We recommend using source installs for advanced users familiar with Go programming and packages for users not willing to be involved in the code building process.

3.1. Using packages
-------------------

3.1.1. Debian Jessie/Wheezy
~~~~~~~~~~~~~~~~~~~~~~~~~~~

This is for the moment the only packaged and the most recommended to use method to install CGRateS.

On the server you want to install CGRateS, simply execute the following commands:

::

   cd /etc/apt/sources.list.d/
   wget -O - http://apt.itsyscom.com/conf/cgrates.gpg.key|apt-key add -
   wget http://apt.itsyscom.com/conf/cgrates.apt.list
   apt-get update
   apt-get install cgrates

Once the installation is completed, one should perform the :ref:`post-install` section in order to have the CGRateS properly set and ready to run.
After *post-install* actions are performed, CGRateS will be configured in **/etc/cgrates/cgrates.json** and enabled in **/etc/default/cgrates**.

3.2. Using source
-----------------

For developing CGRateS and switching between its versions, we are using the **new vendor directory feature** introduced in go 1.6. 
In a nutshell all the dependencies are installed and used from a folder named *vendor* placed in the root of the project.

To manage this *vendor* folder we use a tool named `glide`_ which will download specific versions of the external packages used by CGRateS. 
To configure the project with `glide`_ use the following commands:

::

   go get github.com/Masterminds/glide
   go get github.com/cgrates/cgrates
   cd $GOPATH/src/github.com/cgrates/cgrates
   glide install
   ./build.sh

The **glide install** command will install the external dependencies versions, specified in the **glide.lock** file, in the vendor folder. 
There are different versions for each CGRateS branch, versions that are recorded in the **lock** file when the GCRateS releases are made (using **glide update** command).

.. note:: The *vendor* folder **should not** be registered with the VCS we are using.

For more information and command options use `glide`_ readme page.

.. _installed: http://golang.org/doc/install
.. _configured: http://golang.org/doc/code.html
.. _glide: https://github.com/Masterminds/glide

.. _post-install:

3.3. Post-install
-----------------

3.3.1. Database setup
~~~~~~~~~~~~~~~~~~~~~

For its operation CGRateS uses **one or more** database types, depending on its nature, install and configuration being further necessary.

At present we support the following databases:

- `Redis`_
Can be used as ``tariffplan_db`` - ``data_db`` .
Optimized for real-time information access.
Once installed there should be no special requirements in terms of setup since no schema is necessary.

- `MySQL`_
Can be used as ``stor_db`` .
Optimized for CDR archiving and offline Tariff Plan versioning.
Once MySQL is installed, CGRateS database needs to be set-up out of provided scripts. (example for the paths set-up by debian package)

::

   cd /usr/share/cgrates/storage/mysql/
   ./setup_cgr_db.sh root CGRateS.org localhost

- `PostgreSQL`_
Can be used as ``stor_db`` .
Optimized for CDR archiving and offline Tariff Plan versioning.
Once PostgreSQL is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package)

::

   cd /usr/share/cgrates/storage/postgres/
   ./setup_cgr_db.sh

- `MongoDB`_
Can be used as ``tariffplan_db`` - ``data_db`` - ``stor_db`` .
It is the first database that can be used to store all kinds of data stored from CGRateS from accounts, tariff plans to cdrs and logs.
This is provided as an alternative to Redis and/or MySQL/PostgreSQL and right now there are NO plans to drop support for any of them soon.

Once MongoDB is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package)

::

   cd /usr/share/cgrates/storage/mongo/
   ./setup_cgr_db.sh

.. _Redis: http://redis.io
.. _MySQL: http://www.mysql.org
.. _PostgreSQL: http://www.postgresql.org
.. _MongoDB: http://www.mongodb.org


3.3.2.Git
~~~~~~~~~

The **historys** (History Service) component will use `Git`_ to archive *tariff plan changes* in a local repository, 
hence `Git`_ installation is necessary if you want to use this service.

.. _Git: http://git-scm.com
