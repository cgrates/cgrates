.. _AnalyzerS:

AnalyzerS
=========

**AnalyzerS** is a service component part of the **CGRateS** infrastructure, designed to capture and index API interactions, enabling subsequent querying and analysis of API activity. It operates asynchronously without interfering with the normal API processing pipeline, with captured entries stored on disk and removed based on the configured *ttl* and *cleanup_interval*.

Complete interaction with **AnalyzerS** is possible via `CGRateS RPC APIs <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.


Processing logic
----------------

When an API call is processed by **CGRateS**, **AnalyzerS** captures and indexes the interaction without requiring additional configuration per API. Each captured record stores the full request and response, along with metadata including source, destination, encoding, duration, timestamp and any reply errors.

Each record is uniquely identified by a combination of encoding, source, destination, method and timestamp, so interactions can be precisely located within the index.

Entries are retained based on the configured *ttl* and cleaned up periodically. A cleanup pass also runs at service startup to remove any expired entries left from previous runs.


Parameters
----------

It is configured within the **analyzers** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

db_path
	Path to the folder where the API capture information is stored.

index_type
	The type of index used for storage. Possible values: <\*scorch|\*boltdb|\*leveldb|\*mossdb|\*internal>.

ttl
	Time to wait before removing the API capture.

cleanup_interval
	The interval at which the database is cleaned of expired entries.


Query API
---------

AnalyzerSv1.StringQuery
^^^^^^^^^^^^^^^^^^^^^^^

The primary RPC method for querying captured API interactions. Accepts an optional set of filters and returns matching records.

HeaderFilters
	Filters applied to the top-level metadata fields of the captured record. Supported fields include *RequestMethod*, *RequestSource*, *RequestDestination*, *RequestEncoding*, *RequestDuration* and *RequestStartTime*. Filter expressions follow `Bleve's query string syntax <http://blevesearch.com/docs/Query-String-Query/>`_. This filter type is recommended for performance-sensitive queries.

Limit
	Maximum number of results to return. If not set or 0, all matching records are returned.

Offset
	Number of records to skip before returning results.

ContentFilters
	Filters applied to the payload content of captured interactions. They follow the **CGRateS** inline filter syntax, using the same filtering available across other **CGRateS** subsystems, and support the following data path prefixes:

	- *\*req* — filtering on request parameter fields
	- *\*rep* — filtering on reply fields
	- *\*opts* — filtering on APIOpts fields
	- *\*hdr* — filtering on header fields

	ContentFilters are slower than HeaderFilters. Combining both is recommended when filtering on both metadata and payload content.


Captured record structure
-------------------------

Each captured interaction is stored as a record containing the following fields:

RequestID
	Identifier of the API request.

RequestMethod
	The API method that was called.

RequestParams
	The full set of parameters sent with the request.

RequestSource
	The address from which the request originated.

RequestDestination
	The address to which the request was sent.

RequestEncoding
	The encoding used for the API call.

RequestStartTime
	The time at which the request was received.

RequestDuration
	The total time taken to process the request.

Reply
	The full response returned by the API.

ReplyError
	The error returned by the API, if any. Null if the request was successful.


Use cases
---------

* Debugging and tracing API interactions across the system.
* Keeping track of API activity per tenant for audit purposes.
* Monitoring request duration across API methods for performance analysis.
* Verifying API requests and responses during development and testing.