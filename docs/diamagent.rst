.. _Diameter: https://tools.ietf.org/html/rfc6733

.. _DiameterAgent:

DiameterAgent
=============

**DiameterAgent** translates between Diameter_ and **CGRateS**, sending *RPC* requests towards **CGRateS/SessionS** component and returning replies from it to the *DiameterClient*.

Implements Diameter_ protocol in a standard agnostic manner, giving users the ability to implement own interfaces by defining simple *processor templates* within the :ref:`JSON configuration <configuration>`  files.

Used mostly in modern mobile networks (LTE/xG).


Configuration
-------------

The **DiameterAgent** is configured within *diameter_agent* section from :ref:`JSON configuration <configuration>`.


Sample config
^^^^^^^^^^^^^

With explanations in the comments:

::

 "diameter_agent": {
	"enabled": false,					// enables the diameter agent: <true|false>
	"listen": "127.0.0.1:3868",			// address where to listen for diameter requests <x.y.z.y/x1.y1.z1.y1:1234>
	"listen_net": "tcp",				// transport type for diameter <tcp|sctp>
	"dictionaries_path": "/usr/share/cgrates/diameter/dict/",	// path towards directory
										//   holding additional dictionaries to load
	"sessions_conns": ["*internal"],	// connection towards SessionS
	"origin_host": "CGR-DA",			// diameter Origin-Host AVP used in replies
	"origin_realm": "cgrates.org",		// diameter Origin-Realm AVP used in replies
	"vendor_id": 0,						// diameter Vendor-Id AVP used in replies
	"product_name": "CGRateS",			// diameter Product-Name AVP used in replies
	"synced_conn_requests": false,		// process one request at the time per connection
	"asr_template": "*asr",				// enable AbortSession message being sent to client
	"request_processors": [		// decision logic for message processing
		{
			"id": "SMSes",		// id is used for debug in logs (ie: using *log flag)
			"filters": [		// list of filters to be applied on message for this processor to run
				"*string:~*vars.*cmd:CCR",
				"*string:~*req.CC-Request-Type:4",
				"*string:~*req.Service-Context-Id:LPP"
			],
			"flags": ["*event", "*accounts", "*cdrs"],	// influence processing logic within CGRateS workflow
			"request_fields":[							// data exchanged between Diameter and CGRateS
				{
					"tag": "ToR",			// tag is used in debug, 
					"path": "*cgreq.ToR",	// path is the field on CGRateS side
					"type": "*constant",	// type defines the method to provide the value
					"value": "*sms"}		
				{
					"tag": "OriginID",			// OriginID will identify uniquely 
					"path": "*cgreq.OriginID",	// the session on CGRateS side
					"type": "*variable",		// it's value will be taken from Diameter AVP:
					"mandatory": true,			// Multiple-Services-Credit-Control.Service-Identifier
					"value": "~*req.Multiple-Services-Credit-Control.Service-Identifier"
				},
				{
					"tag": "OriginHost",		// OriginHost combined with OriginID 
					"path": "*cgreq.OriginHost",// is used by CGRateS to build the CGRID
					"mandatory": true,
					"type": "*variable",		// have the value out of special variable: *vars
					"value": "*vars.OriginHost"
				},
				{
					"tag": "RequestType",			// RequestType instructs SessionS 
					"path": "*cgreq.RequestType",	//  about charging type to apply for the event
					"type": "*constant",
					"value": "*prepaid"
				},
				{
					"tag": "Category",			// Category serves for ataching Account
					"path": "*cgreq.Category",	//   and RatingProfile to the request
					"type": "*constant",
					"value": "sms"
				},
				{
					"tag": "Account",			// Account is required by charging
					"path": "*cgreq.Account",
					"type": "*variable",		// value is taken dynamically from a group AVP
					"mandatory": true,			//   where Subscription-Id-Type is 0
					"value": "~*req.Subscription-Id.Subscription-Id-Data<~Subscription-Id-Type(0)>" 
				},
				{
					"tag": "Destination",			// Destination is used for charging
					"path": "*cgreq.Destination",	// value from Diameter will be mediated before sent to CGRateS
					"type": "*variable",
					"mandatory": true,
					"value": "~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:s/^\\+49(\\d+)/int${1}/:s/^0049(\\d+)/int${1}/:s/^49(\\d+)/int${1}/:s/^00(\\d+)/+${1}/:s/^[\\+]?(\\d+)/int${1}/:s/int(\\d+)/+49${1}/"
				},
				{
					"tag": "Destination",		// Second Destination will overwrite the first if filter matches
					"path": "*cgreq.Destination",
					"filters":[					// Only overwrite when filters are matching
						"*notprefix:~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:49",
						"*notprefix:~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:3312"
					],
					"type": "*variable", 
					"mandatory": true,
					"value": "~*req.Service-Information.SMS-Information.Recipient-Info.Recipient-Address.Address-Data:s/^[\\+]?(\\d+)/int${1}/:s/int(\\d+)/+00${1}/"
				},
				{
					"tag": "SetupTime",			// SetupTime is used by charging
					"path": "*cgreq.SetupTime",
					"type": "*variable",
					"value": "~*req.Event-Timestamp",
					"mandatory": true
				},
				{
					"tag": "AnswerTime",		// AnswerTime is used by charging
					"path": "*cgreq.AnswerTime",
					"type": "*variable",
					"mandatory": true,
					"value": "~*req.Event-Timestamp"
				},
				{
					"tag": "Usage",			// Usage is used by charging
					"path": "*cgreq.Usage",				
					"type": "*variable",
					"mandatory": true,
					"value": "~*req.Multiple-Services-Credit-Control.Requested-Service-Unit.CC-Service-Specific-Units"
				},
				{
					"tag": "Originator-SCCP-Address",			// Originator-SCCP-Address is an extra field which we want in CDR
					"path": "*cgreq.Originator-SCCP-Address",	// not used by CGRateS
					"type": "*variable", "mandatory": true,
					"value": "~*req.Service-Information.SMS-Information.Originator-SCCP-Address"
				},
			],
			"reply_fields":[			// fields which are sent back to DiameterClient
				{
					"tag": "CCATemplate",	// inject complete Template defined as *cca above
					"type": "*template",
					"value": "*cca"
				},
				{
					"tag": "ResultCode",  	// Change the ResultCode if the reply received from CGRateS contains a 0 MaxUsage
					"filters": ["*eq:~*cgrep.MaxUsage:0"],
					"path": "*rep.Result-Code", 
					"blocker": true,		// do not consider further fields if this one is processed
					"type": "*constant",
					"value": "4012"},
				{"tag": "ResultCode",		// Change the ResultCode AVP if there was an error received from CGRateS
					"filters": ["*notempty:~*cgrep.Error:"],
					"path": "*rep.Result-Code",
					"blocker": true,
					"type": "*constant",
					"value": "5030"}
			]
		}

	]
		},
		
	],
 },


