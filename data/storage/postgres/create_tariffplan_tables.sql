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
  created_at TIMESTAMP,
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
  created_at TIMESTAMP,
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
  rate NUMERIC(7,4) NOT NULL,
  rate_unit VARCHAR(16) NOT NULL,
  rate_increment VARCHAR(16) NOT NULL,
  group_interval_start VARCHAR(16) NOT NULL,
  created_at TIMESTAMP,
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
  created_at TIMESTAMP,
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
  created_at TIMESTAMP,
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
  direction VARCHAR(8) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  category VARCHAR(32) NOT NULL,
  subject VARCHAR(64) NOT NULL,
  activation_time VARCHAR(24) NOT NULL,
  rating_plan_tag VARCHAR(64) NOT NULL,
  fallback_subjects VARCHAR(64),
  cdr_stat_queue_ids varchar(64),
  created_at TIMESTAMP,
  UNIQUE (tpid, loadid, tenant, category, direction, subject, activation_time)
);
CREATE INDEX tpratingprofiles_tpid_idx ON tp_rating_profiles (tpid);
CREATE INDEX tpratingprofiles_idx ON tp_rating_profiles (tpid,loadid,direction,tenant,category,subject);

--
-- Table structure for table `tp_shared_groups`
--

DROP TABLE IF EXISTS tp_shared_groups;
CREATE TABLE tp_shared_groups (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  account VARCHAR(24) NOT NULL,
  strategy VARCHAR(24) NOT NULL,
  rating_subject VARCHAR(24) NOT NULL,
  created_at TIMESTAMP,
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
  balance_tag VARCHAR(64) NOT NULL,
  balance_type VARCHAR(24) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  units NUMERIC(20,4) NOT NULL,
  expiry_time VARCHAR(24) NOT NULL,
  timing_tags VARCHAR(128) NOT NULL,
  destination_tags VARCHAR(64) NOT NULL,
  rating_subject VARCHAR(64) NOT NULL,
  category VARCHAR(32) NOT NULL,
  shared_group VARCHAR(64) NOT NULL,
  balance_weight NUMERIC(8,2) NOT NULL,
  extra_parameters VARCHAR(256) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP,
  UNIQUE (tpid, tag, action, balance_tag, balance_type, direction, expiry_time, timing_tags, destination_tags, shared_group, balance_weight, weight)
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
  created_at TIMESTAMP,
  UNIQUE  (tpid, tag, actions_tag)
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
  balance_tag VARCHAR(64) NOT NULL,
  balance_type VARCHAR(24) NOT NULL,
  balance_direction VARCHAR(8) NOT NULL,
  threshold_type char(12) NOT NULL,
  threshold_value NUMERIC(20,4) NOT NULL,
  recurrent BOOLEAN NOT NULL,
  min_sleep VARCHAR(16) NOT NULL,
  balance_destination_tags VARCHAR(64) NOT NULL,
  balance_weight NUMERIC(8,2) NOT NULL,
  balance_expiry_time VARCHAR(24) NOT NULL,
  balance_timing_tags VARCHAR(128) NOT NULL,
  balance_rating_subject VARCHAR(64) NOT NULL,
  balance_category VARCHAR(32) NOT NULL,
  balance_shared_group VARCHAR(64) NOT NULL,
  min_queued_items INTEGER NOT NULL,
  actions_tag VARCHAR(64) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP,
  UNIQUE (tpid, tag, balance_tag, balance_type, balance_direction, threshold_type, threshold_value, balance_destination_tags, actions_tag)
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
  direction VARCHAR(8) NOT NULL,
  action_plan_tag VARCHAR(64),
  action_triggers_tag VARCHAR(64),
  created_at TIMESTAMP,
  UNIQUE (tpid, loadid, tenant, account, direction)
);
CREATE INDEX tpaccountactions_tpid_idx ON tp_account_actions (tpid);
CREATE INDEX tpaccountactions_idx ON tp_account_actions (tpid,loadid,tenant,account,direction);

