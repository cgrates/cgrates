.. _Diameter: https://tools.ietf.org/html/rfc6733

.. _DiameterAgent:

DiameterAgent
=============

**DiameterAgent** translates between Diameter_ and **CGRateS**, sending *RPC* requests towards **CGRateS/SessionS** component and returning replies from it to the *DiameterClient*.

Implements Diameter_ protocol in a standard agnostic manner, giving users the ability to implement own interfaces by defining simple *processor templates* within the :ref:`configuration <engine_configuration>`  files.

Used mostly in modern mobile networks (LTE/xG).

The **DiameterAgent** is configured via *diameter_agent* section  within :ref:`configuration <engine_configuration>`.


Configuration
-------------


Sample config 
^^^^^^^^^^^^^

With explanations in the comments:

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


Config params
^^^^^^^^^^^^^

Most of the parameters are explained in :ref:`configuration <engine_configuration>`, hence we mention here only the ones where additional info is necessary or there will  be particular implementation for *DiameterAgent*.


**listen_net**
	The network the *DiameterAgent* will bind to. CGRateS supports both **tcp** and **sctp** specified in Diameter_ standard.

**concurrent_requests**
	The maximum number of active requests processed at one time by the *DiameterAgent*. When this number is reached, new inbound requests will be rejected with *DiameterError* code until the concurrent number drops bellow again. The default value of *-1* imposes no limits.

**asr_template**
	The template (out of templates config section) used to build the AbortSession message. If not specified the ASR message is never sent out.

**templates**
	Group fields based on their usability. Can be used in both processor templates as well as hardcoded within CGRateS functionality (ie *\*err* or *\*asr*). The IDs are unique, defining the same id in multiple configuration places/files will result into overwrite.

	*\*err*: is a hardcoded template used when *DiameterAgent* cannot parse the incoming message. Aside from logging the error via internal logger the message defined via *\*err* template will be sent out.

	*\*asr*: can be activated via *asr_template* config key to enable sending of *Diameter* *ASR* message to *DiameterClient*.

	*\*cca*: defined for convenience to follow the standard for the fields used in *Diameter* *CCA* messages.

