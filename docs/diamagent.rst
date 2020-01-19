.. _Diameter: https://tools.ietf.org/html/rfc6733

.. _DiameterAgent:

DiameterAgent
=============

**DiameterAgent** translates between Diameter_ and **CGRateS**, sending *RPC* requests towards **SessionS** component and returning replies from it to the *DiameterClient*.

Implements Diameter_ protocol in a standard agnostic manner, giving users the ability to implement own interfaces by defining simple *processor templates* within the :ref:`configuration <engine_configuration>`  files.

Used mostly in modern mobile networks (LTE/xG).


Configuration
-------------

The **DiameterAgent** is configured via *diameter_agent* section  within :ref:`configuration <engine_configuration>`.


Sample config (explanation in the comments):

::

 "diameter_agent": {
	"enabled": false,											// enables the diameter agent: <true|false>
	"listen": "127.0.0.1:3868",									// address where to listen for diameter requests <x.y.z.y/x1.y1.z1.y1:1234>
	"listen_net": "tcp",										// transport type for diameter <tcp|sctp>
	"dictionaries_path": "/usr/share/cgrates/diameter/dict/",	// path towards directory holding additional dictionaries to load
	"sessions_conns": ["*internal"],
	"origin_host": "CGR-DA",									// diameter Origin-Host AVP used in replies
	"origin_realm": "cgrates.org",								// diameter Origin-Realm AVP used in replies
	"vendor_id": 0,												// diameter Vendor-Id AVP used in replies
	"product_name": "CGRateS",									// diameter Product-Name AVP used in replies
	"concurrent_requests": -1,									// limit the number of active requests processed by the server <-1|0-n>
	"synced_conn_requests": false,								// process one request at the time per connection
	"asr_template": "*asr",										// enable AbortSession message being sent to client 
																// forcing session disconnection from CGRateS side

	"templates":{												// message templates which can be injected within request/replies
		"*err": [
				{"tag": "SessionId", "field_id": "Session-Id", "type": "*composed",
					"value": "~*req.Session-Id", "mandatory": true},
				{"tag": "OriginHost", "field_id": "Origin-Host", "type": "*composed",
					"value": "~*vars.OriginHost", "mandatory": true},
				{"tag": "OriginRealm", "field_id": "Origin-Realm", "type": "*composed",
					"value": "~*vars.OriginRealm", "mandatory": true},
		],
		"*cca": [
				{"tag": "SessionId", "field_id": "Session-Id", "type": "*composed",
					"value": "~*req.Session-Id", "mandatory": true},
				{"tag": "ResultCode", "field_id": "Result-Code", "type": "*constant",
					"value": "2001"},
				{"tag": "OriginHost", "field_id": "Origin-Host", "type": "*composed",
					"value": "~*vars.OriginHost", "mandatory": true},
				{"tag": "OriginRealm", "field_id": "Origin-Realm", "type": "*composed",
					"value": "~*vars.OriginRealm", "mandatory": true},
				{"tag": "AuthApplicationId", "field_id": "Auth-Application-Id", "type": "*composed",
					 "value": "~*vars.*appid", "mandatory": true},
				{"tag": "CCRequestType", "field_id": "CC-Request-Type", "type": "*composed",
					"value": "~*req.CC-Request-Type", "mandatory": true},
				{"tag": "CCRequestNumber", "field_id": "CC-Request-Number", "type": "*composed",
					"value": "~*req.CC-Request-Number", "mandatory": true},
		],
		"*asr": [
				{"tag": "SessionId", "field_id": "Session-Id", "type": "*variable",
					"value": "~*req.Session-Id", "mandatory": true},
				{"tag": "OriginHost", "field_id": "Origin-Host", "type": "*variable",
					"value": "~*req.Destination-Host", "mandatory": true},
				{"tag": "OriginRealm", "field_id": "Origin-Realm", "type": "*variable",
					"value": "~*req.Destination-Realm", "mandatory": true},
				{"tag": "DestinationRealm", "field_id": "Destination-Realm", "type": "*variable",
					"value": "~*req.Origin-Realm", "mandatory": true},
				{"tag": "DestinationHost", "field_id": "Destination-Host", "type": "*variable",
					"value": "~*req.Origin-Host", "mandatory": true},
				{"tag": "AuthApplicationId", "field_id": "Auth-Application-Id", "type": "*variable",
					 "value": "~*vars.*appid", "mandatory": true},
				{"tag": "UserName", "field_id": "User-Name", "type": "*variable",
					"value": "~*req.User-Name", "mandatory": true},
				{"tag": "OriginStateID", "field_id": "Origin-State-Id", "type": "*constant",
					"value": "1"},
		]
	},
	"request_processors": [ 									// decision logic for message processing
		{
			"id": "SMSes", 										// id is used for debug in logs (ie: using *log flag)
			"filters": [										// list of filters to be applied on message for this processor to run
				"*string:~*vars.*cmd:CCR",
				"*string:~*req.CC-Request-Type:4",
				"*string:~*req.Service-Context-Id:LPP"
			],
			"flags": ["*event", "*accounts", "*cdrs"],			// influence processing logic within CGRateS workflow
			"request_fields":[									// data exchanged between Diameter and CGRateS
				{
					"tag": "TOR", "field_id": "ToR", 			// tag is used in debug, field_id is the field on CGRateS side
					"type": "*constant", "value": "*sms"}		// type defines the method to provide the value
				{
					"tag": "OriginID", "field_id": "OriginID",	// OriginID will identify uniquely the session on CGRateS side
					"type": "*variable", "mandatory": true,		// it's value will be taken from Diameter AVP:
					"value": "~*req.Multiple-Services-Credit-Control.Service-Identifier"// Multiple-Services-Credit-Control.Service-Identifier 
				},
				{
					"tag": "OriginHost", "field_id": "OriginHost",	// OriginHost combined with OriginID is used by CGRateS to build the CGRID
					"mandatory": true, "type": "*constant", "value": "0.0.0.0"
				},
				{
					"tag": "RequestType", "field_id": "RequestType",// RequestType tells SessionS which charging type to apply for the event
					"type": "*constant", "value": "*prepaid"
				},
				{
					"tag": "Category", "field_id": "Category",		// Category serves for ataching Account and RatingProfile to the request
					"type": "*constant", "value": "sms"
				},
				{
					"tag": "Account", "field_id": "Account",		// Account serves for ataching Account and RatingProfile to the request
					"type": "*variable", "mandatory": true,			// value is taken from a groupped AVP (
					"value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]" // where Subscription-Id-Type is 0)
				},
				{
					"tag": "Destination", "field_id": "Destination",	// Destination is used for charging
					"type": "*variable", "mandatory": true,				// value from Diameter will be mediated before sent to CGRateS
					"value": "~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:s/^\\+49(\\d+)/int${1}/:s/^0049(\\d+)/int${1}/:s/^49(\\d+)/int${1}/:s/^00(\\d+)/+${1}/:s/^[\\+]?(\\d+)/int${1}/:s/int(\\d+)/+49${1}/"
				},
				{
					"tag": "Destination", "field_id": "Destination",	// Second Destination will overwrite the first in specific cases
					"filters":[											// Only overwrite when filters are matching
						"*notprefix:~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:49",
						"*notprefix:~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:3958"
					],
					"type": "*variable", "mandatory": true,
					"value": "~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:s/^[\\+]?(\\d+)/int${1}/:s/int(\\d+)/+00${1}/"},
				{
					"tag": "SetupTime", "field_id": "SetupTime",		// SetupTime is used by charging
					"type": "*variable",
					"value": "~*req.Event-Timestamp", "mandatory": true
				},
				{
					"tag": "AnswerTime", "field_id": "AnswerTime",		// AnswerTime is used by charging
					"type": "*variable", , "mandatory": true, "value": "~*req.Event-Timestamp"
				},
				{
					"tag": "Usage", "field_id": "Usage",				// Usage is used by charging
					"type": "*variable", "mandatory": true,
					"value": "~*req.Multiple-Services-Credit-Control.Requested-Service-Unit.CC-Service-Specific-Units"
				},
				{
					"tag": "Originator-SCCP-Address",		// Originator-SCCP-Address is an extra field which we want in CDR
					"field_id": "Originator-SCCP-Address",	// not used by CGRateS
					"type": "*variable", "mandatory": true, 
					"value": "~*req.Service-Information.SMS-Information.Originator-SCCP-Address"
				},
			],
			"reply_fields":[			// fields which are sent back to DiameterClient
				{
					"tag": "CCATemplate", 					// inject complete Template defined as *cca above
					"type": "*template", "value": "*cca"
				},
				{
					"tag": "ResultCode",  						// Change the ResultCode if the reply received from CGRateS contains a 0 MaxUsage
					"filters": ["*eq:~*cgrep.MaxUsage:0"],
					"field_id": "Result-Code", "blocker": true,	// do not consider further fields if this one is processed
					"type": "*constant", "value": "4012"},
				{"tag": "ResultCode",							// Change the ResultCode AVP if there was an error received from CGRateS
					"filters": ["*notempty:~*cgrep.Error:"],
					"field_id": "Result-Code", "blocker": true,
					"type": "*constant", "value": "5030"}
			]
		}

	]
		},
		
	],
 },



