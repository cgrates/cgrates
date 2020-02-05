
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

The translation between external and internal events is done based on field mapping, defined in :ref:`json configuration <engine_configuration>`.


Configuration
-------------

The **EventReaderService** is configured within *ers* section  from :ref:`JSON configuration <engine_configuration>`.


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
			"flags": [					// influence processing logic within CGRateS workflow
				"*cdrs",				//   *cdrs will create CDRs
				"*log"					//   *log will log the events to syslog
			],
			"source_path": "/tmp/ers2/in",		// location of the files
			"processed_path": "/tmp/ers2/out",	// move the files here once processed
			"content_fields":[					// mapping definition between line index in the file and CGRateS field 
				{
					"tag": "OriginID",			// OriginID together with OriginHost will 
					"path": "OriginID",		//   uniquely identify the session on CGRateS side
					"type": "*variable",
					"value": "~*req.0",q		// take the content from line index 0
					"mandatory": true			//   in the request file
				},
				{
					"tag": "RequestType",		// RequestType instructs SessionS
					"path": "RequestType",	//   about charging type to apply for the event
					"type": "*variable",
					"value": "~*req.1",
					"mandatory": true
				},
				{
					"tag": "Category",			// Category serves for ataching Account
					"path": "Category",		//   and RatingProfile to the request
					"type": "*constant",
					"value": "call",
					"mandatory": true
				},
				{
					"tag": "Account",			// Account is required by charging
					"path": "Account",
					"type": "*variable",
					"value": "~*req.3",
					"mandatory": true
				},
				{
					"tag": "Subject",			// Subject is required by charging
					"path": "Subject",
					"type": "*variable",
					"value": "~*req.3",
					"mandatory": true
				},
				{
					"tag": "Destination",		// Destination is required by charging
					"path": "Destination",
					"type": "*variable",
					"value": "~*req.4:s/0([1-9]\\d+)/+49${1}/",
					"mandatory": true			// Additional mediation is performed on number format
				},
				{
					"tag": "AnswerTime",		// AnswerTime is required by charging
					"path": "AnswerTime",
					"type": "*variable",
					"value": "~*req.5",
					"mandatory": true
				},
				{
					"tag": "Usage",				// Usage is required by charging
					"path": "Usage",
					"type": "*variable",
					"value": "~*req.6",
					"mandatory": true
				},
				{
					"tag": "HDRExtra1",			// HDRExtra1 is transparently stored into CDR
					"path": "HDRExtra1",	//   as extra field not used by CGRateS
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

Most of the parameters are explained in :ref:`JSON configuration <engine_configuration>`, hence we mention here only the ones where additional info is necessary or there will be particular implementation for *EventReaderService*.


readers
	List of reader profiles which ERs manages. Simultaneous readers of the same type are possible.

id
	Reader identificator, used mostly for debug. The id should be unique per each reader since it can influence updating configuration from different *.json* configuration.

type
	Reader type. Following types are implemented:

	**\*file_csv**
		Reader for *comma separated* files.

	**\*partial_csv**
		Reader for *comma separated* where content spans over multiple files.

	**\*flatstore**
		Reader for Kamailio_/OpenSIPS_ *db_flatstore* files.

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
	Will auto-populate the Tenant within the API calls sent to CGRateS. It has the form of a RSRField. If undefined, default one from *general* section will be used.

timezone
	Defines the timezone for source content which does not carry that information. If undefined, default one from *general* section will be used.

filters
	List of filters to pass for the reader to process the event. In case of file content without field name, the index will be passed instead of field source path.

flags
	Special tags enforcing the actions/verbs done on an event. There are two types of flags: **main** and **auxiliary**. 

	There can be any number of flags or combination of those specified in the list however the flags have priority one against another and only some simultaneous combinations of *main* flags are possible. 

	The **main** flags will select mostly the action taken on a request.

	The **auxiliary** flags only make sense in combination with **main** ones. 

	Implemented flags are (in order of priority, and not working simultaneously unless specified):

	**\*none**
		Disable transfering the request from *Reader* to *CGRateS* side.






