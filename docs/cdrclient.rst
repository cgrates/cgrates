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

For the moment we support processing CDRs in the following formats:

CDR .CSV
--------

Most widely used format by Telecom Switches. 

Light to read and generic to process. 
CDRC should be able to process in this way any .csv CDR, independent of the Telecom Switch generating them. Incompatibilities here can come out of answer time and duration formats which can vary between CDR writer implementations. 
As answer time we support a number of formats already - rfc3339, SQL/MySQL, unix timestamp. As duration we support nanoseconds granularity in our code, however if time unit is not specified (eg: ms, s, m, h), we assume CDR duration will be in seconds.

CDR fields are extracted based on configured indexes in a file row (0 represents first field).

A particular configuration format it is represented by extra fields which need not only to be extracted by row index but also to be named since .csv format does not save field names/labels. CDRC uses the following convention for extra fields in the configuration: *<index_extrafield_1>:<label_extrafield_1>[,<index_extrafield_n>:<label_extrafield_n>]...*.


