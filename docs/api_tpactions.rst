ApierV1.SetTPActions
++++++++++++++++++

Creates a new Actions profile within a tariff plan.

**Request**:

 Data:
  ::

   type TPActions struct {
	TPid      string   // Tariff plan id
	ActionsId string   // Actions id
	Actions   []Action // Set of actions this Actions profile will perform
   }

   type Action struct {
	Identifier      string  // Identifier mapped in the code
	BalanceType     string  // Type of balance the action will operate on
	Direction       string  // Balance direction
	Units           float64 // Number of units to add/deduct
	ExpiryTime      string  // Time when the units will expire
	DestinationId   string  // Destination profile id
	RatingSubject   string  // Reference a rate subject defined in RatingProfiles
	BalanceWeight   float64 // Balance weight
	ExtraParameters string
	Weight          float64 // Action's weight
   }

 Mandatory parameters: ``[]string{"TPid", "ActionsId", "Actions", "Identifier", "Weight"}``

 *JSON sample*:
  ::

   {
    "id": 39,
    "method": "ApierV1.SetTPActions",
    "params": [
        {
            "Actions": [
                {
                    "BalanceType": "*monetary",
                    "BalanceWeight": 0,
                    "DestinationId": "*any",
                    "Direction": "*out",
                    "ExpiryTime": "0",
                    "Identifier": "*topup_reset",
                    "RatingSubject": "",
                    "Units": 10,
                    "Weight": 10
                }
            ],
            "ActionsId": "TOPUP_10",
            "TPid": "CGR_API_TESTS"
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
    "id": 39, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/ActionsId already present in StorDb.


ApierV1.GetTPActions
++++++++++++++++++

Queries specific Actions profile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPActions struct {
	TPid      string // Tariff plan id
	ActionsId string // Actions id
   }

 Mandatory parameters: ``[]string{"TPid", "ActionsId"}``

 *JSON sample*:
  ::

   {
    "id": 40,
    "method": "ApierV1.GetTPActions",
    "params": [
        {
            "ActionsId": "TOPUP_10",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }
 
**Reply**:

 Data:
  ::

   type TPActions struct {
	TPid      string   // Tariff plan id
	ActionsId string   // Actions id
	Actions   []Action // Set of actions this Actions profile will perform
   }

   type Action struct {
	Identifier      string  // Identifier mapped in the code
	BalanceType     string  // Type of balance the action will operate on
	Direction       string  // Balance direction
	Units           float64 // Number of units to add/deduct
	ExpiryTime      string  // Time when the units will expire
	DestinationId   string  // Destination profile id
	RatingSubject   string  // Reference a rate subject defined in RatingProfiles
	BalanceWeight   float64 // Balance weight
	ExtraParameters string
	Weight          float64 // Action's weight
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 40,
    "result": {
        "Actions": [
            {
                "BalanceType": "*monetary",
                "BalanceWeight": 0,
                "DestinationId": "*any",
                "Direction": "*out",
                "ExpiryTime": "0",
                "ExtraParameters": "",
                "Identifier": "*topup_reset",
                "RatingSubject": "",
                "Units": 10,
                "Weight": 10
            }
        ],
        "ActionsId": "TOPUP_10",
        "TPid": "CGR_API_TESTS"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested Actions profile not found.


ApierV1.GetTPActionIds
++++++++++++++++++++

Queries Actions identities on specific tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPActionIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 41,
    "method": "ApierV1.GetTPActionIds",
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
    "id": 41,
    "result": [
        "TOPUP_10"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There are no Actions profiles defined on the selected TPid.


