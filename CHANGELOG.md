# Changelog


## [0.10.5~dev] - 2024-09-24

## [0.10.4] - 2024-09-24

### Bug fixes

- [Storage] Fixed mongo URL builder adding incorrect square brackets for multiple hosts, restoring cluster functionality. #4209
- [RALs] Corrected ConnectFee handling to avoid deduction from inactive (not expired) balances. #4479
- [RALs] Fixed balance blocker functionality to work beyond just ConnectFee. #4476
- [Indexes] Fixed performance issue when removing filter indexes due to too many database trips. #4362 (Related to #4357)
- [RPCClient] Fixed failover on RPC timeout errors. #4413
- [StatS] Ensured expired metrics are removed before retrieval (not only during ProcessEvent requests). #4436
- [ThresholdS] Fixed MinSleep parameter being reset on every hit regardless of Snooze field status. #4330
- [RadiusAgent] Fixed detection condition for *radReplyCode fields during RADIUS reply parsing. #4119
- [ERs] Fixed slice out of bounds panics caused by xml_root_path being shorter than dynamic *req template values. #4145
- [ERs] Fixed "expr expression is nil" errors by returning '.' instead of empty string when parsing an empty relative path. #4153
- [CDRs] Ensured refund flag is not ignored when rerate is true. #4410
- [CDRs] Ensured compatibility with external agents by adding aliases to ExtraFields, addressing non-backwards compatible changes. #4364
- [EventCost] Fixed field access for Rating/ExtraCharges fields. #4314
- [SchedulerS] Fixed cron expression handling of single digit 0 values, resolving issues with scheduled actions execution timing. #4388
- [Storage] Fixed RemTpData method for consistent db key building, preventing occasional failed deletions. #4356
- [Version] Fixed commit date parsing for Git versions 2.45+, caused by backwards incompatible changes to iso-strict date output formats. #4353
- [cgr-engine] Fixed misleading error messages when requesting help (-h flag). #4494
- [General] Replaced invalid struct tags for tp models. #4076
- [General] Silence harmless _goRPC_.Cancel error. #4408
- [General] Fixed go vet warnings. #4455

### Maintenance

- [Version] Incremented version
- [Dependencies] Updated all dependencies to latest backwards compatible versions. #4464

Full Changelog: https://github.com/cgrates/cgrates/compare/v0.10.3...v0.10

## [0.10.3] - 2023-08-09

