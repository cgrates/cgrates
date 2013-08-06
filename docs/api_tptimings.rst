ApierV1.SetTPTiming
+++++++++++++++++

Creates a new timing within a tariff plan.

**Request**:

 Data:
  ::

   type ApierTPTiming struct {
	TPid      string // Tariff plan id
	TimingId  string // Timing id
	Years     string // semicolon separated list of years this timing is valid on, \*all supported
	Months    string // semicolon separated list of months this timing is valid on, \*none and \*all supported
	MonthDays string // semicolon separated list of month's days this timing is valid on, \*none and \*all supported
	WeekDays  string // semicolon separated list of week day names this timing is valid on \*none and \*all supported
	Time      string // String representing the time this timing starts on
   }

 Mandatory parameters: ``[]string{"TPid", "TimingId", "Years","Months","MonthDays", "WeekDays","Time"}``

 *JSON sample*:
  ::

   {
    "id": 3, 
    "method": "ApierV1.SetTPTiming", 
    "params": [
        {
            "MonthDays": "1;2;3;31", 
            "Months": "1;3;6", 
            "TPid": "SAMPLE_TP", 
            "Time": "13:00:00", 
            "TimingId": "SAMPLE_TIMING_5", 
            "WeekDays": "0", 
            "Years": "2013;2014"
        }
    ]
   }

**Reply**:

 Data:
  ::

   string

 Possible answers:
  ``OK`` - Success.

 *JSON sample*:
  ::

   {
    "error": null, 
    "id": 3, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/TimingId already exists in StorDb.


ApierV1.GetTPTiming
+++++++++++++++++

Queries specific Timing on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPTiming struct {
	TPid     string // Tariff plan id
	TimingId string // Timing id
   }

 Mandatory parameters: ``[]string{"TPid", "TimingId"}``

 *JSON sample*:
  ::

   {
    "id": 4, 
    "method": "ApierV1.GetTPTiming", 
    "params": [
        {
            "TPid": "SAMPLE_TP", 
            "TimingId": "SAMPLE_TIMING_7"
        }
    ]
   }

**Reply**:

 Data:
  ::

   type ApierTPTiming struct {
	TPid      string // Tariff plan id
	TimingId  string // Timing id
	Years     string // semicolon separated list of years this timing is valid on, \*all supported
	Months    string // semicolon separated list of months this timing is valid on, \*none and \*all supported
	MonthDays string // semicolon separated list of month's days this timing is valid on, \*none and \*all supported
	WeekDays  string // semicolon separated list of week day names this timing is valid on \*none and \*all supported
	Time      string // String representing the time this timing starts on
   }

 *JSON sample*:
  ::

   {
    "error": null, 
    "id": 4, 
    "result": {
        "MonthDays": "1;2;3;31", 
        "Months": "1;3;6", 
        "TPid": "SAMPLE_TP", 
        "Time": "13:00:00", 
        "TimingId": "SAMPLE_TIMING_7", 
        "WeekDays": "*all", 
        "Years": "2013;2014"
    }
  }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested timing id not found.


ApierV1.GetTPTimingIds
+++++++++++++++++++++++++

Queries timing identities on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPTimingIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 5, 
    "method": "ApierV1.GetTPTimingIds", 
    "params": [
        {
            "TPid": "SAMPLE_TP"
        }
    ]
   }

**Reply**:

 Data:
  ::

   []string

 *JSON sample*:
  ::

   {
    "error": null, 
    "id": 5, 
    "result": [
        "SAMPLE_TIMING_1", 
        "SAMPLE_TIMING_2", 
        "SAMPLE_TIMING_3", 
        "SAMPLE_TIMING_4", 
        "SAMPLE_TIMING_5"
    ]
   }


**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.


