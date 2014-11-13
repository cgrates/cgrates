
--
-- Table structure for table `cdrs_primary`
--

DROP TABLE IF EXISTS cdrs_primary;
CREATE TABLE cdrs_primary (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  tor  varchar(16) NOT NULL, 
  accid varchar(64) NOT NULL,
  cdrhost varchar(64) NOT NULL,
  cdrsource varchar(64) NOT NULL,
  reqtype varchar(24) NOT NULL,
  direction varchar(8) NOT NULL,
  tenant varchar(64) NOT NULL,
  category varchar(16) NOT NULL,
  account varchar(128) NOT NULL,
  subject varchar(128) NOT NULL,
  destination varchar(128) NOT NULL,
  setup_time datetime NOT NULL,
  answer_time datetime NOT NULL,
  `usage` DECIMAL(30,9) NOT NULL,
  created_at TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY cgrid (cgrid)
);

--
-- Table structure for table `cdrs_extra`
--

DROP TABLE IF EXISTS cdrs_extra;
CREATE TABLE cdrs_extra (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  extra_fields text NOT NULL,
  created_at TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY cgrid (cgrid)
);

--
-- Table structure for table `cost_details`
--

DROP TABLE IF EXISTS cost_details;
CREATE TABLE cost_details (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  runid  varchar(64) NOT NULL,
  tor  varchar(16) NOT NULL,
  direction varchar(8) NOT NULL,
  tenant varchar(128) NOT NULL,
  category varchar(32) NOT NULL,
  account varchar(128) NOT NULL,
  subject varchar(128) NOT NULL,
  destination varchar(128) NOT NULL,
  cost DECIMAL(20,4) NOT NULL,
  timespans text,
  cost_source varchar(64) NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `costid` (`cgrid`,`runid`)
);

--
-- Table structure for table `rated_cdrs`
--
DROP TABLE IF EXISTS rated_cdrs;
CREATE TABLE `rated_cdrs` (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  runid  varchar(64) NOT NULL,
  reqtype varchar(24) NOT NULL,
  direction varchar(8) NOT NULL,
  tenant varchar(64) NOT NULL,
  category varchar(16) NOT NULL,
  account varchar(128) NOT NULL,
  subject varchar(128) NOT NULL,
  destination varchar(128) NOT NULL,
  setup_time datetime NOT NULL,
  answer_time datetime NOT NULL,
  `usage` DECIMAL(30,9) NOT NULL,
  cost DECIMAL(20,4) DEFAULT NULL,
  extra_info text,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `costid` (`cgrid`,`runid`)
);