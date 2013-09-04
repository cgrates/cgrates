6.2. Rating logic
=================

Let's start with the most important function: finding the cost of a certain call. 

The call information comes to CGRateS having the following vital information like  subject, destination, start time and end time. The engine will look up the database for the rates applicable to the received subject and destination. 

::

 type CallDescriptor struct {
	Direction                             
	TOR                                   
	Tenant, Subject, Account, Destination 
	TimeStart, TimeEnd                    
	LoopIndex       // indicates the position of this segment in a cost request loop
	CallDuration    // the call duration so far (partial or final)
	FallbackSubject // the subject to check for destination if not found on primary subject
	ActivationPeriods
 }

When the session manager receives a call start event it will first check if the call is prepaid or postpaid. If the call is postpaid than the cost will be determined only once at the end of the call but if the call is prepaid there will be a debit operation every X seconds (X is configurable).

In prepaid case the rating engine will have to set rates for multiple parts of the call so the *LoopIndex* in the above structure will help the engine add the connect fee only to the first part. The *CallDuration* attribute is used to set the right rate in case the rates database has different costs for the different parts of a call e.g. first minute is more expensive (we can also define the minimum rate unit). 

The **FallbackSubject** is used in case the initial call subject is not found in the rating profiles list (more on this later in this chapter).


What are the activation periods?

    At one given time there is a set of prices that applay to different time intervals when a call can be made. In CGRateS one can define multiple such sets that will become active in various point of time called activation time. The activation period is a structure describing different prices for a call on different intervals of time. This structure has an activation time, which specifies the active prices for a period of time by one ore more (usually more than one) rate intervals. 

::

 type Interval struct {
	Years            
	Months           
	MonthDays        
	WeekDays         
	StartTime, EndTime 
	Weight, ConnectFee 
	Prices  
	RoundingMethod     
	RoundingDecimals   
 }

 type Price struct {
	GroupIntervalStart 
	Value              
	RateIncrement      
	RateUnit 
 }


An **Interval** specifies the Month, the MonthDay, the WeekDays, the StartTime and the EndTime when the Interval's price profile is in effect. 

:Example: The Interval {"Month": [1], "WeekDays":[1,2,3,4,5], "StartTime":"18:00:00"} specifies the *Price* for the first month of each year from Monday to Friday starting 18:00. Most structure elements are optional and they can be combined in any way it makes sense. If an element is omitted it means it is zero or any.

The *ConnectFee* specifies the connection price for the call if this interval is the first one of the call.

The *Weight* will establish which interval will set the price for a call segment if more then one applies to it. 

:Example: Let's assume there is an interval defining price for the weekdays and another interval that defines a special holiday rates. As that holiday is also one of the regular weekdays than both intervals are applicable to a call made on that day so the interval with the smaller Weight will give the price for the call in question. If both intervals have the same Weight than the interval with the smaller price wins. It is, however, a good practice to set the Weight for the defined intervals.

The *RoundingMethod* and the *RoundingDecimals* will adjust the price using the specified function and number of decimals (more on this in the rates definition chapter).

The **Price** structure defines the start (*GroupIntervalStart*) of a section of a call with a specified rate *Value* per *RateUnit* diving and rounding the section in *RateIncrement* subsections.

So when there is a need to define new sets of prices just define new ActivationPeriods with the activation time set to the moment when it becomes active.

Let's get back to the engine. When a GetCost or Debit call comes to the engine it will try to match the best rating profile for the given *Direction*, *Tenant*, *TOR* and *Subject* using the longest *Subject* prefix method or using the *FallbackSubject* if not found. The rating profile contains the activation periods that might apply to the call in question.

At this point in rating process the engine will start splitting the call into various time spans using the following criterias:

1. Minute Balances: first it will handle the call information to the originator user acount to be split by available minute balances. If the user has free or special price minutes for the call destination they will be consumed by the call.

2. Activation periods: if there were not enough special price minutes available than the engine will check if the call spans over multiple activation periods (the call starts in initial rates period and continues in another).

3. Intervals: for each activation period that apply to the call the engine will select the best rate intervals that apply. 

::

 type TimeSpan struct {
	TimeStart, TimeEnd
	Cost              
	ActivationPeriod  
	Interval          
	MinuteInfo        
	CallDuration  // the call duration so far till TimeEnd
 }


The result of this splitting will be a list of *TimeSpan* structures each having attached the MinuteInfo or the Interval that gave the price for it. The *CallDuration* attribute will select the right *Price* from the *Interval* *Prices* list. The final cost for the call will be the sum of the prices of these times spans plus the *ConnectionFee* from the first time span of the call.

6.2.1 User balances
-------------------

The user account contains a map of various balances like money, sms, internet traffic, internet time, etc. Each of these lists contains one or more Balance structure that have a wheight and a possible expiration date.  

::

 type UserBalance struct {
	Type           // prepaid-postpaid
	BalanceMap
	UnitCounters   
	ActionTriggers 
 }

 type Balance struct {
	Value          
	ExpirationDate 
	Weight 
 }

CGRateS treats special priced or free minutes different from the rest of balances. They will be called free minutes further on but they can have a special price.

The free minutes must be handled a little differently because usually they are grouped by specific destinations (e.g. national minutes, ore minutes in the same network). So they are grouped in balances and when a call is made the engine checks all applicable balances to consume minutes according to that call.

When a call cost needs to be debited these minute balances will be queried for call destination first. If the user has special minutes for the specific destination those minutes will be consumed according to call duration.

A standard debit operation consist of selecting a certaing balance type and taking all balances from that list in the weight order to be debited till the total amount is consumed.

CGRateS provide api for adding/substracting user's money credit. The prepaid and postpaid are uniformly treated except that the prepaid is checked to be always greater than zero and the postpaid can go bellow zero.

Both prepaid and postpaid can have a limited number of free SMS and Internet traffic per month and this budget is replenished at regular intervals based on the user tariff plan or as the user buys more free SMSs (for example).

Another special feature allows user to get a better price as the call volume increases each month. This can be added on one ore more thresholds so the more he/she talks the cheaper the calls.

Finally bonuses can be rewarded to users who received a certain volume of calls.
