--
-- Table structure for table `accounts`
--

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  account JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX accounts_idx ON accounts ("id");

DROP TABLE IF EXISTS ip_profiles;
CREATE TABLE ip_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  ip_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX ip_profiles_idx ON ip_profiles ("id");


DROP TABLE IF EXISTS ip_allocations;
CREATE TABLE ip_allocations (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  ip_allocation JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX ip_allocations_idx ON ip_allocations ("id");


DROP TABLE IF EXISTS action_profiles;
CREATE TABLE action_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  action_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX action_profiles_idx ON action_profiles ("id");


DROP TABLE IF EXISTS charger_profiles;
CREATE TABLE charger_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  charger_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX charger_profiles_idx ON charger_profiles ("id");


DROP TABLE IF EXISTS attribute_profiles;
CREATE TABLE attribute_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  attribute_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX attribute_profiles_idx ON attribute_profiles ("id");


DROP TABLE IF EXISTS resource_profiles;
CREATE TABLE resource_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  resource_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX resource_profiles_idx ON resource_profiles ("id");


DROP TABLE IF EXISTS resources;
CREATE TABLE resources (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  resource JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX resources_idx ON resources ("id");

DROP TABLE IF EXISTS stat_queue_profiles;
CREATE TABLE stat_queue_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  stat_queue_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX stat_queue_profiles_idx ON stat_queue_profiles ("id");


DROP TABLE IF EXISTS stat_queues;
CREATE TABLE stat_queues (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  stat_queue JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX stat_queues_idx ON stat_queues ("id");


DROP TABLE IF EXISTS threshold_profiles;
CREATE TABLE threshold_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  threshold_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX threshold_profiles_idx ON threshold_profiles ("id");


DROP TABLE IF EXISTS thresholds;
CREATE TABLE thresholds (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  threshold JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX thresholds_idx ON thresholds ("id");


DROP TABLE IF EXISTS filters;
CREATE TABLE filters (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  filter JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE INDEX filters_idx ON filters ("id");


DROP TABLE IF EXISTS route_profiles;
CREATE TABLE route_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  route_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX route_profiles_idx ON route_profiles ("id");


DROP TABLE IF EXISTS rates;
DROP TABLE IF EXISTS rate_profiles;
CREATE TABLE rate_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  rate_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX rate_profiles_idx ON rate_profiles ("id");
CREATE TABLE rates (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  rate_profile_id VARCHAR(64) NOT NULL,
  rate JSONB NOT NULL,
  UNIQUE (tenant, id, rate_profile_id),
  FOREIGN KEY (rate_profile_id) REFERENCES rate_profiles (id)
);


DROP TABLE IF EXISTS ranking_profiles;
CREATE TABLE ranking_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  ranking_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX ranking_profiles_idx ON ranking_profiles ("id");


DROP TABLE IF EXISTS rankings;
CREATE TABLE rankings (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  ranking JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX rankings_idx ON rankings ("id");


DROP TABLE IF EXISTS trend_profiles;
CREATE TABLE trend_profiles (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  trend_profile JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX trend_profiles_idx ON trend_profiles ("id");


DROP TABLE IF EXISTS trends;
CREATE TABLE trends (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  trend JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX trends_idx ON trends ("id");


DROP TABLE IF EXISTS load_history;
CREATE TABLE load_history (
  key SERIAL PRIMARY KEY,
  load_instance JSONB NOT NULL
);

DROP TABLE IF EXISTS load_ids;
CREATE TABLE load_ids (
  pk SERIAL PRIMARY KEY,
  load_ids JSONB NOT NULL
);