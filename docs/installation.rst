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

    go get github.com/cgrates/cgrates/...

This command will install the trunk version of CGRateS together with all the necessary dependencies.

For developing CGRateS and switching betwen lts versions we are using the new (experimental) vendor directory feature introduced in go 1.5. In a nutshell all the dependencies are installed and used from a folder named vendor placed in the root of the project.

To manage this vendor folder we use a tool named glide_ which will download specific versions of the external packages used by CGRateS. To configure the project with glide use the following commands:
::
   export GO15VENDOREXPERIMENT=1 #this should be placed in the rc script of your shell
   go get github.com/cgrates/cgrates
   cd $GOPATH/src/github.com/cgrates/cgrates
   glide install

The glide up command will install the external dependencies versions specified in the glide.yaml file in the vendor folder. There are different versions for each CGRateS branch, versions that are recorded in the yaml file when the GCRateS releases are made (using glide pin command).

Note that the vendor folder should not be registered with the VCS we are using. For more information and command options for use glide_ readme page.

.. _installed: http://golang.org/doc/install
.. _configured: http://golang.org/doc/code.html
.. _glide: https://github.com/Masterminds/glide


3.3. Post-install
-----------------

Database setup
~~~~~~~~~~~~~~

For it's operation CGRateS uses more database types, depending on it's nature, install and configuration being further necessary.

At present we support the following databases:



- Redis_

Used as DataDb, optimized for real-time information access.
Once installed there should be no special requirements in terms of setup since no schema is necessary.


- MySQL_

Used as StorDb, optimized for CDR archiving and offline Tariff Plan versioning.
Once database is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package)

 ::

  cd /usr/share/cgrates/storage/mysql/
  ./setup_cgr_db.sh root CGRateS.org localhost

- PostgreSQL_

Used as StorDb, optimized for CDR archiving and offline Tariff Plan versioning.
Once database is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package)

 ::

  cd /usr/share/cgrates/storage/postgres/
  ./setup_cgr_db.sh

.. _Redis: http://redis.io/
.. _MySQL: http://www.mysql.org/
.. _PostgreSQL: http://www.postgresql.org/


Git
~~~

The CGR-History component will use Git_ to archive tariff plan changes, hence it's installation is necessary before using CGR-History.

.. _Git: http://git-scm.com/
