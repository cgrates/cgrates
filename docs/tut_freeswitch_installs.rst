Software installation
=====================

As operating system we have chosen Debian Jessie, since all the software components we use provide packaging for it.

CGRateS
--------

**CGRateS** can be installed using the instructions found :ref:`here<installation>`. 




FreeSWITCH_
-----------

More information regarding the installation of FreeSWITCH_ on Debian can be found on it's official `installation wiki <https://freeswitch.org/confluence/display/FREESWITCH/FreeSWITCH+1.6+Video>`_.

Firstly, in order to install FreeSWITCH_, the authentication is required by creating a SignalWire Personal Access Token. Before instalation, it's needed to generate the personal token and this cand be found on :ref:`SignalWire official wiki in creating a personal token<https://developer.signalwire.com/freeswitch/FreeSWITCH-Explained/Installation/HOWTO-Create-a-SignalWire-Personal-Access-Token_67240087/#attachments>`.

To get FreeSWITCH_ installed and configured, we have choosen the simplest method, out of *vanilla* packages, plus one individual module we need: *mod-json-cdr*.

We will install FreeSWITCH_ via following commands:

::
 TOKEN=YOURSIGNALWIRETOKEN # here insert your SignalWire Personal Acces Token
 wget --http-user=signalwire --http-password=$TOKEN -O /usr/share/keyrings/signalwire-freeswitch-repo.gpg https://freeswitch.signalwire.com/repo/deb/debian-release/signalwire-freeswitch-repo.gpg
 echo "machine freeswitch.signalwire.com login signalwire password $TOKEN" > /etc/apt/auth.conf
 chmod 600 /etc/apt/auth.conf
 echo "deb [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ `lsb_release -sc` main" > /etc/apt/sources.list.d/freeswitch.list
 echo "deb-src [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ `lsb_release -sc` main" >> /etc/apt/sources.list.d/freeswitch.list
 # if /etc/freeswitch does not exist, the standard vanilla configuration is deployed
 apt-get update && apt-get install -y freeswitch-meta-allapt-get update && apt-get install -y freeswitch-meta-all

Once installed, we will proceed with loading the configuration out of specific tutorial cases bellow.

.. _FreeSWITCH: https://freeswitch.com//
