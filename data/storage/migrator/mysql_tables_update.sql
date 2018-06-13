use cgrates;
--
-- Table structure for table `cdrs`
--

DROP TABLE IF EXISTS cdrs;
CREATE TABLE cdrs (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  run_id  varchar(64) NOT NULL,
  origin_host varchar(64) NOT NULL,
  source varchar(64) NOT NULL,
  origin_id varchar(64) NOT NULL,
  tor  varchar(16) NOT NULL,
  request_type varchar(24) NOT NULL,
  direction varchar(8) NOT NULL,
  tenant varchar(64) NOT NULL,
  category varchar(64) NOT NULL,
  account varchar(128) NOT NULL,
  subject varchar(128) NOT NULL,
  destination varchar(128) NOT NULL,
  setup_time datetime NOT NULL,
  pdd DECIMAL(12,9) NOT NULL,
  answer_time datetime NOT NULL,
  `usage` DECIMAL(30,9) NOT NULL,
  supplier varchar(128) NOT NULL,
  disconnect_cause varchar(64) NOT NULL,
  extra_fields text NOT NULL,
  cost_source varchar(64) NOT NULL,
  cost DECIMAL(20,4) NOT NULL,
  cost_details text,
  account_summary text,
  extra_info text,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY cdrrun (cgrid, run_id, origin_id)
);

DROP TABLE IF EXISTS sessions_costs;
CREATE TABLE sessions_costs (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  run_id  varchar(64) NOT NULL,
  origin_host varchar(64) NOT NULL,
  origin_id varchar(64) NOT NULL,
  cost_source varchar(64) NOT NULL,
  `usage` DECIMAL(30,9) NOT NULL,
  cost_details text,
  created_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY costid (cgrid, run_id),
  KEY origin_idx (origin_host, origin_id),
  KEY run_origin_idx (run_id, origin_id),
  KEY deleted_at_idx (deleted_at)
);

--
-- Table structure for table `versions`
--

DROP TABLE IF EXISTS versions;
CREATE TABLE versions (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `item` varchar(64) NOT NULL,
  `version` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `item` (`item`)
);
