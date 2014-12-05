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
  time VARCHAR(16) NOT NULL,
  created_at TIMESTAMP,
  UNIQUE  (tpid, tag)
);

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
  created_at TIMESTAMP,
  UNIQUE (tpid, tag , destinations_tag)
);

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
  category VARCHAR(16) NOT NULL,
  subject VARCHAR(64) NOT NULL,
  activation_time VARCHAR(24) NOT NULL,
  rating_plan_tag VARCHAR(64) NOT NULL,
  fallback_subjects VARCHAR(64),
  created_at TIMESTAMP,
  UNIQUE (tpid, loadid, tenant, category, direction, subject, activation_time)
);

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
  destination_tag VARCHAR(64) NOT NULL,
  rating_subject VARCHAR(64) NOT NULL,
  category VARCHAR(16) NOT NULL,
  shared_group VARCHAR(64) NOT NULL,
  balance_weight NUMERIC(8,2) NOT NULL,
  extra_parameters VARCHAR(256) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP,
  UNIQUE (tpid, tag, action, balance_tag, balance_type, direction, expiry_time, destination_tag, shared_group, balance_weight, weight)
);

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

--
-- Table structure for table tp_action_triggers
--

DROP TABLE IF EXISTS tp_action_triggers;
CREATE TABLE tp_action_triggers (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  balance_tag VARCHAR(64) NOT NULL,
  balance_type VARCHAR(24) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  threshold_type char(12) NOT NULL,
  threshold_value NUMERIC(20,4) NOT NULL,
  recurrent BOOLEAN NOT NULL,
  min_sleep BIGINT NOT NULL,
  destination_tag VARCHAR(64) NOT NULL,
  balance_weight NUMERIC(8,2) NOT NULL, 
  balance_expiry_time VARCHAR(24) NOT NULL, 
  balance_rating_subject VARCHAR(64) NOT NULL,
  balance_category VARCHAR(16) NOT NULL,
  balance_shared_group VARCHAR(64) NOT NULL,
  min_queued_items INTEGER NOT NULL,
  actions_tag VARCHAR(64) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP,
  UNIQUE (tpid, tag, balance_tag, balance_type, direction, threshold_type, threshold_value, destination_tag, actions_tag)
);

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

--
-- Table structure for table `tp_lcr_rules`
--

DROP TABLE IF EXISTS tp_lcr_rules;
CREATE TABLE tp_lcr_rules (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  direction	VARCHAR(8) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  customer VARCHAR(64) NOT NULL,
  destination_tag VARCHAR(64) NOT NULL,
  category VARCHAR(16) NOT NULL,
  strategy VARCHAR(16) NOT NULL,
  suppliers	VARCHAR(64) NOT NULL,
  activation_time VARCHAR(24) NOT NULL,
  weight NUMERIC(8,2) NOT NULL,
  created_at TIMESTAMP
);

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
  category VARCHAR(16) NOT NULL,
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
  created_at TIMESTAMP
);


--
-- Table structure for table `tp_cdr_stats`
--

DROP TABLE IF EXISTS tp_cdr_stats;
CREATE TABLE tp_cdr_stats (
  id SERIAL PRIMARY KEY,
  tpid VARCHAR(64) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  queue_length INTEGER NOT NULL,
  time_window VARCHAR(8) NOT NULL,
  metrics VARCHAR(64) NOT NULL,
  setup_interval VARCHAR(64) NOT NULL,
  tor VARCHAR(64) NOT NULL,
  cdr_host VARCHAR(64) NOT NULL,
  cdr_source VARCHAR(64) NOT NULL,
  req_type VARCHAR(64) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  category VARCHAR(16) NOT NULL,
  account VARCHAR(24) NOT NULL,
  subject VARCHAR(64) NOT NULL,
  destination_prefix VARCHAR(64) NOT NULL,
  usage_interval VARCHAR(64) NOT NULL,
  mediation_runids VARCHAR(64) NOT NULL,
  rated_account VARCHAR(64) NOT NULL,
  rated_subject VARCHAR(64) NOT NULL,
  cost_interval VARCHAR(24) NOT NULL,
  action_triggers VARCHAR(64) NOT NULL,
  created_at TIMESTAMP
);
