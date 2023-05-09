Software Installation
=====================

We recommend using Debian 11 (Bullseye) as the operating system because all the software components we use provide packaging for it.

CGRateS
-------

You can install **CGRateS** by following the instructions provided in the :ref:`installation<installation>` section.

FreeSWITCH
----------

For detailed information on installing FreeSWITCH on Debian, please refer to its official `installation wiki <https://developer.signalwire.com/freeswitch/FreeSWITCH-Explained/Installation/Linux/Debian_67240088/>`_.

Before installing FreeSWITCH, you need to authenticate by creating a SignalWire Personal Access Token. To generate your personal token, follow the instructions in the `SignalWire official wiki on creating a personal token <https://developer.signalwire.com/freeswitch/freeswitch-explained/installation/howto-create-a-signalwire-personal-access-token_67240087/>`_.

To install FreeSWITCH and configure it, we have chosen the simplest method using *vanilla* packages.

Install FreeSWITCH by running the following commands:

::

 TOKEN=YOURSIGNALWIRETOKEN # Insert your SignalWire Personal Access Token here
 apt-get update && apt-get install -y gnupg2 wget lsb-release
 wget --http-user=signalwire --http-password=$TOKEN -O /usr/share/keyrings/signalwire-freeswitch-repo.gpg https://freeswitch.signalwire.com/repo/deb/debian-release/signalwire-freeswitch-repo.gpg
 echo "machine freeswitch.signalwire.com login signalwire password $TOKEN" > /etc/apt/auth.conf
 chmod 600 /etc/apt/auth.conf
 echo "deb [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ `lsb_release -sc` main" > /etc/apt/sources.list.d/freeswitch.list
 echo "deb-src [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ `lsb_release -sc` main" >> /etc/apt/sources.list.d/freeswitch.list

 # If /etc/freeswitch does not exist, the standard vanilla configuration is deployed
 apt-get update && apt-get install -y freeswitch-meta-all

After the installation is complete, we will proceed to load the configuration based on the specific tutorial case provided in the subsequent section.

.. _FreeSWITCH: https://freeswitch.com//
