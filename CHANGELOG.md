# Changelog


## [0.11.0]

### Features and major improvements

- [EEs] Added generic event exporting, replacing CDRe. Added NATS, Elasticsearch, SQL, RPC and log exporters.
- [EEs] Added exporter metrics with scheduled resets. Saved failed exports now include exporter options and attempt counts for replay.
- [ERs] Added JSON file, AMQP, AMQPv1, S3, SQS and NATS readers. Added separate exports of original events after successful or failed processing.
- [Templates] Added shared field templates for ERs, EEs and agents.
- [RouteS] Replaced SupplierS. Added routing across multiple profiles and cost sorting for routes that specify only an account.
- [RankingS/TrendS] Added scheduled rankings and trend tracking for StatS metrics. Results can trigger ThresholdS actions and be exported through EEs.
- [PrometheusAgent] Added a Prometheus endpoint for StatS metrics, cache stats and CoreS runtime and process metrics.
- [AnalyzerS] Added capture and search for RPC, BiRPC and internal API calls.
- [CoreS] Added more runtime, garbage collection, process and caps metrics, and APIs to control profiling.
- [Agents/ERs] Added reporting of request outcomes and processing times to StatS and ThresholdS.
- [SessionS] Added configurable active-session backups to DataDB and restore on restart.
- [SessionS] Added `*dynaprepaid`, which runs configured SchedulerS plans when an account is missing, then retries charging or authorization.
- [SessionS] Added support for `TotalUsage` in session updates and reporting to StatS and ThresholdS. Added actions that alter or disconnect active sessions.
- [SessionS] Added APIs for creating and verifying STIR identities.
- [DiameterAgent] Added Diameter Sy spending-limit reporting and spending-status notifications.
- [DiameterAgent] Added Re-Auth-Request and Disconnect-Peer support and events reporting peer status.
- [RadiusAgent] Added RADIUS Dynamic Authorization with CoA and Disconnect Message support.
- [RadiusAgent] Added PAP, CHAP and MS-CHAPv2 password verification within the agent, and support for Status-Server requests.
- [DiameterAgent/RadiusAgent/DNSAgent] Added support for multiple listeners.
- [AsteriskAgent] Added session reconstruction using fields stored in channel variables.
- [FreeSWITCHAgent] Added configurable transfers when maximum usage is reached and announcements when the balance is low.
- [KamailioAgent] Added support for `dlg.briefing` replies when querying active dialogs.
- [SIPAgent] Added SIP request processing for authorization and event handling, including redirect responses.
- [JanusAgent] Added initial Janus gateway integration.
- [IPs] Added an IP address management (IPAM) service with authorization, allocation and release APIs.
- [RegistrarC] Added periodic registration of dispatcher and RPC hosts with registrars.
- [ConfigS] Added HTTP serving of configuration files and directories.
- [DataDB/StorDB] Added storage on disk, backups and restore for the internal databases.
- [DataDB] Added buffering and batching for replication. Failed writes can be saved for manual replay.
- [CacheS] Added cache replication and reads from remote caches.
- [DataDB/StorDB] Added Redis cluster, TLS and pool settings, MongoDB connection schemes, PostgreSQL schema and SSL options, SQL logging controls, and MySQL time zone and DSN parameters.
- [CDRs] Added optional gzip compression of stored CDR `CostDetails`.
- [Actions] Added actions that create or update profiles using event fields, such as an AttributeS profile for the current account.
- [SchedulerS] Added schedules that run after a delay such as `+10m` or repeat at intervals such as `*recurring+1h`.
- [AttributeS] Added attributes using live data from accounts, resources and stats. Added subtraction, multiplication, division, date formatting and HTTP requests, plus options to rerun profiles and ignore filters.
- [FilterS] Added regex, contains, CIDR and HTTP filters, plus APIBAN and SentryPeer checks. Added providers for ranking and trend data.
- [StatS] Added highest and lowest value metrics and metrics for successful and failed replies.
- [Converters] Added 3GPP ULI decoding, MCC-MNC names and more date, URL, SIP and number conversions.
- [RALs] Added balance transfer actions, balance weights based on the current time, and direct monetary debit.
- [RALs] Added balance filters across balance types when removing expired balances.
- [APIer] Added account and profile ID searches.
- [cgr-loader] Added support for the new profile types in cgr-loader and TPReader. Added recursive folder loading.

Full Changelog: https://github.com/cgrates/cgrates/compare/v0.10.5...v0.11.0
