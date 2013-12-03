ApierV1.SetTPTiming
===================

Creates a new timing within a tariff plan.

**Request**:

 Data:
  ::

   type ApierTPTiming struct {
	TPid      string // Tariff plan id
	TimingId  string // Timing id
	Years     string // semicolon separated list of years this timing is valid on, \*any supported
	Months    string // semicolon separated list of months this timing is valid on, \*any supported
	MonthDays string // semicolon separated list of month's days this timing is valid on, \*any supported
	WeekDays  string // semicolon separated list of week day names this timing is valid on \*any supported
	Time      string // String representing the time this timing starts on
   }

 Mandatory parameters: ``[]string{"TPid", "TimingId", "Years", "Months", "MonthDays", "WeekDays", "Time"}``

 *JSON sample*:
  ::

   {
    "id": 3,
    "method": "ApierV1.SetTPTiming",
    "params": [
        {
            "MonthDays": "*any",
            "Months": "*any",
            "TPid": "TEST_SQL",
            "Time": "00:00:00",
            "TimingId": "ALWAYS",
            "WeekDays": "*any",
            "Years": "*any"
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


ApierV1.GetTPTiming
===================

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
    "id": 5,
    "method": "ApierV1.GetTPTiming",
    "params": [
        {
            "TPid": "TEST_SQL",
            "TimingId": "ALWAYS"
        }
    ]
   }
   

**Reply**:

 Data:
  ::

   type ApierTPTiming struct {
	TPid      string // Tariff plan id
	TimingId  string // Timing id
	Years     string // semicolon separated list of years this timing is valid on, \*any supported
	Months    string // semicolon separated list of months this timing is valid on, \*any supported
	MonthDays string // semicolon separated list of month's days this timing is valid on, \*any supported
	WeekDays  string // semicolon separated list of week day names this timing is valid on \*any supported
	Time      string // String representing the time this timing starts on
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 5,
    "result": {
        "MonthDays": "*any",
        "Months": "*any",
        "TPid": "TEST_SQL",
        "Time": "00:00:00",
        "TimingId": "ALWAYS2",
        "WeekDays": "*any",
        "Years": "*any"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested timing id not found.


ApierV1.GetTPTimingIds
======================

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
    "id": 4,
    "method": "ApierV1.GetTPTimingIds",
    "params": [
        {
            "TPid": "CGR_API_TESTS"
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
    "id": 4,
    "result": [
        "ASAP"
    ]
   }


**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.


