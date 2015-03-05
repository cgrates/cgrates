
--
-- Table structure for table `cdrs_primary`
--

DROP TABLE IF EXISTS cdrs_primary;
CREATE TABLE cdrs_primary (
  id SERIAL PRIMARY KEY,
  cgrid CHAR(40) NOT NULL,
  tor  VARCHAR(16) NOT NULL, 
  accid VARCHAR(64) NOT NULL,
  cdrhost VARCHAR(64) NOT NULL,
  cdrsource VARCHAR(64) NOT NULL,
  reqtype VARCHAR(24) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  category VARCHAR(32) NOT NULL,
  account VARCHAR(128) NOT NULL,
  subject VARCHAR(128) NOT NULL,
  destination VARCHAR(128) NOT NULL,
  setup_time TIMESTAMP NOT NULL,
  answer_time TIMESTAMP NOT NULL,
  usage NUMERIC(30,9) NOT NULL,
  created_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE (cgrid)
);
CREATE INDEX answer_time_idx ON cdrs_primary (answer_time);
CREATE INDEX deleted_at_cp_idx ON cdrs_primary (deleted_at);

--
-- Table structure for table `cdrs_extra`
--

DROP TABLE IF EXISTS cdrs_extra;
CREATE TABLE cdrs_extra (
  id SERIAL PRIMARY KEY,
  cgrid CHAR(40) NOT NULL,
  extra_fields text NOT NULL,
  created_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE (cgrid)
);
CREATE INDEX deleted_at_ce_idx ON cdrs_extra (deleted_at);

--
-- Table structure for table `cost_details`
--

DROP TABLE IF EXISTS cost_details;
CREATE TABLE cost_details (
  id SERIAL PRIMARY KEY,
  cgrid CHAR(40) NOT NULL,
  runid  VARCHAR(64) NOT NULL,
  tor  VARCHAR(16) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  tenant VARCHAR(128) NOT NULL,
  category VARCHAR(32) NOT NULL,
  account VARCHAR(128) NOT NULL,
  subject VARCHAR(128) NOT NULL,
  destination VARCHAR(128) NOT NULL,
  cost NUMERIC(20,4) NOT NULL,
  timespans text,
  cost_source VARCHAR(64) NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE (cgrid, runid)
);
CREATE INDEX deleted_at_cd_idx ON cost_details (deleted_at);

--
-- Table structure for table `rated_cdrs`
--
DROP TABLE IF EXISTS rated_cdrs;
CREATE TABLE rated_cdrs (
  id SERIAL PRIMARY KEY,
  cgrid CHAR(40) NOT NULL,
  runid  VARCHAR(64) NOT NULL,
  reqtype VARCHAR(24) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  tenant VARCHAR(64) NOT NULL,
  category VARCHAR(32) NOT NULL,
  account VARCHAR(128) NOT NULL,
  subject VARCHAR(128) NOT NULL,
  destination VARCHAR(128) NOT NULL,
  setup_time TIMESTAMP NOT NULL,
  answer_time TIMESTAMP NOT NULL,
  usage NUMERIC(30,9) NOT NULL,
  cost NUMERIC(20,4) DEFAULT NULL,
  extra_info text,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE (cgrid, runid)
);
CREATE INDEX deleted_at_rc_idx ON rated_cdrs (deleted_at);