--
-- Table structure for table `tp_lcr_rules`
--

DROP TABLE IF EXISTS tp_lcr_rules;
CREATE TABLE tp_lcr_rules (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  category VARCHAR(32) NOT NULL,
  account VARCHAR(24) NOT NULL,
  subject VARCHAR(64) NOT NULL,
  destination_tag VARCHAR(64) NOT NULL,
  rp_category VARCHAR(32) NOT NULL,
  strategy VARCHAR(16) NOT NULL,
  strategy_params VARCHAR(256) NOT NULL,
  activation_time VARCHAR(24) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP
);
CREATE INDEX tplcr_tpid_idx ON tp_lcr_rules (tpid);
CREATE INDEX tplcr_idx ON tp_lcr_rules (tpid,tenant,category,direction,account,subject,destination_tag);

--
-- Table structure for table `tp_derived_chargers`
--

DROP TABLE IF EXISTS tp_derived_chargers;
CREATE TABLE tp_derived_chargers (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  loadid VARCHAR(64) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  category VARCHAR(32) NOT NULL,
  account VARCHAR(24) NOT NULL,
  subject VARCHAR(64) NOT NULL,
  runid  VARCHAR(24) NOT NULL,
  run_filters  VARCHAR(256) NOT NULL,
  req_type_field  VARCHAR(24) NOT NULL,
  direction_field  VARCHAR(24) NOT NULL,
  tenant_field  VARCHAR(24) NOT NULL,
  category_field  VARCHAR(24) NOT NULL,
  account_field  VARCHAR(24) NOT NULL,
  subject_field  VARCHAR(24) NOT NULL,
  destination_field  VARCHAR(24) NOT NULL,
  setup_time_field  VARCHAR(24) NOT NULL,
  answer_time_field  VARCHAR(24) NOT NULL,
  usage_field  VARCHAR(24) NOT NULL,
  supplier_field  VARCHAR(24) NOT NULL,
  disconnect_cause_field  VARCHAR(24) NOT NULL,
  created_at TIMESTAMP
);
CREATE INDEX tpderivedchargers_tpid_idx ON tp_derived_chargers (tpid);
CREATE INDEX tpderivedchargers_idx ON tp_derived_chargers (tpid,loadid,direction,tenant,category,account,subject);


--
-- Table structure for table `tp_cdrstats`
--

DROP TABLE IF EXISTS tp_cdrstats;
CREATE TABLE tp_cdrstats (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  queue_length INTEGER NOT NULL,
  time_window VARCHAR(8) NOT NULL,
  metrics VARCHAR(64) NOT NULL,
  setup_interval VARCHAR(64) NOT NULL,
  tors VARCHAR(64) NOT NULL,
  cdr_hosts VARCHAR(64) NOT NULL,
  cdr_sources VARCHAR(64) NOT NULL,
  req_types VARCHAR(64) NOT NULL,
  directions VARCHAR(8) NOT NULL,
  tenants VARCHAR(64) NOT NULL,
  categories VARCHAR(32) NOT NULL,
  accounts VARCHAR(24) NOT NULL,
  subjects VARCHAR(64) NOT NULL,
  destination_prefixes VARCHAR(64) NOT NULL,
  usage_interval VARCHAR(64) NOT NULL,
  suppliers VARCHAR(64) NOT NULL,
  disconnect_causes VARCHAR(64) NOT NULL,
  mediation_runids VARCHAR(64) NOT NULL,
  rated_accounts VARCHAR(64) NOT NULL,
  rated_subjects VARCHAR(64) NOT NULL,
  cost_interval VARCHAR(24) NOT NULL,
  action_triggers VARCHAR(64) NOT NULL,
  created_at TIMESTAMP
);
CREATE INDEX tpcdrstats_tpid_idx ON tp_cdrstats (tpid);
CREATE INDEX tpcdrstats_idx ON tp_cdrstats (tpid,tag);
