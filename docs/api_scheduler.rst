Scheduler APIs
==============

Set of scheduler related APIs.


ApierV1.ReloadScheduler
-----------------------

When called CGRateS will reorder/reschedule tasks based on data available in dataDb. This command is necessary after each data load, in some cases being automated in the administration tools (eg: inside *cgr-loader*)

**Request**:

Data:

 ::

  string

 Mandatory parameters: none

 *JSON sample*:
  ::

   {
    "id": 0,
    "method": "ApierV1.ReloadScheduler",
    "params": [
        ""
    ]
   }


**Reply**:

 Data:
  ::

   string

 Possible answers: **OK**

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 0,
    "result": "OK"
   }

**Errors**:

 ``SERVER_ERROR`` - Server error occurred.
