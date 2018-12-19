Software installation
=====================

We have chosen Debian Jessie as operating system.

Asterisk_
---------

We got Asterisk14_  installed via following commands:
::

 apt-get install autoconf build-essential openssl libssl-dev libsrtp-dev libxml2-dev libncurses5-dev uuid-dev sqlite3 libsqlite3-dev pkg-config libedit-dev
 cd /tmp
 wget --no-check-certificate https://raw.githubusercontent.com/asterisk/third-party/master/pjproject/2.7.2/pjproject-2.7.2.tar.bz2
 wget --no-check-certificate https://raw.githubusercontent.com/asterisk/third-party/master/jansson/2.11/jansson-2.11.tar.bz2
 wget http://downloads.asterisk.org/pub/telephony/asterisk/asterisk-16-current.tar.gz
 tar xzvf asterisk-16-current.tar.gz
 cd asterisk-16.1.0/
 ./configure --with-jansson-bundled
 make
 make install
 adduser --quiet --system --group --disabled-password --shell /bin/false --gecos "Asterisk" asterisk || true


Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _Asterisk14: http://www.asterisk.org/
