CDR Exporter
============

Component to retrieve rated CDRs from internal CDRs database. 

Although nowadays is custom to read a storage/database with tools, we do not recommend doing it so due to possibility that reads can slow down complete rating system. For this purpose we have created exporter plugins which are meant to work in tight relationship with CGRateS internal components in order to best optimize performance and avoid system locks.

For the moment we support exporting CDRs over the following interfaces:


CGR-CSV 
-------

Simplest way to export CDRs in a format internally defined (with parts like *CDRExtraFields* configurable in main configuration file).

Principles behind exports:

- Exports are to be manually requested (although automated is planned for the future via built in scheduled actions) via exposed JSON-RPC api. Example of api call from python call provided as sample script:

 ::

  rpc.call("ApierV1.ExportCsvCdrs",{"TimeStart":"1383823746","TimeEnd":"1383833746"} )

- On each export call there will be a .csv format file generated using **,** as separator. Location of the export folder is definable inside *cgrates.cfg*.
- File name of the export will appear like: *cdrs_$(timestamp).csv* where $(timestamp) will be replaced by unix timestamp of the server running the export process.
- Each exported file will have as content all the CDRs inside time interval defined in the API call. Both TimeStart and TimeEnd are optional, hence being able to obtain a full export of the CDRs with one API call.
- To be noted here that CGRateS does not keep anywhere a history of exports, hence it is the responsibility of the system administrator to make sure that his exports are not doubled.
- Each line within the exported file will follow an internally predefined template:

 ::
   
 $(cgrid),$(accid),$(cdrhost),$(reqtype),$(direction),$(tenant),$(tor),$(account),$(subject),$(destination),$(answer_time),$(duration),$(cost),$(extra_cdr_fields)

 The significance of the fields exported:
   - cgrid: unique identifier for one CDR within CGRateS records
   - accid: represents the unique accounting id given by the switch generating the CDR
   - cdrhost: represents the ip of the host generating the CDR
   - reqtype: matching the supported request types by the CGRateS
   - direction: matching the supported direction identifiers of the CGRateS
   - tenant: tenant whom this call belongs
   - tor: TypeOfRecord for the CDR
   - account: account id (accounting subsystem) the record should be attached to
   - subject: rating subject (rating subsystem) this call should be attached to
   - destination: destination to be charged
   - answer_time: time of the record (in case of tor=call this would be answer time of the call). This will arrive as either unix timestamp or datetime RFC3339 compatible.
   - duration: used in case of tor=call like, representing the total duration of the call
   - extra_cdr_fields:
   - selected list of cdr_extra fields via *cgrates.cfg* configuration or
   - alphabetical order of the cdr extra fields stored in cdr_extra table

