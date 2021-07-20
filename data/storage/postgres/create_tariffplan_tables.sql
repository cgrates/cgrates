
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
  "action_profile_ids" varchar(64) NOT NULL,
  "async" BOOLEAN NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
CREATE INDEX tp_thresholds_idx ON tp_thresholds (tpid);
CREATE INDEX tp_thresholds_unique ON tp_thresholds  ("tpid","tenant", "id","filter_ids","action_profile_ids");

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
  `connect_attempts` INTEGER NOT NULL,
  `reconnects` INTEGER NOT NULL,
  `connect_timeout` varchar(64) NOT NULL,
  `reply_timeout` varchar(64) NOT NULL,
  `tls` BOOLEAN NOT NULL,
  `client_key` varchar(64) NOT NULL,
  `client_certificate` varchar(64) NOT NULL,
  `ca_certificate` varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
  );
  CREATE INDEX tp_dispatchers_hosts_ids ON tp_dispatcher_hosts (tpid);
  CREATE INDEX tp_dispatcher_hosts_unique ON tp_dispatcher_hosts  ("tpid",  "tenant", "id",
    "address");

--
-- Table structure for table `tp_rate_profiles`
--

  DROP TABLE IF EXISTS tp_rate_profiles;
  CREATE TABLE tp_rate_profiles (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant" varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "weights" varchar(64) NOT NULL,
  "min_cost" decimal(8,4) NOT NULL,
  "max_cost" decimal(8,4) NOT NULL,
  "max_cost_strategy" VARCHAR(64) NOT NULL,
  "rate_id" VARCHAR(64) NOT NULL,
  "rate_filter_ids" VARCHAR(64) NOT NULL,
  "rate_activation_times" VARCHAR(64) NOT NULL,
  "rate_weights" varchar(64) NOT NULL,
  "rate_blocker" BOOLEAN NOT NULL,
  "rate_interval_start" VARCHAR(64) NOT NULL,
  "rate_fixed_fee" decimal(8,4) NOT NULL,
  "rate_recurrent_fee" decimal(8,4) NOT NULL,
  "rate_unit" VARCHAR(64) NOT NULL,
  "rate_increment" VARCHAR(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
  );
  CREATE INDEX tp_rate_profiles_ids ON tp_rate_profiles (tpid);
  CREATE INDEX tp_rate_profiles_unique ON tp_rate_profiles  ("tpid",  "tenant", "id",
    "filter_ids", "rate_id");

--
-- Table structure for table `tp_action_profiles`
--


DROP TABLE IF EXISTS tp_action_profiles;
CREATE TABLE tp_action_profiles (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant" varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "weight" decimal(8,2) NOT NULL,
  "schedule" varchar(64) NOT NULL,
  "target_type" varchar(64) NOT NULL,
  "target_ids" varchar(64) NOT NULL,
  "action_id" varchar(64) NOT NULL,
  "action_filter_ids" varchar(64) NOT NULL,
  "action_blocker" BOOLEAN NOT NULL,
  "action_ttl" varchar(64) NOT NULL,
  "action_type" varchar(64) NOT NULL,
  "action_opts" varchar(256) NOT NULL,
  "action_path" varchar(64) NOT NULL,
  "action_value" varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
  );
  CREATE INDEX tp_action_profiles_ids ON tp_action_profiles (tpid);
  CREATE INDEX tp_action_profiles_unique ON tp_action_profiles  ("tpid",  "tenant", "id",
    "filter_ids", "action_id");


DROP TABLE IF EXISTS tp_accounts;
CREATE TABLE tp_accounts (
  "pk" SERIAL PRIMARY KEY,
  "tpid" varchar(64) NOT NULL,
  "tenant" varchar(64) NOT NULL,
  "id" varchar(64) NOT NULL,
  "filter_ids" varchar(64) NOT NULL,
  "activation_interval" varchar(64) NOT NULL,
  "weights" varchar(64) NOT NULL,
  "opts" varchar(256) NOT NULL,
  "balance_id" varchar(64) NOT NULL,
  "balance_filter_ids" varchar(64) NOT NULL,
  "balance_weights" varchar(64) NOT NULL,
  "balance_type" varchar(64) NOT NULL,
  "balance_units" decimal(16,4) NOT NULL,
  "balance_unit_factors" varchar(64) NOT NULL,
  "balance_opts" varchar(256) NOT NULL,
  "balance_cost_increments" varchar(64) NOT NULL,
  "balance_attribute_ids" varchar(64) NOT NULL,
  "balance_rate_profile_ids" varchar(64) NOT NULL,
  "threshold_ids" varchar(64) NOT NULL,
  "created_at" TIMESTAMP WITH TIME ZONE
);
 CREATE INDEX tp_accounts_ids ON tp_accounts (tpid);
 CREATE INDEX tp_accounts_unique ON tp_accounts  ("tpid",  "tenant", "id",
   "filter_ids", "balance_id");

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
