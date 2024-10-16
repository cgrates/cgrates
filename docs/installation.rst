.. _Redis: https://redis.io/
.. _MySQL: https://dev.mysql.com/
.. _PostgreSQL: https://www.postgresql.org/
.. _MongoDB: https://www.mongodb.com/

.. _installation:

Installation
============

.. contents::
   :local:
   :depth: 2

CGRateS can be installed either via packages or through an automated Go source installation. We recommend the latter for advanced users familiar with Go programming, and package installations for those not wanting to engage in the code building process.

After completing the installation, you need to perform the :ref:`post-install configuration <post_install>` steps to set up CGRateS properly and prepare it to run. After these steps, CGRateS will be configured in **/etc/cgrates/cgrates.json** and the service can be managed using the **systemctl** command.

Package Installation
--------------------

Package installation method varies according to the Linux distribution:

Debian or Debian-based Distributions 
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

You can add the CGRateS repository to your system's sources list, depending of the Debian version you are running, as follows:

.. tabs::

   .. group-tab:: Bookworm

      .. code-block:: bash

         # Install dependencies
         sudo apt-get install wget gnupg -y

         # Download and move the GPG Key to the trusted area
         wget https://apt.cgrates.org/apt.cgrates.org.gpg.key -O apt.cgrates.org.asc
         sudo mv apt.cgrates.org.asc /etc/apt/trusted.gpg.d/

         # Add the repository to the apt sources list
         echo "deb http://apt.cgrates.org/debian/ master-bookworm main" | sudo tee /etc/apt/sources.list.d/cgrates.list

         # Update the system repository and install CGRateS
         sudo apt-get update -y
         sudo apt-get install cgrates -y

      Alternatively, you can manually install a specific .deb package as follows:

      .. code-block:: bash

         wget http://pkg.cgrates.org/deb/master/bookworm/cgrates_current_amd64.deb
         sudo dpkg -i ./cgrates_current_amd64.deb

   .. group-tab:: Bullseye

      .. code-block:: bash

         # Install dependencies
         sudo apt-get install wget gnupg -y

         # Download and move the GPG Key to the trusted area
         wget https://apt.cgrates.org/apt.cgrates.org.gpg.key -O apt.cgrates.org.asc
         sudo mv apt.cgrates.org.asc /etc/apt/trusted.gpg.d/

         # Add the repository to the apt sources list
         echo "deb http://apt.cgrates.org/debian/ master-bullseye main" | sudo tee /etc/apt/sources.list.d/cgrates.list

         # Update the system repository and install CGRateS
         sudo apt-get update -y
         sudo apt-get install cgrates -y

      Alternatively, you can manually install a specific .deb package as follows:

      .. code-block:: bash

         wget http://pkg.cgrates.org/deb/master/bullseye/cgrates_current_amd64.deb
         sudo dpkg -i ./cgrates_current_amd64.deb


.. note::
   A complete archive of CGRateS packages is available at http://pkg.cgrates.org/deb/.


Redhat-based Distributions
^^^^^^^^^^^^^^^^^^^^^^^^^^

For .rpm distros, we are using copr to manage the CGRateS packages:

-  If using a version of Linux with dnf:

   .. code-block:: bash

      # sudo yum install -y dnf-plugins-core on RHEL 8 or CentOS Stream
      sudo dnf install -y dnf-plugins-core 
      sudo dnf copr -y enable cgrates/master 
      sudo dnf install -y cgrates

-  For older distributions: 

   .. code-block:: bash

      sudo yum install -y yum-plugin-copr
      sudo yum copr -y enable cgrates/master
      sudo yum install -y cgrates

To install a specific version of the package, run:

.. code-block:: bash

   sudo dnf install -y cgrates-<version>.x86_64

Alternatively, you can manually install a specific .rpm package as follows:

.. code-block:: bash

   wget http://pkg.cgrates.org/rpm/nightly/epel-9-x86_64/cgrates-current.rpm
   sudo dnf install ./cgrates_current.rpm


.. note::
   The entire archive of CGRateS rpm packages is available at https://copr.fedorainfracloud.org/coprs/cgrates/master/packages/ or http://pkg.cgrates.org/rpm/nightly/.

Installing from Source
----------------------

Prerequisites:
^^^^^^^^^^^^^^

- **Git**

.. code-block:: bash

   sudo apt-get install -y git
   # sudo dnf install -y git for .rpm distros