Config params
^^^^^^^^^^^^^

Most of the parameters are explained in :ref:`JSON configuration <configuration>`, hence we mention here only the ones where additional info is necessary or there will be particular implementation for *DiameterAgent*.


listen_net
	The network the *DiameterAgent* will bind to. CGRateS supports both **tcp** and **sctp** specified in Diameter_ standard.

asr_template
	The template (out of templates config section) used to build the AbortSession message. If not specified the ASR message is never sent out.

templates
	Group fields based on their usability. Can be used in both processor templates as well as hardcoded within CGRateS functionality (ie *\*err* or *\*asr*). The IDs are unique, defining the same id in multiple configuration places/files will result into overwrite.

	**\*err**
		Is a hardcoded template used when *DiameterAgent* cannot parse the incoming message. Aside from logging the error via internal logger the message defined via *\*err* template will be sent out.

	**\*asr**
		Can be activated via *asr_template* config key to enable sending of *Diameter* *ASR* message to *DiameterClient*.

	**\*cca**
		Defined for convenience to follow the standard for the fields used in *Diameter* *CCA* messages.

request_processors
	List of processor profiles applied on request/replies. 

	Once a request processor will be matched (it's *filters* should match), the *request_fields* will be used to craft a request object and the flags will decide what sort of procesing logic will be applied to the crafted request. 

	After request processing, there will be a second part executed: reply. The reply object will be built based on the *reply_fields* section in the  
	request processor.

	Once the *reply_fields* are finished, the object converted and returned to the *DiameterClient*, unless *continue* flag is enabled in the processor, which makes the next request processor to be considered.


filters
	Will specify a list of filter rules which need to match in order for the processor to run (or field to be applied).

	For the dynamic content (prefixed with *~*) following special variables are available:

	**\*vars**
		Request related shared variables between processors, populated especially by core functions. The data put inthere is not automatically transfered into requests sent to CGRateS, unless instructed inside templates. 

		Following vars are automatically set by core: 

		* **OriginHost**: agent configured *origin_host*
		* **OriginRealm**: agent configured *origin_realm*
		* **ProductName**: agent configured *product_name*
		* **RemoteHost**: the Address of the remote client
		* **\*app**: current request application name (out of diameter dictionary)
		* **\*appid**: current request application id (out of diameter dictionary)
		* **\*cmd**: current command short naming (out of diameter dictionary) plus *R" as suffix - ie: *CCR*
	
	**\*req**
		Diameter request as it comes from the *DiameterClient*. 

		Special selector format defined in case of groups *\*req.Path.To.Attribute[$groupIndex]* or *\*req.Absolute.Path.To.Attribute<~AnotherAttributeRelativePath($valueAnotherAttribute)>*. 

		Example 1: *~\*req.Multiple-Services-Credit-Control.Rating-Group<1>* translates to: value of the group attribute at path Multiple-Services-Credit-Control.Rating-Group which is located in the second group (groups start at index 0).
		Example 2: *~\*req.Multiple-Services-Credit-Control.Used-Service-Unit.CC-Input-Octets<~Rating-Group(1)>* which translates to: value of the group attribute at path: *Multiple-Services-Credit-Control.Used-Service-Unit.CC-Input-Octets* where Multiple-Services-Credit-Control.Used-Service-Unit.Rating-Group has value of "1".

	**\*rep**
		Diameter reply going to *DiameterClient*. 

	**\*cgreq**
		Request sent to CGRateS.

	**\*cgrep** 
		Reply coming from CGRateS.

	**\*diamreq**
		Diameter request generated by CGRateS (ie: *ASR*).

flags
	Found within processors, special tags enforcing the actions/verbs done on a request. There are two types of flags: **main** and **auxiliary**. 

	There can be any number of flags or combination of those specified in the list however the flags have priority one against another and only some simultaneous combinations of *main* flags are possible. 

	The **main** flags will select mostly the action taken on a request.

	The **auxiliary** flags only make sense in combination with **main** ones. 

	Implemented **main** flags are (in order of priority, and not working simultaneously unless specified):

	**\*log**
		Logs the Diameter request/reply. Can be used together with other *main* actions.

	**\*none**
		Disable transfering the request from *Diameter* to *CGRateS* side. Used mostly to pasively answer *Diameter* requests or troubleshoot (mostly in combination with *\*log* flag).

	**\*dryrun**
		Together with not transfering the request on CGRateS side will also log the *Diameter* request/reply, useful for troubleshooting.

	**\*auth**
		Sends the request for authorization on CGRateS.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts**, **\*routes**, **\*routes_ignore_errors**, **\*routes_event_cost**, **\*routes_maxcost** which are used to influence the auth behavior on CGRateS side. More info on that can be found on the **SessionS** component's API behavior.

	**\*initiate**
		Initiates a session out of request on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts** which are used to influence the auth behavior on CGRateS side.

	**\*update**
		Updates a session with the request on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*accounts** which are used to influence the behavior on CGRateS side.

	**\*terminate**
		Terminates a session using the request on CGRateS side.

		Auxiliary flags available: **\*thresholds**, **\*stats**, **\*resources**, **\*accounts** which are used to influence the behavior on CGRateS side.

	**\*message**
		Process the request as individual message charging on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts**, **\*routes**, **\*routes_ignore_errors**, **\*routes_event_cost**, **\*routes_maxcost** which are used to influence the behavior on CGRateS side.


	**\*event**
		Process the request as generic event on CGRateS side.

		Auxiliary flags available: all flags supported by the "SessionSv1.ProcessEvent" generic API

	**\*cdrs**
		Build a CDR out of the request on CGRateS side. Can be used simultaneously with other flags (except **\*dryrun**)


path
	Defined within field, specifies the path where the value will be written. Possible values:

	**\*vars**
		Write the value in the special container, *\*vars*, available for the duration of the request.

	**\*cgreq**
		Write the value in the request object which will be sent to CGRateS side.

	**\*cgrep**
		Write the value in the reply returned by CGRateS.

	**\*rep**
		Write the value to reply going out on *Diameter* side.

	**\*diamreq**
		Write the value to request built by *DiameterAgent* to be sent out on *Diameter* side.

type
	Defined within field, specifies the logic type to be used when writing the value of the field. Possible values:

	**\*none**
		Pass

	**\*filler**
		Fills the values with an empty string

	**\*constant**
		Writes out a constant

	**\*variable**
		Writes out the variable value, overwriting previous one set

	**\*composed**
		Writes out the variable value, postpending to previous value set

	**\*group**
		Writes out the variable value, postpending to the list of variables with the same path

	**\*usage_difference**
		Calculates the usage difference between two arguments passed in the *value*. Requires 2 arguments: *$stopTime;$startTime*

	**\*cc_usage**
		Calculates the usage out of *CallControl* message. Requires 3 arguments: *$reqNumber;$usedCCTime;$debitInterval*

	**\*sum**
		Calculates the sum of all arguments passed within *value*. It supports summing up duration, time, float, int autodetecting them in this order.

	**\*difference**
		Calculates the difference between all arguments passed within *value*. Possible value types are (in this order): duration, time, float, int.

	**\*value_exponent**
		Calculates the exponent of a value. It requires two values: *$val;$exp*

	**\*template**
		Specifies a template of fields to be injected here. Value should be one of the template ids defined.

value
	The captured value. Possible prefixes for dynamic values are:

		**\*req**
			Take data from current request coming from diameter client.

		**\*vars**
			Take data from internal container labeled *\*vars*. This is valid for the duration of the request.

		**\*cgreq**
			Take data from the request being sent to :ref:`SessionS`. This is valid for one active request.

		**\*cgrep**
			Take data from the reply coming from :ref:`SessionS`. This is valid for one active reply.

		**\*diamreq**
			Take data from the diameter request being sent to the client (ie: *ASR*). This is valid for one active reply.

		**\*rep**
			Take data from the diameter reply being sent to the client.

mandatory
	Makes sure that the field cannot have empty value (errors otherwise).

tag
	Used for debug purposes in logs.

width
	Used to control the formatting, enforcing the final value to a specific number of characters.

strip
	Used when the value is higher than *width* allows it, specifying the strip strategy. Possible values are:

	**\*right**
		Strip the suffix.

	**\*xright**
		Strip the suffix, postpending one *x* character to mark the stripping.

	**\*left**
		Strip the prefix.

	**\*xleft**
		Strip the prefix, prepending one *x* character to mark the stripping.

padding
	Used to control the formatting. Applied when the data is smaller than the *width*. Possible values are:

	**\*right**
		Suffix with spaces.

	**\*left**
		Prefix with spaces.

	**\*zeroleft**
		Prefix with *0* chars.
