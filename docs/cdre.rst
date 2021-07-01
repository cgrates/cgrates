.. _AMQP: https://www.amqp.org/
.. _SQS: https://aws.amazon.com/de/sqs/
.. _S3: https://aws.amazon.com/de/s3/
.. _Kafka: https://kafka.apache.org/


.. _CDRe:

CDRe
====


**CDRe** is an extension of :ref:`CDRs`, responsible for exporting the *CDR* events processed by :ref:`CDRs`. It is accessed via `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_ and configured within *cdre* section inside :ref:`JSON configuration <configuration>`.


Export types
------------

There are two types of exports with common configuration but different data sources:


Online exports
^^^^^^^^^^^^^^

Are real-time exports, triggered by the CDR event processed by :ref:`CDRs`, and take these events as data source. 

The *online exports* are enabled via *online_cdr_exports* :ref:`JSON configuration <configuration>` option within *cdrs*. 

You can control the templates which are to be executed via the filters which are applied for each export template individually.


Offline exports
^^^^^^^^^^^^^^^

Are exports which are triggered via `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_ and they have as data source the CDRs stored within *StorDB*.



Parameters
----------

CDRe
^^^^

**CDRe** it is configured within **cdre** section from :ref:`JSON configuration <configuration>`.

One **export profile** includes the following parameters:

export_format
	Specify the type of export which will run. Possible values are:

	**\*fileCSV**
		Exports into a comma separated file format.

	**\*fileFWV**
		Exports into a fixed width file format.

	**\*httpPost**
		Will post the CDR to a HTTP server. The export content will be a HTTP form encoded representation of the `internal CDR object <https://godoc.org/github.com/cgrates/cgrates/engine#CDR>`_.

	**\*http_json_cdr**
		Will post the CDR to a HTTP server. The export content will be a JSON serialized representation of the `internal CDR object <https://godoc.org/github.com/cgrates/cgrates/engine#CDR>`_.

	**\*httpJSONMap**
		Will post the CDR to a HTTP server. The export content will be a JSON serialized hmap with fields defined within the *fields* section of the template.

	**\*amqp_json_cdr**
		Will post the CDR to an AMQP_ queue. The export content will be a JSON serialized representation of the `internal CDR object <https://godoc.org/github.com/cgrates/cgrates/engine#CDR>`_. Uses AMQP_ protocol version 0.9.1.

	**\*amqpJSONMap**
		Will post the CDR to an AMQP_ queue. The export content will be a JSON serialized hmap with fields defined within the *fields* section of the template. Uses AMQP_ protocol version 1.0.

	**\*amqpv1JSONMap**
		Will post the CDR to an AMQP_ queue. The export content will be a JSON serialized hmap with fields defined within the *fields* section of the template. Uses AMQP_ protocol version 1.0.

	**\*sqsJSONMap**
		Will post the CDR to an `Amazon SQS queue <SQS>`_. The export content will be a JSON serialized hmap with fields defined within the *fields* section of the template.

	**\*s3JSONMap**
		Will post the CDR to `Amazon S3 storage <S3>`_. The export content will be a JSON serialized hmap with fields defined within the *fields* section of the template.

	**\*kafkaJSONMap**
		Will post the CDR to an `Apache Kafka <Kafka>`_. The export content will be a JSON serialized hmap with fields defined within the *fields* section of the template.

export_path
	Specify the export path. It has special format depending of the export type.

	**\*fileCSV**, **\*fileFWV**
		Standard unix-like filesystem path.

	**\*httpPost**, **\*http_json_cdr**, **\*httpJSONMap**
		Full HTTP URL

	**\*amqpJSONMap**, **\*amqpv1JSONMap**
		AMQP URL with extra parameters. 

		Sample: *amqp://guest:guest@localhost:5672/?queue_id=cgrates_cdrs&exchange=exchangename&exchange_type=fanout&routing_key=cgr_cdrs*

	**\*sqsJSONMap**
		SQS URL with extra parameters.

		Sample: *http://sqs.eu-west-2.amazonaws.com/?aws_region=eu-west-2&aws_key=testkey&aws_secret=testsecret&queue_id=cgrates-cdrs*

	**\*s3JSONMap**
		S3 URL with extra parameters.

		Sample: *http://s3.us-east-2.amazonaws.com/?aws_region=eu-west-2&aws_key=testkey&aws_secret=testsecret&queue_id=cgrates-cdrs*

	**\*kafkaJSONMap**
		Kafka URL with extra parameters.

		Sample: *localhost:9092?topic=cgrates_cdrs*


filters
	List of filters to pass for the export profile to execute. For the dynamic content (prefixed with *~*) following special variables are available:

	**\*req**
		The *CDR* event itself.

	**\*ec**
		The *EventCost* object with subpaths for all of it's nested objects.

tenant
	Tenant owning the template. It will be used mostly to match inside :ref:`FilterS`.

synchronous
	Block further exports until this one finishes. In case of *false* the control will be given to the next export template as soon as this one was started.

attempts
	Number of attempts before giving up on the export and writing the failed request to file. The failed request will be written to *failed_posts_dir* defined in *general* section.

field_separator
	Field separator to be used in some export types (ie. *\*fileCSV*).

attributes_context
	The context used when sending the CDR event to :ref:`AttributeS` for modifications. If empty, there will be no event sent to :ref:`AttributeS`.

fields
	List of fields for the exported event. Not affecting templates like *\*http_json_cdr* or *\*amqp_json_cdr* with fixed content.


One **field template** will contain the following parameters:

path
	Path for the exported content. Possible prefixes here are:

	*\*exp*
		Reference to the exported record.

	*\*hdr*
		Reference to the header content. Available in case of **\*fileCSV** and **\*fileFWV** export types.

	*\*trl*
		Reference to the trailer content. Available in case of **\*fileCSV** and **\*fileFWV** export types.

type
	The field type will give out the logic for generating the value. Values used depend on the type of prefix used in path.

	For *\*exp*, following field types are implemented:

	**\*variable**
		Writes out the variable value, overwriting previous one set.

	**\*composed**
		Writes out the variable value, postpending to previous value set

	**\*filler**
		Fills the values with a fixed lentgh string. 

	**\*constant**
		Writes out a constant

	**\*datetime**
		Parses the value as datetime and reformats based on the *layout* attribute.

	**\*combimed**
		Writes out a combined mediation considering events with the same *CGRID*.

	**\*masked_destination**
		Masks the destination using *\** as suffix. Matches the destination field against the list defined via *mask_destinationd_id* field.

	**\*httpPost**
		Uses a HTTP server as datasource for the value exported.

	For *\*hdr* and *\*trl*, following field types are possible:

	**\*filler**
		Fills the values with a string.

	**\*constant**
		Writes out a constant

	**\*handler**
		Will obtain the content via a handler. This works in tandem with the attribute *handler_id*.

value
	The exported value. Works in tandem with *type* attribute. Possible prefixes for dynamic values:

	**\*req**
		Data is taken from the current request coming from the *CDRs* component.

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

mask_destinationd_id
	The destinations profile where we match the *masked_destinations*.

hander_id
	The identifier of the handler to be executed in case of *\*handler* *type*.