- [RALs] Now, balance update events from RALs to ThresholdS (when negative) are only sent once.
- [SessionS] Updated to use rals_conns when refund rounding is sent.
- [SupplierS] Now requires a connection to rals for calculating AccountIDs and RatingPlanIDs.
- [SessionS] Implemented the compilation of SRun.EventCost before storing and passing it further.
- [ApierS] Improved error handling for APIerSv1.GetActionTriggers.
- [SessionS] Added condition to assess if increment should be considered roundIncrement.
- [SessionS] When appending to the EventCost,the charging interval is now being cloned.
- [FilterS] Enhanced automated index fields matching for optimization.
- [AgentS] Introduced \*routes_maxcost flag.
- [SessionS] max_call_duration config replaced with default_usage per ToR.
- [SessionS] If replication_conns are set, sessions will not terminate on shutdown.
- [EventCost] Improved FieldAsInterface function to prevent crashes when a RatingPlan doesn't exist (#2743).
- [EventCost] Added nil check when creating EventCost DataProvider, preventing crashes when cgr-engine is manually restarted during an ongoing call (#2764).
- [DispatcherS] Fixed panic when sending Ping request through DispatcherS.
- [CacheS] Tenant now passed to automatic cache calls (#2928).
- [DispatcherS] Added missing Responder methods to DispatcherS (#2954).
- [DispatcherS] The ArgDispatcher field for ThresholdS methods is now mandatory only if a connection to AttributeS has been defined (#2981).
- [DataManager] Revised caching logic for ActionPlans.
- [AttributeS] Introduced \*sipcid field type.
- [FilterS] New APIs for index status checks have been implemented.
- [SessionS] Tenant is set to default if not specified for SessionS APIs.
- [RALs] Fixed issue with \*any subject not considered when removing RatingProfiles (#3161).
- [ResourceS] ResourceS APIs updated for concurrent usage safety.
- [APIs] Addressed potential panic risk caused by API parameter validator function.
- [cgr-loader] Added tenant flag.
- [ApierS] Cache now reloaded when setting/removing RatingProfiles (#3186).
- [SessionS] Session synchronization no longer occurs with no active sessions.
- [RALs] Updated EventCost rounding increment handling (#3018).
- [SessionS] Protection added for missing events.
- [Config] Resolved issue with appending default port to multiple mongodb hosts in config file (#3673).
- [FSock] Fixed cgr-engine panic at startup when trying to connect to freeswitch_agent with logger set to \*stdout (#3678).
- [AttributeS/DispatcherS] Context/Subsystems now set to \*any if not specified.
- [FilterS] Added support for reverse filter indexes.
- [FSock] Addressed an issue where parsing responses from FreeSWITCH sometimes resulted in an unexpected number of values (#3749).
- [FSock] Resolved a connectivity issue where, if the connection between cgr-engine and the freeswitch agent was terminated during use, no reconnection attempts would be made (#3794).
- [Fsock] Corrected a parsing error where separators between parentheses were not ignored, leading to improper parsing of replies from the 'show channels' API call.
- [CDRe] Retained export_path as is for amqp, amqpv1, sqs, s3 and kafka exporters.
- [ServiceManager] Rectified a problem that prevented the RALs service from starting when the Responder was already running.
- [CDRs] Refund process now precedes debit during CDR rerating, fixing potential inaccuracies.
- [LoaderS] Introduced inline filter validation before DB write, preventing late-stage errors.
- [RPCClient] Updated to the latest version, addressing potential panic, deadlock, and data race issues.
- [CDRe] *exp.Cost path population no longer hardcoded to Cost found in *req map, user now can choose.
- [CDRe] Resolved an issue where the RoundingDecimals, if not explicitly set by the user, defaulted to 0 instead of the value defined under the "general" section in the configuration.
- [CDRe] Fixed a template problem where attempts to overwrite existing fields would lead to appending new values at the end of old ones, rather than replacing them.
- [CDRe] Overwriting preexisting fields in a template no longer appends new values at the end.
- [Storage] Introduced error handling for a previously overlooked case. Specifically, when GetCDRs is called for mongo with the remove flag set to true and the process returns an error, it previously led to a panic. This issue has now been addressed.
- [CDRe] Fixed support for \*combimed field type.
- [Docs] Updated installation documentation: https://cgrates.readthedocs.io/en/v0.10/installation.html.
- [CDRe] Added the possibility to override the exporter filter field through the API request signature.
- Updated all associated libraries to their most recent versions.
- Enhanced the testing suite and increased coverage.
- Fixes, updates and general quality of life changes that can be only noticed on the developer side so we will not be going into much detail:
  - updated ansible bash/ansible scripts;
  - improved formatting, readability;
  - reducing complexity of some functions.
- Implemented various fixes, updates, and enhancements primarily noticeable to developers (so we will not go into too much detail), including:

  - Enhanced code formatting and readability for better maintainability.
  - Simplified some complex functions to increase efficiency and ease of understanding.
  - Updated outdated Ansible and Bash scripts.

Full Changelog: https://github.com/cgrates/cgrates/compare/v0.10.2...v0.10.3

## [0.10.2] - 2020-10-08

- [SupplierS] Uniformize the logic in model_helpers.go
- [FilterS] Updated error message in case of unknown prefix
- [Server] Corectly log the server listen error
- [ERs] Add \*none EventReader type
- [ERs] Renamed \*default reader folders
- [General] Added *mo+extraDuration time support (e.g. *mo+1h will be time.Now() + 1 month + 1 hour)
- [SessionS] Use correctly SessionTTLUsage when calculate end usage in case of terminate session from ttl mechanism
- [RSRParsers] Removed attribute sistem from RSRParser
- [RSRParsers] Added grave accent(`) char as a delimiter to not split tge RSR value
- [SessionS] Rename from ResourceMessage to ResourceAllocation
- [AgentS] Correctly verify flags for setting max usage in ProcessEvent
- [AgentS] DiameterAgent return NOT_FOUND instead of "filter not passing" error and let other subsystem to handle this (e.g. FilterS)

## [0.10.1] - 2020-05-12

- [FilterS] Removed rals_conns in favor of reading the account
  directly from DataDB
- [SessionS] Added check for missing CGRevent
- [DiameterAgent] Using String function from diam.Message instead of
  ToJSON for request String method
- [DiameterAgent] Updated 3gp_vendor dictionary
- [Templates] Added new dataconverter: \*ip2hex
- [AgentS] Added support for *group type and correctly overwrite
  the values in case of *variable
- [ERs] Correctly populate ConcurrentRequest from config in
  EventReader
- [SupplierS] In case of missing usage from Event use 1 minute as
  default value
- [DataDB] Mongo support different marshaler than msgpack
- [ConnManager] Fixed rpc_conns handling id with two connections and one of
  it \*internal
- [Replicator] Added Limit and StaticTTL otions for Items from
  DataDB/StorDB
- [Migrator] Auto discover tenant from key instead of taking it from config
- [Templates] Fixed missing "\*" for strip and padding strategy
- [SessionS] Update subflags for *rals ( *authorize and \*initiate )
- [AgentRequest] Improved NavigableMap
- [AgentRequest] FieldAsInterface return Data instead of NMItem
- [SupplierS] Allow multiple suppliers with the same ID
- [Engine] Skip caching if limit is 0
- [CacheS] Avoid long recaching
- [SessionS] Use correctly SessionTTLUsage when calculate end usage in case of terminate session from ttl mechanism
- [SessionS] Add SessionTLLLastUsage as option for an extra debit in case of ttl mechanism
- [Templates] Added new dataconverter: \*string2hex
- [SessionS] Properly charge terminate without initiate event

## [0.10.0] - 2020-02-06

- Creating first stable branch.
