.. _JanusAgent:

JanusAgent
=============


**JanusAgent** is an api endpoint that connects to JanusServer through **CGRateS**. 
It authorizes webrtc events in **CGRateS** for each user and after managing and creating sessions in JanusServer.

The **JanusAgent** is configured within *janus_agent* section from :ref:`JSON configuration <configuration>`.
It will listen on http port 2080 in /janus endpoint as specified in config ,it will accept same http requests that would  be sent normally to JanusServer.

Sample config

::

 "janus_agent": {
	"enabled": false,                         // enables the Janus agent: <true|false>
	"url": "/janus",
	"sessions_conns": ["*internal"],
	"janus_conns": [{                         // instantiate connections to multiple Janus Servers
		"address": "127.0.0.1:8088",          // janus API address 
		"type": "*ws",                        // type of the transport to interact via janus API
		"admin_address": "localhost:7188",    // janus admin address used to retrive more information for sessions and handles
		"admin_password": "",                 // secret to pass restriction to communicate to the endpoint
	}],
	"request_processors": [],                 // request processors to be applied to Janus messages
},

Config params
^^^^^^^^^^^^^

Most of the parameters are explained in :ref:`JSON configuration <configuration>`, hence we mention here only the ones where additional info is necessary or there will be particular implementation for *JanusAgent*.

Software Installation
---------------------

 For detailed information on installing JanusServer on Debian, please refer to its official `repository  <https://github.com/meetecho/janus-gateway?tab=readme-ov-file#dependencies/>`_.

