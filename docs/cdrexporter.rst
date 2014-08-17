CDR Exporter
============

Component to retrieve rated CDRs from internal CDRs database. 

Although nowadays it is custom to read a storage/database with tools, we do not recommend doing it so due to possibility that reads can slow down complete rating system. For this purpose we have created exporter plugins which are meant to work in tight relationship with CGRateS internal components in order to best optimize performance and avoid system locks.


Export Templates
----------------

For advanced needs CGRateS Export Templates are configurable via *.cfg*, *.xml* as well as directly within RPC call requesting the export to be performed.
Inside each Export Template one can either specify simple CDR field ids or use CGR-RSR fields capturing both Regexp as well as static rules.

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


Export interfaces implemented:


CGR-CSV 
-------

Simplest way to export CDRs in a format internally defined (with parts like *CDRExtraFields* configurable in main configuration file).

Principles behind exports:

- Exports are to be manually requested (although automated is planned for the future through the used of built-in scheduled actions) via exposed JSON-RPC api. Example of api call from python call provided as sample script:

 ::

  rpc.call("ApierV1.ExportCsvCdrs",{"TimeStart":"1383823746","TimeEnd":"1383833746"} )

- On each export call there will be a .csv format file generated using configured separator. Location of the export folder is definable inside *cgrates.cfg*.
- File name of the export will appear like: *cdrs_$(timestamp).csv* where $(timestamp) will be replaced by unix timestamp of the server running the export process or requested via API call.
- Each exported file will have as content all the CDRs inside time interval defined in the API call. Both TimeStart and TimeEnd are optional, hence being able to obtain a full export of the available CDRs with one API call.
- To be noted here that CGRateS does not keep anywhere a history of exports, hence it is the responsibility of the system administrator to make sure that his exports are not doubled.
- If not otherwise defined, each line within the exported file will follow an internally predefined template:
cgrid,mediation_runid,tor,accid,reqtype,direction,tenant,category,account,subject,destination,setup_time,answer_time,usage,cost
 ::
   
 $(cgrid),$(mediation_runid),$(tor),$(accid),$(reqtype),$(direction),$(direction),$(tenant),$(category),$(account),$(subject),$(destination),$(setup_time),$(answer_time),$(usage),$(cost)

 The significance of the fields exported:
   - tor: type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms>
   - accid: represents the unique accounting id given by the telecom switch generating the CDR
   - cdrhost: represents the IP address of the host generating the CDR (automatically populated by the server)
   - cdrsource: formally identifies the source of the CDR (free form field)
   - reqtype: matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
   - direction: matching the supported direction identifiers of the CGRateS <*out>
   - tenant: tenant whom this record belongs
   - category: free-form filter for this record, matching the category defined in rating profiles.
   - account: account id (accounting subsystem) the record should be attached to
   - subject: rating subject (rating subsystem) this record should be attached to
   - destination: destination to be charged
   - setup_time: set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
   - answer_time: answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
   - usage: event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
   - extra_cdr_fields:
      - selected list of cdr_extra fields via *cgrates.cfg* configuration or
      - alphabetical order of the cdr extra fields stored in cdr_extra table


Sample CDR export file content which was made available at path: */var/log/cgrates/cdr/out/cgr/csv/cdrs_1384104724.csv*
::

 dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,rated,*out,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.0100


CGR-FWV 
-------

Fixed width form of export CDR. Advanced template configuration available via *.xml* configuration file.


Hybrid CSV-FWV
--------------

For advanced needs **CGRateS** supports exporting the CDRs as combination between *.csv* and *.fwv* formats.