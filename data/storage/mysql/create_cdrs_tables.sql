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
  category varchar(32) NOT NULL,
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
  extra_info text,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY cdrrun (cgrid, run_id)
);
