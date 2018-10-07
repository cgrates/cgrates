CDR Server
==========

An important component of every rating system is represented by the CDR Server. CGRateS includes an out of the box CDR Server component, controlable in the configuration file and supporting multiple interfaces for CDR feeds. This component makes the CDRs real-time accessible (influenced by the time of receiving them) to CGRateS subsystems.

Following interfaces are supported:


CDR-CGR 
-------

Available as handler within http server.

To feed CDRs in via this interface, one must use url of the form: <http://$ip_configured:$port_configured/cdr_http>.

The CDR fields are received via http form (although for simplicity we support inserting them within query parameters as well) and are expected to be urlencoded in order to transport special characters reliably. All fields are expected by CGRateS as string, particular conversions being done on processing each CDR.
The fields received are split into two different categories based on CGRateS interest in them:

Primary fields: the fields which CGRateS needs for it's own operations and are stored into cdrs_primary table of storDb.


- ToR: type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms>
- OriginID: represents the unique accounting id given by the telecom switch generating the CDR
- OrderID: Stor order id used as export order id
- OriginHost: represents the IP address of the host generating the CDR (automatically populated by the server)
- Source: formally identifies the source of the CDR (free form field)
- RequestType: matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
- Category: free-form filter for this record, matching the category defined in rating profiles.
- Tenant: tenant whom this record belongs
- Account: account id (accounting subsystem) the record should be attached to
- Subject: rating subject (rating subsystem) this record should be attached to
- Destination: destination to be charged
- SetupTime: set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
- AnswerTime: answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
- Usage: event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
- CostSource: The source of this cost
- Rated: Mark the CDR as rated so we do not process it during rating

Extra fields: any field coming in via the http request and not a member of primary fields list. These fields are stored as json encoded into *cdrs_extra* table of storDb.

Example of sample CDR generated simply using curl:
::

 curl --data "ToR=*voice \
  &Source=curl_cdr \
  &OrderID=abcde \
  &OriginHost=192.168.1.2 \
  &Source=sbc1 \
  &OriginID=qwerty3234567 \
  &ToR=*voice \
  &RequestType=*raw \
  &Tenant=192.168.56.66 \
  &Category=call \
  &Account=1004 \
  &Subject=1004 \
  &Destination=%2B4986517174963 \
  &SetupTime=2018-05-21T12:32:50Z \
  &AnswerTime=2018-05-21T12:32:56Z \
  &Usage=306 \
  &CostSource=*cdrs" http://127.0.0.1:2080/cdr_http



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
      - cgr_category: call category (optional)
      - cgr_tenant: tenant this call belongs to (optional)
      - cgr_account: account id in CGRateS (optional)
      - cgr_subject: rating subject in CGRateS (optional)
      - cgr_destination: destination being rated (optional)
      - user_name: username as seen by FreeSWITCH (considered if cgr_subject or cgr_account not present)
      - dialed_extension: destination number considered if cgr_destination is missing
   - Fields stored at request in cdr_extra and definable in configuration file under *extra_fields*.
- Once the content will be filtered, the real CDR object will be processed, stored into storDb under *cdrs_primary* and *cdrs_extra* tables and, if configured, it will be passed further for mediation.


CDR-RPC 
-------

Available as RPC handler on top of CGR APIs exposed (in-process as well as GOB-RPC and JSON-RPC). This interface is used for example by CGR-SM component capturing the CDRs over event interface (eg: OpenSIPS or FreeSWITCH-ZeroConfig scenario)

The RPC function signature looks like this:
::

 CDRSV1.ProcessCdr(cdr *utils.StoredCdr, reply *string) error


The simplified StoredCdr object is represented by following:
::

 type StoredCdr struct {
   CgrId          string
   OrderId        int64             // Stor order id used as export order id
   TOR            string            // type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms>
   AccId          string            // represents the unique accounting id given by the telecom switch generating the CDR
   CdrHost        string            // represents the IP address of the host generating the CDR (automatically populated by the server)
   CdrSource      string            // formally identifies the source of the CDR (free form field)
   ReqType        string            // matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
   Direction      string            // matching the supported direction identifiers of the CGRateS <*out>
   Tenant         string            // tenant whom this record belongs
   Category       string            // free-form filter for this record, matching the category defined in rating profiles.
   Account        string            // account id (accounting subsystem) the record should be attached to
   Subject        string            // rating subject (rating subsystem) this record should be attached to
   Destination    string            // destination to be charged
   SetupTime      time.Time         // set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
   AnswerTime     time.Time         // answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
   Usage          time.Duration     // event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
   ExtraFields    map[string]string // Extra fields to be stored in CDR
 }

