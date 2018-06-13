--
-- Table structure for table `cdrs`
--

DROP TABLE IF EXISTS cdrs;
CREATE TABLE cdrs (
 id SERIAL PRIMARY KEY,
 cgrid VARCHAR(40) NOT NULL,
 run_id VARCHAR(64) NOT NULL,
 origin_host VARCHAR(64) NOT NULL,
 source VARCHAR(64) NOT NULL,
 origin_id VARCHAR(128) NOT NULL,
 tor VARCHAR(16) NOT NULL,
 request_type VARCHAR(24) NOT NULL,
 tenant VARCHAR(64) NOT NULL,
 category VARCHAR(64) NOT NULL,
 account VARCHAR(128) NOT NULL,
 subject VARCHAR(128) NOT NULL,
 destination VARCHAR(128) NOT NULL,
 setup_time TIMESTAMP WITH TIME ZONE NOT NULL,
 answer_time TIMESTAMP WITH TIME ZONE NOT NULL,
 usage BIGINT NOT NULL,
 extra_fields jsonb NOT NULL,
 cost_source VARCHAR(64) NOT NULL,
 cost NUMERIC(20,4) DEFAULT NULL,
 cost_details jsonb,
 extra_info text,
 created_at TIMESTAMP WITH TIME ZONE,
 updated_at TIMESTAMP WITH TIME ZONE NULL,
 deleted_at TIMESTAMP WITH TIME ZONE NULL,
 UNIQUE (cgrid, run_id, origin_id)
);
;
DROP INDEX IF EXISTS deleted_at_cp_idx;
CREATE INDEX deleted_at_cp_idx ON cdrs (deleted_at);


DROP TABLE IF EXISTS sessions_costs;
CREATE TABLE sessions_costs (
  id SERIAL PRIMARY KEY,
  cgrid VARCHAR(40) NOT NULL,
  run_id  VARCHAR(64) NOT NULL,
  origin_host VARCHAR(64) NOT NULL,
  origin_id VARCHAR(128) NOT NULL,
  cost_source VARCHAR(64) NOT NULL,
  usage BIGINT NOT NULL,
  cost_details jsonb,
  created_at TIMESTAMP WITH TIME ZONE,
  deleted_at TIMESTAMP WITH TIME ZONE NULL,
  UNIQUE (cgrid, run_id)
);
DROP INDEX IF EXISTS cgrid_sessionscost_idx;
CREATE INDEX cgrid_sessionscost_idx ON sessions_costs (cgrid, run_id);
DROP INDEX IF EXISTS origin_sessionscost_idx;
CREATE INDEX origin_sessionscost_idx ON sessions_costs (origin_host, origin_id);
DROP INDEX IF EXISTS run_origin_sessionscost_idx;
CREATE INDEX run_origin_sessionscost_idx ON sessions_costs (run_id, origin_id);
DROP INDEX IF EXISTS deleted_at_sessionscost_idx;
CREATE INDEX deleted_at_sessionscost_idx ON sessions_costs (deleted_at);
