Api Calls
=========

JSON RPC
--------
GetCost
	Creates a CallCost structure with the cost information calculated for the received CallDescriptor.

DebitBalance
	Interface method used to add/substract an amount of cents from user's money budget.
	The amount filed has to be filled in call descriptor.

DebitSMS
	Interface method used to add/substract an amount of units from user's SMS budget.
	The amount filed has to be filled in call descriptor.

DebitSeconds
	Interface method used to add/substract an amount of seconds from user's minutes budget.
	The amount filed has to be filled in call descriptor.

GetMaxSessionTime
	Returns the approximate max allowed session for user budget. It will try the max amount received in the call descriptor 
	and will decrease it by 10% for nine times. So if the user has little credit it will still allow 10% of the initial amount.
	If the user has no credit then it will return 0.

AddVolumeDiscountSeconds
	Interface method used to add an amount to the accumulated placed call seconds to be used for volume discount.
	The amount filed has to be filled in call descriptor.

ResetVolumeDiscountSeconds
	Resets the accumulated volume discount seconds (to zero).

AddRecievedCallSeconds
	Adds the specified amount of seconds to the received call seconds. When the threshold specified in the user's tariff plan is reached then the received call budget is reseted and the bonus specified in the tariff plan is applied.
	The amount filed has to be filled in call descriptor.

ResetUserBudget
	Resets user budgets value to the amounts specified in the tariff plan.

HTTP
----

getcost
	:Example: curl "http://127.0.0.1:8000/getcost?cstmid=vdf&subj=rif&dest=0257"
debitbalance
	:Example: curl "http://127.0.0.1:8000/debitbalance?cstmid=vdf&subj=rif&dest=0257@amount=100"
debitsms
	:Example: curl "http://127.0.0.1:8000/debitsms?cstmid=vdf&subj=rif&dest=0257@amount=100"
debitseconds
	:Example: curl "http://127.0.0.1:8000/debitseconds?cstmid=vdf&subj=rif&dest=0257@amount=100"
getmaxsessiontime
	:Example: curl "http://127.0.0.1:8000/getmaxsessiontime?cstmid=vdf&subj=rif&dest=0257@amount=100"
addvolumediscountseconds
	:Example: curl "http://127.0.0.1:8000/addvolumediscountseconds?cstmid=vdf&subj=rif&dest=0257@amount=100"
resetvolumediscountseconds
	:Example: curl "http://127.0.0.1:8000/resetvolumediscountseconds?cstmid=vdf&subj=rif&dest=0257"
addrecievedcallseconds
	:Example: curl "http://127.0.0.1:8000/addrecievedcallseconds?cstmid=vdf&subj=rif&dest=0257@amount=100"
resetuserbudget
	:Example: curl "http://127.0.0.1:8000/resetuserbudget?cstmid=vdf&subj=rif&dest=0257"
