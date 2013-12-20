DROP TABLE IF EXISTS cdrs_primary;

CREATE TABLE cdrs_primary (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  accid varchar(64) NOT NULL,
  cdrhost varchar(64) NOT NULL,
  cdrsource varchar(64) NOT NULL,
  reqtype varchar(24) NOT NULL,
  direction varchar(8) NOT NULL,
  tenant varchar(64) NOT NULL,
  tor varchar(16) NOT NULL,
  account varchar(128) NOT NULL,
  subject varchar(128) NOT NULL,
  destination varchar(128) NOT NULL,
  answer_time datetime NOT NULL,
  duration int NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY cgrid (cgrid)
);

DROP TABLE IF EXISTS cdrs_extra;

CREATE TABLE cdrs_extra (
  id int(11) NOT NULL AUTO_INCREMENT,
  cgrid char(40) NOT NULL,
  extra_fields text NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY cgrid (cgrid)
);
