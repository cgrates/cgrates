CDR Server
==========

An important component of every rating system is represented by the CDR Server. CGRateS includes an out of the box CDR Server component, controlable in the configuration file and supporting multiple interfaces for CDR feeds. This component makes the CDRs real-time accessible (raported to the time of receiving them) to CGRateS subsystems.

For the moment we support receiving CDRs over the following interfaces:


CDR-CGR 
-------

Available as handler within http server, it represents the lightest and fastest way to get CDRs inside CGRateS in real-time.

To feed CDRs in via this interface, one must use url of the form: <http://$ip_configured:$port_configured/cgr>.

The CDR fields are received via http form (although for simplicity we support inserting them within query parameters as well) and are expected to be urlencoded in order to transport special characters reliably. All fields are expected by CGRateS as string, particular conversions being done on processing each CDR.
The fields received are splitt into two different categories based on CGRateS interest in them:

Primary fields: the fields which CGRateS needs for it's own operations and are stored into cdrs_primary table of storDb.

- accid: represents the unique accounting id given by the switch generating the CDR
- cdrhost: represents the ip of the host generating the CDR
- reqtype: matching the supported request types by the CGRateS
- direction: matching the supported direction identifiers of the CGRateS
- tenant: tenant whom this call belongs
- tor: TypeOfRecord for the CDR
- account: account id (accounting subsystem) the record should be attached to
- subject: rating subject (rating subsystem) this call should be attached to
- destination: destination to be charged
- time_answer: time of the record (in case of tor=call this would be answer time of the call). This will arive as either unix timestamp or datetime RFC3339 compatible.
- duration: used in case of tor=call like, representing the total duration of the call

Extra fields: any field coming in via the http request and not a member of primary fields list. These fields are stored as json encoded into *cdrs_extra* table of storDb.


CDR-FS_JSON 
-----------

Available as handler within http server, it implements the mechanism to store CDRs received from FreeSWITCH mod_json_cdr.

This interface is available at url:  <http://$ip_configured:$port_configured/freeswitch_json>.

This handler has a different implementation logic than the previous CDR-CGR, filtering fields received in the CDR from FreeSWITCH based on predefined configuration.
The mechanism of extracting CDR information out of JSON encoded CDR received from FreeSWITCH is the following:

- When receiving the CDR from FreeSWITCH, CGRateS will extract the content of ''variables'' object.
- Content of the ''variables'' will be filtered out and the following information will be stored into an internal CDR object:
   - Fields used by CGRateS in primary mediation, known as primary fields. These are:
      - uuid: internally generated uuid by FreeSWITCH for the call
      - sip_local_network_addr: IP address of the FreeSWITCH box generating the CDR
      - sip_call_id: call id out of SIP protocol
      - cgr_reqtype: request type as understood by the CGRateS
      - cgr_tor: TypeOfRecord (optional)
      - cgr_tenant: tenant this call belongs to (optional)
      - cgr_account: account id in CGRateS (optional)
      - cgr_subject: rating subject in CGRateS (optional)
      - cgr_destination: destination being rated (optional)
      - user_name: username as seen by FreeSWITCH (considered if cgr_subject or cgr_account not present)
      - dialed_extension: destination number considered if cgr_destination is missing
   - Fields stored at request in cdr_extra and definable in configuration file under *extra_fields*.
- Once the content will be filtered, the real CDR object will be processed, stored into storDb under *cdrs_primary* and *cdrs_extra* tables and, if configured, it will be passed further for mediation.

