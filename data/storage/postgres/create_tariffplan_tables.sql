--
-- Table structure for table `tp_timings`
--
DROP TABLE IF EXISTS tp_timings;
CREATE TABLE tp_timings (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  years VARCHAR(255) NOT NULL,
  months VARCHAR(255) NOT NULL,
  month_days VARCHAR(255) NOT NULL,
  week_days VARCHAR(255) NOT NULL,
  time VARCHAR(32) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE  (tpid, tag)
);
CREATE INDEX tptimings_tpid_idx ON tp_timings (tpid);
CREATE INDEX tptimings_idx ON tp_timings (tpid,tag);

--
-- Table structure for table `tp_destinations`
--

DROP TABLE IF EXISTS tp_destinations;
CREATE TABLE tp_destinations (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  prefix VARCHAR(24) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, tag, prefix)
);
CREATE INDEX tpdests_tpid_idx ON tp_destinations (tpid);
CREATE INDEX tpdests_idx ON tp_destinations (tpid,tag);

--
-- Table structure for table `tp_rates`
--

DROP TABLE IF EXISTS tp_rates;
CREATE TABLE tp_rates (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  connect_fee NUMERIC(7,4) NOT NULL,
  rate NUMERIC(10,4) NOT NULL,
  rate_unit VARCHAR(16) NOT NULL,
  rate_increment VARCHAR(16) NOT NULL,
  group_interval_start VARCHAR(16) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, tag, group_interval_start)
);
CREATE INDEX tprates_tpid_idx ON tp_rates (tpid);
CREATE INDEX tprates_idx ON tp_rates (tpid,tag);

--
-- Table structure for table `destination_rates`
--

DROP TABLE IF EXISTS tp_destination_rates;
CREATE TABLE tp_destination_rates (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  destinations_tag VARCHAR(64) NOT NULL,
  rates_tag VARCHAR(64) NOT NULL,
  rounding_method VARCHAR(255) NOT NULL,
  rounding_decimals SMALLINT NOT NULL,
  max_cost NUMERIC(7,4) NOT NULL,
  max_cost_strategy VARCHAR(16) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, tag , destinations_tag)
);
CREATE INDEX tpdestrates_tpid_idx ON tp_destination_rates (tpid);
CREATE INDEX tpdestrates_idx ON tp_destination_rates (tpid,tag);

--
-- Table structure for table `tp_rating_plans`
--

DROP TABLE IF EXISTS tp_rating_plans;
CREATE TABLE tp_rating_plans (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  destrates_tag VARCHAR(64) NOT NULL,
  timing_tag VARCHAR(64) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, tag, destrates_tag, timing_tag)
);
CREATE INDEX tpratingplans_tpid_idx ON tp_rating_plans (tpid);
CREATE INDEX tpratingplans_idx ON tp_rating_plans (tpid,tag);


--
-- Table structure for table `tp_rate_profiles`
--

DROP TABLE IF EXISTS tp_rating_profiles;
CREATE TABLE tp_rating_profiles (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  loadid VARCHAR(64) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  category VARCHAR(32) NOT NULL,
  subject VARCHAR(64) NOT NULL,
  activation_time VARCHAR(26) NOT NULL,
  rating_plan_tag VARCHAR(64) NOT NULL,
  fallback_subjects VARCHAR(64),
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, loadid, tenant, category, subject, activation_time)
);
CREATE INDEX tpratingprofiles_tpid_idx ON tp_rating_profiles (tpid);
CREATE INDEX tpratingprofiles_idx ON tp_rating_profiles (tpid,loadid,tenant,category,subject);

--
-- Table structure for table `tp_shared_groups`
--

DROP TABLE IF EXISTS tp_shared_groups;
CREATE TABLE tp_shared_groups (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  account VARCHAR(64) NOT NULL,
  strategy VARCHAR(24) NOT NULL,
  rating_subject VARCHAR(24) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, tag, account , strategy , rating_subject)
);
CREATE INDEX tpsharedgroups_tpid_idx ON tp_shared_groups (tpid);
CREATE INDEX tpsharedgroups_idx ON tp_shared_groups (tpid,tag);

--
-- Table structure for table `tp_actions`
--