**request_processors**
	List of processor profiles applied on request/replies. 

	Once a request processor will be matched (it's *filters* should match), the *request_fields* will be used to craft a request object and the flags will decide what sort of procesing logic will be applied to the crafted request. 

	After request processing, there will be a second part executed: reply. The reply object will be built based on the *reply_fields* section in the  
	request processor.

	Once the *reply_fields* are finished, the object converted and returned to the *DiameterClient*, unless *continue* flag is enabled in the processor, which makes the next request processor to be considered.


processor or field **filters**
	Will specify a list of filter rules which need to match in order for the processor to run (or field to be applied).

	For the dynamic content (prefixed with *~*) following special variables are available:

	* **\*vars**
		Request related shared variables between processors, populated especially by core functions. The data put inthere is not automatically transfered into requests sent to CGRateS, unless instructed inside templates. 

		Following vars are automatically set by core: 

		* **OriginHost**: agent configured *origin_host*
		* **OriginRealm**: agent configured *origin_realm*
		* **ProductName**: agent configured *product_name*
		* **\*app**: current request application name (out of diameter dictionary)
		* **\*appid**: current request application id (out of diameter dictionary)
		* **\*cmd**: current command short naming (out of diameter dictionary) plus *R" as suffix - ie: *CCR*
	
	* **\*req**
		Diameter request as it comes from the *DiameterClient*. 

		Special selector format defined in case of groups *\*req.Path.To.Attribute[$groupIndex]* or *\*req.Absolute.Path.To.Attribute[~AnotherAttributeRelativePath($valueAnotherAttribute)]*. 

		Example 1: *~\*req.Multiple-Services-Credit-Control.Rating-Group[1]* translates to: value of the group attribute at path Multiple-Services-Credit-Control.Rating-Group which is located in the second group (groups start at index 0).
		Example 2: *~\*req.Multiple-Services-Credit-Control.Used-Service-Unit.CC-Input-Octets[~Rating-Group(1)]* which translates to: value of the group attribute at path: *Multiple-Services-Credit-Control.Used-Service-Unit.CC-Input-Octets* where Multiple-Services-Credit-Control.Used-Service-Unit.Rating-Group has value of "1".

	* **\*cgreq**
		Request which was sent to CGRateS (mostly useful in replies).

	* **\*cgrep** 
		Reply coming from CGRateS.

	* **\*cgrareq**
		Active request in relation to CGRateS side. It can be used in both *request_fields*, referring to CGRRequest object being built, or in *reply_fields*, referring to CGRReply object.

processor **flags**
	Special tags enforcing the actions/verbs done on a request. There are two types of flags: **main** and **auxiliary**. 

	There can be any number of flags or combination of those specified in the list however the flags have priority one against another and only some simultaneous combinations of *main* flags are possible. 

	The **main** flags will select mostly the action taken on a request.

	The **auxiliary** flags only make sense in combination with **main** ones. 

	Implemented flags are (in order of priority, and not working in simultaneously unless specified):

	* **\*log**
		Logs the Diameter request/reply. Can be used together with other *main* actions.

	* **\*none**
		Disable transfering the request from *Diameter* to *CGRateS* side. Used mostly to pasively answer *Diameter* requests or troubleshoot (mostly in combination with *\*log* flag).

	* **\*dryrun**
		Together with not transfering the request on CGRateS side will also log the *Diameter* request/reply, useful for troubleshooting.

	* **\*auth**
		Sends the request for authorization on CGRateS.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts**, **\*suppliers**, **\*suppliers_ignore_errors**, **\*suppliers_event_cost** which are used to influence the auth behavior on CGRateS side. More info on that can be found on the **SessionS** component APIs behavior.

	* **\*initiate**
		Initiates a session out of request on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts** which are used to influence the auth behavior on CGRateS side.

	* **\*update**
		Updates a session with the request on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*accounts** which are used to influence the auth behavior on CGRateS side.

	* **\*terminate**
		Terminates a session using the request on CGRateS side.

		Auxiliary flags available: **\*thresholds**, **\*stats**, **\*resources**, **\*accounts** which are used to influence the auth behavior on CGRateS side.

	* **\*message**
		Process the request as individual message charging on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts**, **\*suppliers**, **\*suppliers_ignore_errors**, **\*suppliers_event_cost** which are used to influence the auth behavior on CGRateS side.


	* **\*event**
		Process the request as generic event on CGRateS side.

		Auxiliary flags available: all flags supported by the "SessionSv1.ProcessEvent" generic API

	* **\*cdrs**
		Build a CDR out of the request on CGRateS side. Can be used simultaneously with other flags (except *\*dry_run)


field **path**
	Specifies the path where the value will be written. Possible values:

	* **\*cgreq**
		Write the value in the request object which will be sent to CGRateS side.

	* **\*req**
		Write the value to request built by *DiameterAgent* to be sent out on *Diameter* side.

	* **\*rep**
		Write the value to reply going out on *Diameter* side.

field **type**
	Specifies the logic type to be used when writing the value of the field. Possible values:

	* **\*none**
		Pass

	* **\*filler**
		Fills the values with an empty string

	* **\*constant**
		Writes out a constant

	* **\*remote_host**
		Writes out the Address of the remote *DiameterClient* sending us the request

	* **\*variable**
		Writes out the variable value, overwriting previous one set

	* **\*composed**
		Writes out the variable value, postpending to previous value set

	* **\*usage_difference**
		Calculates the usage difference between two arguments passed in the *value*. Requires 2 arguments: *$stopTime;$startTime*

	* **\*cc_usage**
		Calculates the usage out of *CallControl* message. Requires 3 arguments: *$reqNumber;$usedCCTime;$debitInterval*

	* **\*sum**
		Calculates the sum of all arguments passed within *value*. It supports summing up duration, time, float, int autodetecting them in this order.

	* **\*difference**
		Calculates the difference between all arguments passed within *value*. Possible value types are (in this order): duration, time, float, int.

	* **\*value_exponent**
		Calculates the exponent of a value. It requires two values: *$val;$exp*