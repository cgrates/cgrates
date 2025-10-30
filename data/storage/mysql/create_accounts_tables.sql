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