DROP TABLE IF EXISTS tp_actions;
CREATE TABLE tp_actions (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  action VARCHAR(24) NOT NULL,
  extra_parameters VARCHAR(256) NOT NULL,
  filters VARCHAR(256) NOT NULL,
  balance_tag VARCHAR(64) NOT NULL,
  balance_type VARCHAR(24) NOT NULL,
  categories VARCHAR(32) NOT NULL,
  destination_tags VARCHAR(64) NOT NULL,
  rating_subject VARCHAR(64) NOT NULL,
  shared_groups VARCHAR(64) NOT NULL,
  expiry_time VARCHAR(26) NOT NULL,
  timing_tags VARCHAR(128) NOT NULL,
  units VARCHAR(256) NOT NULL,
  balance_weight VARCHAR(10) NOT NULL,
  balance_blocker VARCHAR(5) NOT NULL,
  balance_disabled VARCHAR(5) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, tag, action, balance_tag, balance_type, expiry_time, timing_tags, destination_tags, shared_groups, balance_weight, weight)
);
CREATE INDEX tpactions_tpid_idx ON tp_actions (tpid);
CREATE INDEX tpactions_idx ON tp_actions (tpid,tag);

--
-- Table structure for table `tp_action_timings`
--

DROP TABLE IF EXISTS tp_action_plans;
CREATE TABLE tp_action_plans (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  actions_tag VARCHAR(64) NOT NULL,
  timing_tag VARCHAR(64) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE  (tpid, tag, actions_tag, timing_tag)
);
CREATE INDEX tpactionplans_tpid_idx ON tp_action_plans (tpid);
CREATE INDEX tpactionplans_idx ON tp_action_plans (tpid,tag);

--
-- Table structure for table tp_action_triggers
--

DROP TABLE IF EXISTS tp_action_triggers;
CREATE TABLE tp_action_triggers (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  unique_id VARCHAR(64) NOT NULL,
  threshold_type VARCHAR(64) NOT NULL,
  threshold_value NUMERIC(20,4) NOT NULL,
  recurrent BOOLEAN NOT NULL,
  min_sleep VARCHAR(16) NOT NULL,
  expiry_time VARCHAR(26) NOT NULL,
  activation_time VARCHAR(26) NOT NULL,
  balance_tag VARCHAR(64) NOT NULL,
  balance_type VARCHAR(24) NOT NULL,
  balance_categories VARCHAR(32) NOT NULL,
  balance_destination_tags VARCHAR(64) NOT NULL,
  balance_rating_subject VARCHAR(64) NOT NULL,
  balance_shared_groups VARCHAR(64) NOT NULL,
  balance_expiry_time VARCHAR(26) NOT NULL,
  balance_timing_tags VARCHAR(128) NOT NULL,
  balance_weight VARCHAR(10) NOT NULL,
  balance_blocker VARCHAR(5) NOT NULL,
  balance_disabled VARCHAR(5) NOT NULL,
  actions_tag VARCHAR(64) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, tag, balance_tag, balance_type, threshold_type, threshold_value, balance_destination_tags, actions_tag)
);
CREATE INDEX tpactiontrigers_tpid_idx ON tp_action_triggers (tpid);
CREATE INDEX tpactiontrigers_idx ON tp_action_triggers (tpid,tag);

--
-- Table structure for table tp_account_actions
--

DROP TABLE IF EXISTS tp_account_actions;
CREATE TABLE tp_account_actions (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  loadid VARCHAR(64) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  account VARCHAR(64) NOT NULL,
  action_plan_tag VARCHAR(64),
  action_triggers_tag VARCHAR(64),
  allow_negative BOOLEAN NOT NULL,
  disabled BOOLEAN NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE,
  UNIQUE (tpid, loadid, tenant, account)
);
CREATE INDEX tpaccountactions_tpid_idx ON tp_account_actions (tpid);
CREATE INDEX tpaccountactions_idx ON tp_account_actions (tpid,loadid,tenant,account);


--
-- Table structure for table `tp_resources`
--

