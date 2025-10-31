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