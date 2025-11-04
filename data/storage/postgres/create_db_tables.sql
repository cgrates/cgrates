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
