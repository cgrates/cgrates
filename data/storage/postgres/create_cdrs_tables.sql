--
-- Table structure for table `cdrs`
--

DROP TABLE IF EXISTS cdrs;
CREATE TABLE cdrs (
 id SERIAL PRIMARY KEY,
 tenant VARCHAR(40) NOT NULL,
 opts jsonb NOT NULL,
 event jsonb NOT NULL,
 created_at TIMESTAMP WITH TIME ZONE,
 updated_at TIMESTAMP WITH TIME ZONE NULL,
 deleted_at TIMESTAMP WITH TIME ZONE NULL
);
CREATE UNIQUE INDEX opts_cdrid_idx ON cdrs( (opts->>'*cdrID') );
