Software installation
=====================

As operating system we have choosen Debian stable.

Asterisk_
---------

We got Asterisk14_  installed via following commands:
::

 apt-get install autoconf build-essential openssl libssl-dev libsrtp-dev libxml2-dev libncurses5-dev uuid-dev sqlite3 libsqlite3-dev pkg-config libjansson-dev
 cd /tmp/
 wget http://downloads.asterisk.org/pub/telephony/asterisk/asterisk-14-current.tar.gz
 tar xzvf asterisk-14-current.tar.gz
 cd asterisk-14.0.2/
 ./configure --with-pjproject-bundled
 make
 make install
 adduser --quiet --system --group --disabled-password --shell /bin/false --gecos "Asterisk" asterisk || true

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.

.. _Asterisk14: http://www.asterisk.org/
