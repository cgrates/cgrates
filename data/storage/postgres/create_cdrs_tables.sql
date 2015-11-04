--
-- Table structure for table `cdrs`
--

DROP TABLE IF EXISTS cdrs;
CREATE TABLE cdrs (
 id SERIAL PRIMARY KEY,
 cgrid CHAR(40) NOT NULL,
 runid VARCHAR(64) NOT NULL,
 tor VARCHAR(16) NOT NULL,
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
 pdd NUMERIC(12,9) NOT NULL,
 answer_time TIMESTAMP NOT NULL,
 usage NUMERIC(30,9) NOT NULL,
 supplier VARCHAR(128) NOT NULL,
 disconnect_cause VARCHAR(64) NOT NULL,
 extra_fields jsonb NOT NULL,
 cost NUMERIC(20,4) DEFAULT NULL,
 timespans jsonb,
 cost_source VARCHAR(64) NOT NULL,
 extra_info text,
 created_at TIMESTAMP,
 updated_at TIMESTAMP,
 deleted_at TIMESTAMP,
 UNIQUE (cgrid)
);
