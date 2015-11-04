
--
-- Table structure for table `cdrs`
--

DROP TABLE IF EXISTS cdrs;
CREATE TABLE cdrs (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  runid  varchar(64) NOT NULL,
  tor  varchar(16) NOT NULL,
  accid varchar(64) NOT NULL,
  cdrhost varchar(64) NOT NULL,
  cdrsource varchar(64) NOT NULL,
  reqtype varchar(24) NOT NULL,
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
  cost DECIMAL(20,4) NOT NULL,
  timespans text,
  cost_source varchar(64) NOT NULL,
  extra_info text,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY cgrid (cgrid),
  KEY answer_time_idx (answer_time),
  KEY deleted_at_idx (deleted_at)
);
