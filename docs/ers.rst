
.. _MySQL: https://dev.mysql.com/
.. _PostgreSQL: https://www.postgresql.org/
.. _MSSQL: https://www.microsoft.com/en-us/sql-server/
.. _Kamailio: https://www.kamailio.org/w/
.. _OpenSIPS: https://opensips.org/
.. _Kafka_: https://kafka.apache.org/

.. EventReaderService:

EventReaderService
==================


**EventReaderService/ERs** is a subsystem designed to read events coming from external sources and convert them into internal ones. The converted events are then sent to other CGRateS subsystems, like *SessionS* where further processing logic is applied to them.

The translation between external and internal events is done based on field mapping, defined in :ref:`JSON configuration <configuration>`.


Configuration
-------------

The **EventReaderService** is configured within *ers* section  from :ref:`JSON configuration <configuration>`.


Sample config 
^^^^^^^^^^^^^

With explanations in the comments:

::

 "ers": {
	"enabled": true,					// enable the service
	"sessions_conns": ["*internal"],	// connection towards SessionS
	"readers": [						// list of active readers
		{
			"id": "file_reader2",		// file_reader2 reader
			"run_delay":  "-1",			// reading of events it is triggered outside of ERs
			"field_separator": ";",		// field separator definition
			"type": "*file_csv",		// type of reader, *file_csv can read .csv files
			"row_length" : 0,			// Number of fields from csv file
			"flags": [					// influence processing logic within CGRateS workflow
				"*cdrs",				//   *cdrs will create CDRs
				"*log"					//   *log will log the events to syslog
			],
			"source_path": "/tmp/ers2/in",		// location of the files
			"processed_path": "/tmp/ers2/out",	// move the files here once processed
			"fields":[					// mapping definition between line index in the file and CGRateS field 
				{
					"tag": "OriginID",			// OriginID together with OriginHost will 
					"path": "*cgreq.OriginID",	//   uniquely identify the session on CGRateS side
					"type": "*variable",
					"value": "~*req.0",q		// take the content from line index 0
					"mandatory": true			//   in the request file
				},
				{
					"tag": "RequestType",		// RequestType instructs SessionS
					"path": "*cgreq.RequestType",//   about charging type to apply for the event
					"type": "*variable",
					"value": "~*req.1",
					"mandatory": true
				},
				{
					"tag": "Category",			// Category serves for ataching Account
					"path": "*cgreq.Category",	//   and RatingProfile to the request
					"type": "*constant",
					"value": "call",
					"mandatory": true
				},
				{
					"tag": "Account",			// Account is required by charging
					"path": "*cgreq.Account",
					"type": "*variable",
					"value": "~*req.3",
					"mandatory": true
				},
				{
					"tag": "Subject",			// Subject is required by charging
					"path": "*cgreq.Subject",
					"type": "*variable",
					"value": "~*req.3",
					"mandatory": true
				},
				{
					"tag": "Destination",		// Destination is required by charging
					"path": "*cgreq.Destination",
					"type": "*variable",
					"value": "~*req.4:s/0([1-9]\\d+)/+49${1}/",
					"mandatory": true			// Additional mediation is performed on number format
				},
				{
					"tag": "AnswerTime",		// AnswerTime is required by charging
					"path": "*cgreq.AnswerTime",
					"type": "*variable",
					"value": "~*req.5",
					"mandatory": true
				},
				{
					"tag": "Usage",				// Usage is required by charging
					"path": "*cgreq.Usage",
					"type": "*variable",
					"value": "~*req.6",
					"mandatory": true
				},
				{
					"tag": "HDRExtra1",			// HDRExtra1 is transparently stored into CDR
					"path": "*cgreq.HDRExtra1",	//   as extra field not used by CGRateS
					"type": "*composed",
					"value": "~*req.6",
					"mandatory": true
				}
			],
		}
	]
 }


Config params
^^^^^^^^^^^^^

Most of the parameters are explained in :ref:`JSON configuration <configuration>`, hence we mention here only the ones where additional info is necessary or there will be particular implementation for *EventReaderService*.


readers
	List of reader profiles which ERs manages. Simultaneous readers of the same type are possible.

id
	Reader identificator, used mostly for debug. The id should be unique per each reader since it can influence updating configuration from different *.json* configuration.

type
	Reader type. Following types are implemented:

	**\*file_csv**
		Reader for *comma separated* files.

	**\*file_xml**
		Reader for *.xml* formatted files.

	**\*file_fwv**
		Reader for *fixed width value* formatted files.

	**\*kafka_json_map**
		Reader for hashmaps within Kafka_ database.

	**\*sql**
		Reader for generic content out of *SQL* databases. Supported databases are: MySQL_, PostgreSQL_ and MSSQL_.

