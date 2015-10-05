
--
-- Table structure for table cdrs_primary
--

drop table if exists cgrates.cdrs_primary;
create table cgrates.cdrs_primary (
  id uuid,
  cgrid text,
  tor text,
  accid text,
  cdrhost text,
  cdrsource text,
  reqtype text,
  direction text,
  tenant text,
  category text,
  account text,
  subject text,
  destination text,
  setup_time timestamp,
  pdd decimal,
  answer_time timestamp,
  usage decimal,
  supplier text,
  disconnect_cause text,
  created_at timestamp,
  deleted_at timestamp,
  PRIMARY KEY (id, cgrid),
);

--
-- Table structure for table cdrs_extra
--

DROP TABLE IF EXISTS cgrates.cdrs_extra;
CREATE TABLE cgrates.cdrs_extra (
  id uuid,
  cgrid text,
  extra_fields text,
  created_at timestamp,
  deleted_at timestamp,
  PRIMARY KEY (id, cgrid),
);

--
-- Table structure for table cost_details
--

DROP TABLE IF EXISTS cgrates.cost_details;
CREATE TABLE cgrates.cost_details (
  id uuid,
  cgrid text,
  runid  text,
  tor  text,
  direction text,
  tenant text,
  category text,
  account text,
  subject text,
  destination text,
  cost decimal,
  timespans text,
  cost_source text,
  created_at timestamp,
  updated_at timestamp,
  deleted_at timestamp,
  PRIMARY KEY (id, cgrid),
);

--
-- Table structure for table rated_cdrs
--
DROP TABLE IF EXISTS cgrates.rated_cdrs;
CREATE TABLE cgrates.rated_cdrs (
  id uuid,
  cgrid text,
  runid  text,
  reqtype text,
  direction text,
  tenant text,
  category text,
  account text,
  subject text,
  destination text,
  setup_time timestamp,
  pdd decimal,
  answer_time timestamp,
  usage decimal,
  supplier text,
  disconnect_cause text,
  cost decimal,
  extra_info text,
  created_at timestamp,
  updated_at timestamp,
  deleted_at timestamp,
  PRIMARY KEY (id, cgrid),
);

 DROP TABLE IF EXISTS cgrates.string_index;
 CREATE TABLE cgrates.string_index (
    field text,
    value text,
    id uuid,
    PRIMARY KEY((field, value), id)
 );

 DROP TABLE IF EXISTS cgrates.decimal_index;
 CREATE TABLE cgrates.decimal_index (
    field text,
    value decimal,
    id uuid,
    PRIMARY KEY((field, value), id)
 );

 DROP TABLE IF EXISTS cgrates.time_index;
 CREATE TABLE cgrates.time_index (
    field text,
    value timestamp,
    id uuid,
    PRIMARY KEY((field, value), id)
 );
