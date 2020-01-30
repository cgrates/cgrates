7. Suppliers Management
=======================

7.1 List Suppliers
------------------

:Hint:
    suppliers Tenant="cgrates.org" ID="SPP_1"

*Request*

::

    {
    	"method": "APIerSv1.GetSupplierProfile",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"ID": "SPP_1"
    	}],
    	"id": 6
    }

*Response*

::

    {
    	"id": 6,
    	"result": {
    		"Tenant": "cgrates.org",
    		"ID": "SPP_1",
    		"FilterIDs": ["FLTR_ACNT_dan", "FLTR_DST_DE"],
    		"ActivationInterval": {
    			"ActivationTime": "2017-07-29T15:00:00Z",
    			"ExpiryTime": "0001-01-01T00:00:00Z"
    		},
    		"Sorting": "*lowest_cost",
    		"SortingParams": [],
    		"Suppliers": [{
    			"ID": "supplier1",
    			"FilterIDs": ["FLTR_ACNT_dan"],
    			"AccountIDs": [],
    			"RatingPlanIDs": ["RPL_1"],
    			"ResourceIDs": ["ResGroup1"],
    			"StatIDs": ["Stat1"],
    			"Weight": 10
    		}],
    		"Blocker": false,
    		"Weight": 10
    	},
    	"error": null
    }
