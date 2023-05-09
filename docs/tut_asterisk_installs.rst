Software installation
=====================

We have chosen Debian 11 (Bullseye) as the operating system.

CGRateS
-------

*CGRateS* can be installed using the instructions found :ref:`here<installation>`.

Asterisk
--------

To install Asterisk, follow these steps:

1. Install the necessary dependencies:

   ::

      sudo apt-get install build-essential libasound2-dev autoconf openssl libssl-dev libxml2-dev libncurses5-dev uuid-dev sqlite3 libsqlite3-dev pkg-config libedit-dev libjansson-dev

2. Download Asterisk:

   Replace `<ASTERISK_VERSION>` with the desired version, e.g., `20-current`.

   ::

      wget https://downloads.asterisk.org/pub/telephony/asterisk/asterisk-<ASTERISK_VERSION>.tar.gz -P /tmp

3. Extract the downloaded archive:

   ::

      sudo tar -xzvf /tmp/asterisk-<ASTERISK_VERSION>.tar.gz -C /usr/src

4. Change the working directory to the extracted Asterisk source:

   Replace `<ASTERISK_MAJOR_VERSION>` with the major version number of the downloaded Asterisk version, e.g., `20`.

   ::

      cd /usr/src/asterisk-<ASTERISK_MAJOR_VERSION>*/

5. Compile and install Asterisk:

   ::

      sudo ./configure --with-jansson-bundled
      sudo make menuselect.makeopts
      sudo make
      sudo make install
      sudo make samples
      sudo make config
      sudo ldconfig

6. Create the Asterisk system user:

   ::

      sudo adduser --quiet --system --group --disabled-password --shell /bin/false --gecos "Asterisk" asterisk

After the installation is complete, we will proceed to load the configuration based on the specific tutorial case provided in the subsequent section.

.. _Asterisk: http://www.asterisk.org/
