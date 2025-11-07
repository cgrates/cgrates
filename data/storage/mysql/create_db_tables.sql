--
-- Table structure for table `accounts`
--

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `account` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX accounts_idx ON accounts (`id`);

DROP TABLE IF EXISTS ip_profiles;
CREATE TABLE ip_profiles (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `ip_profile` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX ip_profiles_idx ON ip_profiles (`id`);

DROP TABLE IF EXISTS ip_allocations;
CREATE TABLE ip_allocations (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `ip_allocation` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX ip_allocations_idx ON ip_allocations (`id`);

DROP TABLE IF EXISTS action_profiles;
CREATE TABLE action_profiles (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `action_profile` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX action_profiles_idx ON action_profiles (`id`);

DROP TABLE IF EXISTS charger_profiles;
CREATE TABLE charger_profiles (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `charger_profile` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX charger_profiles_idx ON charger_profiles (`id`);

DROP TABLE IF EXISTS attribute_profiles;
CREATE TABLE attribute_profiles (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `attribute_profile` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX attribute_profiles_idx ON attribute_profiles (`id`);

DROP TABLE IF EXISTS resource_profiles;
CREATE TABLE resource_profiles (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `resource_profile` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX resource_profiles_idx ON resource_profiles (`id`);

DROP TABLE IF EXISTS resources;
CREATE TABLE resources (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `resource` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX resources_idx ON resources (`id`);

DROP TABLE IF EXISTS stat_queue_profiles;
CREATE TABLE stat_queue_profiles (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `stat_queue_profile` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX stat_queue_profiles_idx ON stat_queue_profiles (`id`);

DROP TABLE IF EXISTS stat_queues;
CREATE TABLE stat_queues (
 `pk` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `id` VARCHAR(64) NOT NULL,
 `stat_queue` JSON NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY unique_tenant_id (`tenant`, `id`)
);
CREATE UNIQUE INDEX stat_queues_idx ON stat_queues (`id`);