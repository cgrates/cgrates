DerivedCharging
===============

DerivedCharging is the process of forking original request into a number (configured) of emulated ones, derived from the original parameters. This mechanism used in combination with multi-tenancy supported by default by **CGRateS** can give out complex charging scenarios, needed for example in case of whitelabel-ing.

DerivedCharging occurs in two separate places:

- SessionManager: necessary to handle each derived (emulated) session in it's individuall loop (eg: individual resellers will have their own charging policies implemented, some paying per minute, others per second and so on) and keep them in sync (eg: one reseller is left out of money, original call should be disconnected and all emulated sessions should end their debit loops).
- Mediator: necessary to fork the CDRs into a number of derived ones influenced by the derived charging configuration and rate them individually.

Configuration
-------------

DerivedCharging is configured in two places:

- Platform level configured within *cgrates.cfg* file.
- Account level configured as part of TarrifPlans defition or interactively via RPC methods.

One DerivedCharger object will be configured by an internal object like:
::

 type DerivedCharger struct {
	RunId               string      // Unique runId in the chain
	RunFilters          string      // Only run the charger if all the filters match
	ReqTypeField        string      // Field containing request type info, number in case of csv source, '^' as prefix in case of static values
	DirectionField      string      // Field containing direction info
	TenantField         string      // Field containing tenant info
	CategoryField       string      // Field containing tor info
	AccountField        string      // Field containing account information
	SubjectField        string      // Field containing subject information
	DestinationField    string      // Field containing destination information
	SetupTimeField      string      // Field containing setup time information
	AnswerTimeField     string      // Field containing answer time information
	UsageField          string      // Field containing usage information
 }

**CGRateS** is able to attach an unlimited number of DerivedChargers to a single request, based on configuration.