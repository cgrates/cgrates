.. _Redis: https://redis.io/
.. _MySQL: https://dev.mysql.com/
.. _PostgreSQL: https://www.postgresql.org/
.. _MongoDB: https://www.mongodb.com/




.. _installation:

Installation
============

CGRateS can be installed via packages as well as Go automated source install.
We recommend using source installs for advanced users familiar with Go programming and packages for users not willing to be involved in the code building process.


Using packages
--------------

Depending on the packaged distribution, the following methods are available:


Debian 
^^^^^^

There are two main ways of installing the maintained packages:


Aptitude repository 
~~~~~~~~~~~~~~~~~~~


Add the gpg key:

::

    sudo wget -O - https://apt.cgrates.org/apt.cgrates.org.gpg.key | sudo apt-key add -

Add the repository in apt sources list:

::

    echo "deb http://apt.cgrates.org/debian/ nightly main" | sudo tee /etc/apt/sources.list.d/cgrates.list

Update & install:

::

    sudo apt-get update
    sudo apt-get install cgrates


Once the installation is completed, one should perform the :ref:`post_install` section in order to have the CGRateS properly set and ready to run.
After *post-install* actions are performed, CGRateS will be configured in **/etc/cgrates/cgrates.json** and enabled in **/etc/default/cgrates**.


Manual installation of .deb package out of archive server
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~


Run the following commands:

::

    wget http://pkg.cgrates.org/deb/nightly/cgrates_current_amd64.deb
    dpkg -i cgrates_current_amd64.deb

As a side note on http://pkg.cgrates.org/deb/ one can find an entire archive of CGRateS packages.


Redhat/Fedora/CentOS
^^^^^^^^^^^^^^^^^^^^

There are two main ways of installing the maintained packages:


YUM repository
~~~~~~~~~~~~~~


To install CGRateS out of YUM execute the following commands

::

    sudo tee -a /etc/yum.repos.d/cgrates.repo > /dev/null <<EOT
    [cgrates]
    name=CGRateS
    baseurl=http://yum.cgrates.org/yum/nightly/
    enabled=1
    gpgcheck=1
    gpgkey=https://yum.cgrates.org/yum.cgrates.org.gpg.key
    EOT
    sudo yum update
    sudo yum install cgrates

Once the installation is completed, one should perform the :ref:`post_install` section in order to have the CGRateS properly set and ready to run.
After *post-install* actions are performed, CGRateS will be configured in **/etc/cgrates/cgrates.json** and enabled in **/etc/default/cgrates**.


Manual installation of .rpm package out of archive server
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~


Run the following command:

::

    sudo rpm -i http://pkg.cgrates.org/rpm/nightly/cgrates_current.rpm

As a side note on http://pkg.cgrates.org/rpm/ one can find an entire archive of CGRateS packages.


Using source
------------

For developing CGRateS and switching between its versions, we are using the **go mods feature** introduced in go 1.13.

.. _InstallGO:

Install GO Lang
^^^^^^^^^^^^^^^

First we have to setup the GO Lang to our OS. Feel free to download 
the latest GO binary release from https://golang.org/dl/
In this Tutorial we are going to install Go 1.16

::

   sudo rm -rf /usr/local/go
   cd /tmp
   wget https://golang.org/dl/go1.16.linux-amd64.tar.gz
   sudo tar -xvf go1.16.linux-amd64.tar.gz -C /usr/local/
   export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin


Build CGRateS from Source
^^^^^^^^^^^^^^^^^^^^^^^^^

Configure the project with the following commands:

::

   go get github.com/cgrates/cgrates
   cd $HOME/go/src/github.com/cgrates/cgrates
   ./build.sh


Create Debian / Ubuntu Packages from Source
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

After compiling the source code you are ready to create the .deb packages
for your Debian like OS. First lets install some dependencies: 

::

   sudo apt-get install build-essential fakeroot dh-systemd

Finally we are ready to create the system package. Before creation we make
sure that we delete the old one first.

::

   cd $HOME/go/src/github.com/cgrates/cgrates/packages
   rm -rf $HOME/go/src/github.com/cgrates/*.deb
   make deb

After some time and maybe some console warnings, your CGRateS package will be ready.


Install Custom Debian / Ubuntu Package
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

::

   cd $HOME/go/src/github.com/cgrates
   sudo dpkg -i cgrates_*.deb


Generate RPM Packages from Source
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Prerequisites
 * :ref:`Install Golang <InstallGO>`
 * Git

   ::

    sudo apt-get install git


 * RPM

   ::

    sudo apt-get install rpm

Execute the following commands

::

    cd $HOME/go/src/github.com/cgrates/cgrates
    export gitLastCommit=$(git rev-parse HEAD)
    export rpmTag=$(git log -1 --format=%ci | date +%Y%m%d%H%M%S)+$(git rev-parse --short HEAD)
    mkdir -p $HOME/cgr_build/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
    wget -P $HOME/cgr_build/SOURCES https://github.com/cgrates/cgrates/archive/$gitLastCommit.tar.gz
    cp $HOME/go/src/github.com/cgrates/cgrates/packages/redhat_fedora/cgrates.spec $HOME/cgr_build/SPECS
    cd $HOME/cgr_build
    rpmbuild -bb --define "_topdir $HOME/cgr_build" SPECS/cgrates.spec


.. _post_install:

Post-install
------------


Database setup
^^^^^^^^^^^^^^

For its operation CGRateS uses **one or more** database types, depending on its nature, install and configuration being further necessary.

At present we support the following databases:

`Redis`_
  Can be used as :ref:`DataDB`.
  Optimized for real-time information access.
  Once installed there should be no special requirements in terms of setup since no schema is necessary.

`MySQL`_
  Can be used as :ref:`StorDB`.
  Optimized for CDR archiving and offline Tariff Plan versioning.
  Once MySQL is installed, CGRateS database needs to be set-up out of provided scripts. (example for the paths set-up by debian package)

  ::

    cd /usr/share/cgrates/storage/mysql/
    ./setup_cgr_db.sh root CGRateS.org localhost

`PostgreSQL`_
  Can be used as :ref:`StorDB`.
  Optimized for CDR archiving and offline Tariff Plan versioning.
  Once PostgreSQL is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package).

  ::

    cd /usr/share/cgrates/storage/postgres/
    ./setup_cgr_db.sh

`MongoDB`_
  Can be used as :ref:`DataDB` as well as :ref:`StorDB`.
  It is the first database that can be used to store all kinds of data stored from CGRateS from accounts, tariff plans to cdrs and logs.
  Once MongoDB is installed, CGRateS database needs to be set-up out of provided scripts (example for the paths set-up by debian package)

  ::

    cd /usr/share/cgrates/storage/mongo/
    ./setup_cgr_db.sh


Set versions data
^^^^^^^^^^^^^^^^^

Once database setup is completed, we need to write the versions data. To do this, run migrator tool with the parameters specific to your database. 

Sample usage for MySQL: 
::

   cgr-migrator -stordb_passwd="CGRateS.org" -exec="*set_versions"

