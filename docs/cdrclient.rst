CDR Client (cdrc) 
=================

It's role is to gather offline CDRs and post them to CDR Server(CDRS) component.

Part of the *cgr-engine*, can be started on a remote server as standalone component.

Controlled within *cdrc* section of the configuration file.

Has two modes of operation:

- Automated: CDR file processing is triggered on file creation/move.
- Periodic: CDR file processing will be triggered at configured time interval (delay/sleep between processes) and it will be performed on all files present in the folder (IN) at run time.

Principles behind functionality:

- Monitor/process a CDR folder (IN) as outlined above.
- For every file processed, extract the information based on configuration and post it via configured mechanism to CDRS.
- The fields extracted out of each CDR row are the same ones depicted in the CDRS documentation (following primary and extra fields concept).
- Once the file processing completes, move it in it's original format in another folder (OUT) in order to avoid re-processing. Here it's worth mentioning the auto-detection of duplicated CDRs at server side based on accid and host fields.
- Advanced configuration like forking a number of simultaneous client instances monitoring different folders possible through the use of *.xml* configuration.

Import Templates
----------------

To specifiy custom imports (for various sources) one can specify *Import Templates*. These are definable within both *.cfg* as well as *.xml* advanced configuration files.
For increased flexibility the Import Template can be defined using CGR-RSR fields capturing both ReGexp as well as static rules. The static values will be way faster in processing but limited in functionality.

CGR-RSR Regexp Rule
~~~~~~~~~~~~~~~~~~~

Format:
::

 ~field_id:s/regexp_search_and_capture_rule/output_teplate/

Example of usage:
::

 Input CDR field: 
 	{
 	"account": "First-Account123"
 	}
 Capture Rule:
 	~account:s/^*+(Account123)$/$1-processed/
 Result after processing:
 	{
 	"account": "Account123-processed"
 	}


CGR-RSR Static Rule
~~~~~~~~~~~~~~~~~~~

Format:
::

 ^field_id:static_value

Example of usage:
::

Input CDR field: 
 	{
 	"account": "First-Account123"
 	}
 Capture Rule:
 	^account:MasterAccount
 Result after processing:
 	{
 	"account": "MasterAccount"
 	}


CDR Formats supported:

CDR .CSV
--------

Most widely used format by Telecom Switches. 

Light to read and generic to process. 
CDRC should be able to process in this way any .csv CDR, independent of the Telecom Switch generating them. Incompatibilities here can come out of answer time and duration formats which can vary between CDR writer implementations. 
As answer time we support a number of formats already - rfc3339, SQL/MySQL, unix timestamp. As duration we support nanoseconds granularity in our code. Time unit can be specified (eg: ms, s, m, h), or if missing, will default to nanoseconds.

In case of *.csv* files the Import Template will contain indexes for the possition where primary fields are located (0 representing the first field) and fieldname/position format for extra fields which need not only to be extracted by row index but also to be named since .csv format does not save field names/labels. CDRC uses the following convention for extra fields in the configuration: *<label_extrafield_1>:<index_extrafield_1>[...,<label_extrafield_n>:<index_extrafield_n>]...*.