- **Go** (refer to the official Go installation docs: https://go.dev/doc/install)

To install the latest Go version at the time of writing this documentation, run:

.. code-block:: bash

   sudo apt-get install -y wget tar 
   # sudo dnf install -y wget tar for .rpm distros
   sudo rm -rf /usr/local/go
   cd /tmp
   wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin

Installation:
^^^^^^^^^^^^^

.. code-block:: bash

   mkdir -p $HOME/go/src/github.com/cgrates/cgrates
   git clone https://github.com/cgrates/cgrates.git $HOME/go/src/github.com/cgrates/cgrates
   cd $HOME/go/src/github.com/cgrates/cgrates

   # Compile the binaries and move them to $GOPATH/bin
   ./build.sh

   # Create a symbolic link to the data folder
   sudo ln -s $HOME/go/src/github.com/cgrates/cgrates/data /usr/share/cgrates

   # Make cgr-engine binary available system-wide
   sudo ln -s $HOME/go/bin/cgr-engine /usr/bin/cgr-engine

   # Optional: Additional useful symbolic links
   sudo ln -s $HOME/go/bin/cgr-loader /usr/bin/cgr-loader
   sudo ln -s $HOME/go/bin/cgr-migrator /usr/bin/cgr-migrator
   sudo ln -s $HOME/go/bin/cgr-console /usr/bin/cgr-console
   sudo ln -s $HOME/go/bin/cgr-tester /usr/bin/cgr-tester

Creating Your Own Packages
--------------------------

After compiling the source code, you may choose to create your own packages.

For Debian-based distros:
^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

   # Install dependencies
   sudo apt-get install build-essential fakeroot dh-systemd -y

   cd $HOME/go/src/github.com/cgrates/cgrates/packages

   # Delete old ones, if any
   rm -rf $HOME/go/src/github.com/cgrates/*.deb

   make deb

.. note::
   You might see some console warnings, which can be safely ignored.

To install the generated package, run:

.. code-block:: bash

   cd $HOME/go/src/github.com/cgrates
   sudo dpkg -i cgrates_*.deb

For Redhat-based distros:
^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

   sudo dnf install -y rpm-build wget curl tar

   # Create build directories
   mkdir -p $HOME/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

   # Fetch source code
   cd $HOME/go/src/github.com/cgrates/cgrates
   export gitLastCommit=$(git rev-parse HEAD)
   export rpmTag=$(git log -1 --format=%ci | date +%Y%m%d%H%M%S)+$(git rev-parse --short HEAD)

   #Create the tarball from the source code
   cd ..
   tar -czvf  $HOME/rpmbuild/SOURCES/$gitLastCommit.tar.gz cgrates

   # Copy RPM spec file
   cp $HOME/go/src/github.com/cgrates/cgrates/packages/redhat_fedora/cgrates.spec $HOME/rpmbuild/SPECS

   # Build RPM package
   cd $HOME/rpmbuild
   rpmbuild -bb  SPECS/cgrates.spec

.. _post_install:

Post-install Configuration
--------------------------

Database Setup
^^^^^^^^^^^^^^

CGRateS supports multiple database types for various operations, based on your installation and configuration.

Currently, we support the following databases:

`Redis`_
  This can be used as :ref:`DataDB`. It is optimized for real-time information access. Post-installation, no additional setup is required as Redis doesn't require a specific schema.

`MySQL`_
  This can be used as :ref:`StorDB` and is optimized for CDR archiving and offline Tariff Plan versioning. Post-installation, you need to set up the CGRateS database using the provided scripts:

.. code-block:: bash

   cd /usr/share/cgrates/storage/mysql/
   sudo ./setup_cgr_db.sh root CGRateS.org localhost

`PostgreSQL`_
  Like MySQL, PostgreSQL can be used as :ref:`StorDB`. Post-installation, you need to set up the CGRateS database using the provided scripts:

.. code-block:: bash

   cd /usr/share/cgrates/storage/postgres/
   ./setup_cgr_db.sh

`MongoDB`_
  MongoDB can be used as both :ref:`DataDB` and :ref:`StorDB`. This is the first database that can store all types of data from CGRateS - from accounts, tariff plans to CDRs and logs. Post-installation, you need to set up the CGRateS database using the provided scripts:

.. code-block:: bash

   cd /usr/share/cgrates/storage/mongo/
   ./setup_cgr_db.sh

Set Versions Data
^^^^^^^^^^^^^^^^^

After completing the database setup, you need to write the versions data. To do this, run the migrator tool with the parameters specific to your database. 

Sample usage for MySQL: 

.. code-block:: bash

   cgr-migrator -stordb_passwd="CGRateS.org" -exec="*set_versions"
