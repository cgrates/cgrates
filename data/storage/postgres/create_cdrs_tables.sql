--
-- Table structure for table `cdrs`
--

DROP TABLE IF EXISTS cdrs;
CREATE TABLE cdrs (
 id SERIAL PRIMARY KEY,
 cgrid CHAR(40) NOT NULL,
 run_id VARCHAR(64) NOT NULL,
 origin_host VARCHAR(64) NOT NULL,
 source VARCHAR(64) NOT NULL,
 origin_id VARCHAR(64) NOT NULL,
 tor VARCHAR(16) NOT NULL,
 request_type VARCHAR(24) NOT NULL,
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
 extra_fields jsonb,
 cost_source VARCHAR(64) NOT NULL,
 cost NUMERIC(20,4) DEFAULT NULL,
 cost_details jsonb,
 extra_info text,
 created_at TIMESTAMP,
 updated_at TIMESTAMP,
 deleted_at TIMESTAMP,
 UNIQUE (cgrid, run_id, origin_id)
);
;
DROP INDEX IF EXISTS deleted_at_cp_idx;
CREATE INDEX deleted_at_cp_idx ON cdrs (deleted_at);


DROP TABLE IF EXISTS sm_costs;
CREATE TABLE sm_costs (
  id SERIAL PRIMARY KEY,
  cgrid CHAR(40) NOT NULL,
  run_id  VARCHAR(64) NOT NULL,
  origin_host VARCHAR(64) NOT NULL,
  origin_id VARCHAR(64) NOT NULL,
  cost_source VARCHAR(64) NOT NULL,
  usage NUMERIC(30,9) NOT NULL,
  cost_details jsonb,
  created_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE (cgrid, run_id)
);
DROP INDEX IF EXISTS cgrid_smcost_idx;
CREATE INDEX cgrid_smcost_idx ON sm_costs (cgrid, run_id);
DROP INDEX IF EXISTS origin_smcost_idx;
CREATE INDEX origin_smcost_idx ON sm_costs (origin_host, origin_id);
DROP INDEX IF EXISTS deleted_at_smcost_idx;
CREATE INDEX deleted_at_smcost_idx ON sm_costs (deleted_at);