DROP TABLE IF EXISTS tp_resources;
CREATE TABLE tp_resources (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant"varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "usage_ttl" varchar(32) NOT NULL,
  "limit" varchar(64) NOT NULL,
  "allocation_message" varchar(64) NOT NULL,
  "blocker" BOOLEAN NOT NULL,
  "stored" BOOLEAN NOT NULL,
  "weight" NUMERIC(8,2) NOT NULL,
  "threshold_ids" varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
CREATE INDEX tp_resources_idx ON tp_resources (tpid);
CREATE INDEX tp_resources_unique ON tp_resources  ("tpid",  "tenant", "id", "filter_ids");


--
-- Table structure for table `tp_stats`
--

DROP TABLE IF EXISTS tp_stats;
CREATE TABLE tp_stats (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant"varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "queue_length" INTEGER NOT NULL,
  "ttl" varchar(32) NOT NULL,
  "min_items" INTEGER NOT NULL,
  "metric_ids" VARCHAR(128) NOT NULL,
  "metric_filter_ids" VARCHAR(128) NOT NULL,
  "stored" BOOLEAN NOT NULL,
  "blocker" BOOLEAN NOT NULL,
  "weight" decimal(8,2) NOT NULL,
  "threshold_ids" varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
CREATE INDEX tp_stats_idx ON tp_stats (tpid);
CREATE INDEX tp_stats_unique ON tp_stats  ("tpid","tenant", "id", "filter_ids","metric_ids");

--
-- Table structure for table `tp_rankings`
--

DROP TABLE IF EXISTS tp_rankings;
CREATE TABLE tp_rankings(
  "pk"  SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant" varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "schedule" varchar(32) NOT NULL,
  "stat_ids" varchar(64) NOT NULL,
  "metric_ids" varchar(64) NOT NULL,
  "sorting" varchar(32) NOT NULL,
  "sorting_parameters" varchar(64) NOT NULL,
  "stored" BOOLEAN NOT NULL,
  "threshold_ids" varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
CREATE INDEX tp_rankings_idx ON tp_rankings (tpid);
CREATE INDEX tp_rankings_unique ON tp_rankings  ("tpid","tenant", "id","stat_ids");

--
-- Table structure for tabls `tp_trends`
--

DROP TABLE IF EXISTS tp_trends;
CREATE TABLE tp_trends(
 "pk" SERIAL PRIMARY KEY,
 "tpid" varchar(64) NOT NULL,
 "tenant" varchar(64) NOT NULL,
 "id" varchar(64) NOT NULL,
 "schedule" varchar(64) NOT NULL,
 "stat_id" varchar(64) NOT NULL,
 "metrics" varchar(128) NOT NULL,
 "ttl" varchar(32) NOT NULL,
 "queue_length" INTEGER NOT NULL,
 "min_items" INTEGER NOT NULL,
 "correlation_type" varchar(64) NOT NULL,
 "tolerance" decimal(8,2) NOT NULL,
 "stored" BOOLEAN NOT NULL,
 "threshold_ids" varchar(64) NOT NULL,
 "created_at" TIMESTAMP
);
  CREATE INDEX tp_trends_idx ON tp_trends(tpid);
  CREATE INDEX tp_trends_unique ON  tp_trends("tpid","tenant","id","stat_id");

--
-- Table structure for table `tp_threshold_cfgs`
--

DROP TABLE IF EXISTS tp_thresholds;
CREATE TABLE tp_thresholds (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant"varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "max_hits" INTEGER NOT NULL,
  "min_hits" INTEGER NOT NULL,
  "min_sleep" varchar(16) NOT NULL,
  "blocker" BOOLEAN NOT NULL,
  "weight" decimal(8,2) NOT NULL,
  "action_ids" varchar(64) NOT NULL,
  "async" BOOLEAN NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
CREATE INDEX tp_thresholds_idx ON tp_thresholds (tpid);
CREATE INDEX tp_thresholds_unique ON tp_thresholds  ("tpid","tenant", "id","filter_ids","action_ids");

--
-- Table structure for table `tp_filter`
--

DROP TABLE IF EXISTS tp_filters;
CREATE TABLE tp_filters (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant" varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "type" varchar(16) NOT NULL,
  "element" varchar(64) NOT NULL,
  "values" varchar(256) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
  CREATE INDEX tp_filters_idx ON tp_filters (tpid);
  CREATE INDEX tp_filters_unique ON tp_filters  ("tpid","tenant", "id", "type", "element");

--
-- Table structure for table `tp_routes`
--

DROP TABLE IF EXISTS tp_routes;
CREATE TABLE tp_routes (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant"varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "sorting" varchar(32) NOT NULL,
  "sorting_parameters" varchar(64) NOT NULL,
  "route_id" varchar(32) NOT NULL,
  "route_filter_ids" varchar(64) NOT NULL,
  "route_account_ids" varchar(64) NOT NULL,
  "route_ratingplan_ids" varchar(64) NOT NULL,
  "route_resource_ids" varchar(64) NOT NULL,
  "route_stat_ids" varchar(64) NOT NULL,
  "route_weight" decimal(8,2) NOT NULL,
  "route_blocker" BOOLEAN NOT NULL,
  "route_parameters" varchar(64) NOT NULL,
  "weight" decimal(8,2) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
CREATE INDEX tp_routes_idx ON tp_routes (tpid);
CREATE INDEX tp_routes_unique ON tp_routes  ("tpid",  "tenant", "id",
  "filter_ids","route_id","route_filter_ids","route_account_ids",
  "route_ratingplan_ids","route_resource_ids","route_stat_ids");

  --
  -- Table structure for table `tp_attributes`
  --

  DROP TABLE IF EXISTS tp_attributes;
  CREATE TABLE tp_attributes (
    "pk" SERIAL PRIMARY KEY,
    "tpid" varchar(64) NOT NULL,
    "tenant"varchar(64) NOT NULL,
    "id" varchar(64) NOT NULL,
    "contexts" varchar(64) NOT NULL,
    "filter_ids" varchar(64) NOT NULL,
    "activation_interval" varchar(64) NOT NULL,
    "attribute_filter_ids" varchar(64) NOT NULL,
    "path" varchar(64) NOT NULL,
    "type" varchar(64) NOT NULL,
    "value" varchar(64) NOT NULL,
    "blocker" BOOLEAN NOT NULL,
    "weight" decimal(8,2) NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE
  );
  CREATE INDEX tp_attributes_ids ON tp_attributes (tpid);
  CREATE INDEX tp_attributes_unique ON tp_attributes  ("tpid",  "tenant", "id",
    "filter_ids","path","value");

  --
  -- Table structure for table `tp_chargers`
  --

  DROP TABLE IF EXISTS tp_chargers;
  CREATE TABLE tp_chargers (
    "pk" SERIAL PRIMARY KEY,
    "tpid" varchar(64) NOT NULL,
    "tenant"varchar(64) NOT NULL,
    "id" varchar(64) NOT NULL,
    "filter_ids" varchar(64) NOT NULL,
    "activation_interval" varchar(64) NOT NULL,
    "run_id" varchar(64) NOT NULL,
    "attribute_ids" varchar(64) NOT NULL,
    "weight" decimal(8,2) NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE
  );
  CREATE INDEX tp_chargers_ids ON tp_chargers (tpid);
  CREATE INDEX tp_chargers_unique ON tp_chargers  ("tpid",  "tenant", "id",
    "filter_ids","run_id","attribute_ids");

  --
  -- Table structure for table `tp_dispatchers`
  --

  DROP TABLE IF EXISTS tp_dispatcher_profiles;
  CREATE TABLE tp_dispatcher_profiles (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant" varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "subsystems" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "strategy" varchar(64) NOT NULL,
  "strategy_parameters" varchar(64) NOT NULL,
  "conn_id" varchar(64) NOT NULL,
  "conn_filter_ids" varchar(64) NOT NULL,
  "conn_weight" decimal(8,2) NOT NULL,
  "conn_blocker" BOOLEAN NOT NULL,
  "conn_parameters" varchar(64) NOT NULL,
  "weight" decimal(8,2) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
  );
  CREATE INDEX tp_dispatcher_profiles_ids ON tp_dispatcher_profiles (tpid);
  CREATE INDEX tp_dispatcher_profiles_unique ON tp_dispatcher_profiles  ("tpid",  "tenant", "id",
    "filter_ids","strategy","conn_id","conn_filter_ids");

--
-- Table structure for table `tp_dispatchers`
--

  DROP TABLE IF EXISTS tp_dispatcher_hosts;
  CREATE TABLE tp_dispatcher_hosts (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant" varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "address" varchar(64) NOT NULL,
  "transport" varchar(64) NOT NULL,
  "connect_attempts" INTEGER NOT NULL,
  "reconnects" INTEGER NOT NULL,
  "max_reconnect_interval" varchar(64) NOT NULL,
  "connect_timeout" varchar(64) NOT NULL,
  "reply_timeout" varchar(64) NOT NULL,
  "tls" BOOLEAN NOT NULL,
  "client_key" varchar(64) NOT NULL,
  "client_certificate" varchar(64) NOT NULL,
  "ca_certificate" varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
  );
  CREATE INDEX tp_dispatchers_hosts_ids ON tp_dispatcher_hosts (tpid);
  CREATE INDEX tp_dispatcher_hosts_unique ON tp_dispatcher_hosts  ("tpid",  "tenant", "id",
    "address");



--
-- Table structure for table `versions`
--

DROP TABLE IF EXISTS versions;
CREATE TABLE versions (
  "id" SERIAL PRIMARY KEY,
  "item" varchar(64) NOT NULL,
  "version" INTEGER NOT NULL,
  UNIQUE ("id","item")
);
