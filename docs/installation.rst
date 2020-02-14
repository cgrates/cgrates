.. _installation:

3. Installation
===============

CGRateS can be installed via packages as well as Go automated source install.
We recommend using source installs for advanced users familiar with Go programming and packages for users not willing to be involved in the code building process.


3.1. Using packages
~~~~~~~~~~~~~~~~~~~

Depending on the packaged distribution, following methods are available:


3.1.1. Debian 
-------------

This is for the moment the only packaged and the most recommended to use method to install CGRateS. CGRateS development team maintains official debian packages out of master branch, released under nightly tag in aptitude. 

There are two main ways of installing the maintained packages:


3.1.1.1. Aptitude repository 
++++++++++++++++++++++++++++


Add the gpg key:

::

    sudo wget -O - http://apt.cgrates.org/apt.cgrates.org.gpg.key | sudo apt-key add -

Add the repository in apt sources list:

::

    echo "deb http://apt.cgrates.org/debian/ nightly main" | sudo tee /etc/apt/sources.list.d/cgrates.list

Update & install:

::

    sudo apt-get update
    sudo apt-get install cgrates


Once the installation is completed, one should perform the :ref:`post-install` section in order to have the CGRateS properly set and ready to run.
After *post-install* actions are performed, CGRateS will be configured in **/etc/cgrates/cgrates.json** and enabled in **/etc/default/cgrates**.


3.1.1.2. Manual installation of .deb package out of archive server
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++


Run the following commands:

::

    wget http://pkg.cgrates.org/deb/master/cgrates_current_amd64.deb
    dpkg -i cgrates_current_amd64.deb

As a side note on http://pkg.cgrates.org one can find an entire archive of CGRateS packages.


3.1.2. Redhat/Fedora/CentOS
-------------

There are two main ways of installing the maintained packages:


3.1.2.1. YUM repository
++++++++++++++++++++++++++++


To install CGRateS out of YUM execute the following commands

::

    sudo tee -a /etc/yum.repos.d/cgrates.repo > /dev/null <<EOT
    [cgrates]
    name=CGRateS
    baseurl=http://yum.cgrates.org/yum/master/
    enabled=1
    gpgcheck=1
    gpgkey=http://yum.cgrates.org/yum.cgrates.org.gpg.key
    EOT
    sudo yum update
    sudo yum install cgrates

Once the installation is completed, one should perform the :ref:`post-install` section in order to have the CGRateS properly set and ready to run.
After *post-install* actions are performed, CGRateS will be configured in **/etc/cgrates/cgrates.json** and enabled in **/etc/default/cgrates**.


3.1.2.2. Manual installation of .rpm package out of archive server
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++


Run the following commands:

::

    sudo rpm -i http://pkg.cgrates.org/rpm/master/cgrates_current.rpm

As a side note on http://pkg.cgrates.org one can find an entire archive of CGRateS packages.


3.2. Using source
~~~~~~~~~~~~~~~~~

For developing CGRateS and switching between its versions, we are using the **new go mods feature** introduced in go 1.13.


3.2.1 Install GO Lang
---------------------

First we have to setup the GO Lang to our OS. Feel free to download 
the latest GO binary release from https://golang.org/dl/
In this Tutorial we are going to install Go 1.13

::

   rm -rf /usr/local/go
   cd /tmp
   wget https://dl.google.com/go/go1.13.1.linux-amd64.tar.gz
   sudo tar -xvf go1.13.1.linux-amd64.tar.gz -C /usr/local/
   export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin


3.2.2 Build CGRateS from Source
-------------------------------

Configure the project with the following commands:

::

   go get github.com/cgrates/cgrates
   cd $GOPATH/src/github.com/cgrates/cgrates
   ./build.sh


3.2.3 Create Debian / Ubuntu Packages from Source
-------------------------------------------------

After compiling the source code you are ready to create the .deb packages
for your Debian like OS. But First lets install some dependencies. 

::

   sudo apt-get install build-essential fakeroot dh-systemd

Finally we are ready to create the system package. Before creation we make
sure that we delete the old one first.

::

   cd $GOPATH/src/github.com/cgrates/cgrates/packages
   rm -rf $GOPATH/src/github.com/cgrates/*.deb
   make deb

After some time and maybe some console warnings, your CGRateS package will be ready.


3.2.4 Install Custom Debian / Ubuntu Package
--------------------------------------------

::

   cd $GOPATH/src/github.com/cgrates
   sudo dpkg -i cgrates_*.deb


.. _post-install:
3.3. Post-install
~~~~~~~~~~~~~~~~~

3.3.1. Database setup
---------------------

For its operation CGRateS uses **one or more** database types, depending on its nature, install and configuration being further necessary.

At present we support the following databases:

- `Redis`_
Can be used as ``data_db`` .
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
Can be used as ``data_db`` - ``stor_db`` .
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

3.3.2 Set versions data
------------------------
Once database setup is completed, we need to write the versions data. To do this, run migrator tool with the parameters specific to your database. 

Sample usage for MySQL: 
::

   cgr-migrator -stordb_passwd="CGRateS.org" -exec="*set_versions"