run_delay
	Duration interval between consecutive reads from source. If 0 or less, *ERs* relies on external source (ie. Linux inotify for files) for starting the reading process.

concurrent_requests
	Limits the number of concurrent reads from source (ie: the number of simultaneously opened files).

source_path
	Path towards the events source

processed_path
	Optional path for moving the events source to after processing.

xml_root_path
	Used in case of XML content and will specify the prefix path applied to each xml element read.

tenant
	Will auto-populate the Tenant within the API calls sent to CGRateS. It has the form of a RSRParser. If undefined, default one from *general* section will be used.

timezone
	Defines the timezone for source content which does not carry that information. If undefined, default one from *general* section will be used.

filters
	List of filters to pass for the reader to process the event. For the dynamic content (prefixed with *~*) following special variables are available:

	**\*vars**
		Request related shared variables between processors, populated especially by core functions. The data put inthere is not automatically transfered into requests sent to CGRateS, unless instructed inside templates.

	**\*tmp**
		Temporary container to be used when exchanging information between fields.

	**\*req**
		Request read from the source. In case of file content without field name, the index will be passed instead of field source path.

	**\*hdr**
		Header values (available only in case of *\*file_fwv*). In case of file content without field name, the index will be passed instead of field source path.

	**\*trl**
		Trailer values (available only in case of *\*file_fwv*). In case of file content without field name, the index will be passed instead of field source path.

flags
	Special tags enforcing the actions/verbs done on an event. There are two types of flags: **main** and **auxiliary**. 

	There can be any number of flags or combination of those specified in the list however the flags have priority one against another and only some simultaneous combinations of *main* flags are possible. 

	The **main** flags will select mostly the action taken on a request.

	The **auxiliary** flags only make sense in combination with **main** ones. 

	Implemented **main** flags are (in order of priority, and not working simultaneously unless specified):

	**\*log**
		Logs the Event read. Can be used together with other *main* flags.

	**\*none**
		Disable transfering the Event from *Reader* to *CGRateS* side.

	**\*dryrun**
		Together with not transfering the Event on CGRateS side will also log it, useful for troubleshooting.

	**\*auth**
		Sends the Event for authorization on CGRateS.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts**, **\*routes**, **\*routes_ignore_errors**, **\*routes_event_cost**, **\*routes_maxcost** which are used to influence the auth behavior on CGRateS side. More info on that can be found on the **SessionS** component's API behavior.

	**\*initiate**
		Initiates a session out of Event on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts** which are used to influence the behavior on CGRateS side.

	**\*update**
		Updates a session with the Event on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*accounts** which are used to influence the behavior on CGRateS side.

	**\*terminate**
		Terminates a session using the Event on CGRateS side.

		Auxiliary flags available: **\*thresholds**, **\*stats**, **\*resources**, **\*accounts** which are used to influence the behavior on CGRateS side.

	**\*message**
		Process the Event as individual message charging on CGRateS side.

		Auxiliary flags available: **\*attributes**, **\*thresholds**, **\*stats**, **\*resources**, **\*accounts**, **\*routes**, **\*routes_ignore_errors**, **\*routes_event_cost**, **\*routes_maxcost** which are used to influence the behavior on CGRateS side.

	**\*event**
		Process the Event as generic event on CGRateS side.

		Auxiliary flags available: all flags supported by the "SessionSv1.ProcessEvent" generic API

	**\*cdrs**
		Build a CDR out of the Event on CGRateS side. Can be used simultaneously with other flags (except *\*dry_run)

path
	Defined within field, specifies the path where the value will be written. Possible values:

	**\*vars**
		Write the value in the special container, *\*vars*, available for the duration of the request.

	**\*cgreq**
		Write the value in the request object which will be sent to CGRateS side.

	**\*hdr**
		Header values (available only in case of *\*file_fwv*). In case of file content without field name, the index will be passed instead of field source path.

	**\*trl**
		Trailer values (available only in case of *\*file_fwv*). In case of file content without field name, the index will be passed instead of field source path.


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

	**\*usage_difference**
		Calculates the usage difference between two arguments passed in the *value*. Requires 2 arguments: *$stopTime;$startTime*

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
			Take data from current request coming from the reader.

		**\*vars**
			Take data from internal container labeled *\*vars*. This is valid for the duration of the request.

		**\*cgreq**
			Take data from the request being sent to :ref:`SessionS`. This is valid for one active request.

		**\*cgrep**
			Take data from the reply coming from :ref:`SessionS`. This is valid for one active reply.

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





