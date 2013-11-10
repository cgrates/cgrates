CDR APIs
========

Set of APIs CDR related.


ApierV1.ExportCsvCdrs
---------------------

Used to request a new CDR export file. In can include specific interval for CDRs *answer_time*. Any of the two interval limits can be left unspecified hence resulting in the possibility to export complete database of CDRs with one API call.
 *NOTE*: Since CGRateS does not keep anywhere a history of exports, it becomes the responsibility of the system administrator to make sure that his exports are not doubled.


**Request**:

Data:

 ::

  type AttrExpCsvCdrs struct {
	TimeStart    string // If provided, will represent the starting of the CDRs interval (>=)
	TimeEnd      string // If provided, will represent the end of the CDRs interval (<)
   }

 Mandatory parameters: none

 *JSON sample*:
  ::

   {
    "id": 3,
    "method": "ApierV1.ExportCsvCdrs",
    "params": [
        {
            "TimeEnd": "1383823746"
        }
    ]
   }

**Reply**:

 Data:
  ::

   type ExportedCsvCdrs struct {
	ExportedFilePath          string // Full path to the newly generated export file
        NumberOfCdrs              int    // Number of CDRs in the export file
   }


 *JSON sample*:
  ::

   {
    "error": null,
    "id": 3,
    "result": {
        "ExportedFilePath": "/var/log/cgrates/cdr/out/cgr/csv/cdrs_1384104724.csv",
        "NumberOfCdrs": 2
    }
   }

**Errors**:

 ``SERVER_ERROR`` - Server error occurred.
