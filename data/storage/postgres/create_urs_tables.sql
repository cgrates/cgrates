--
-- Table structure for table `urs`
--

DROP TABLE IF EXISTS urs;
CREATE TABLE urs (
 id SERIAL PRIMARY KEY,
 tenant VARCHAR(40) NOT NULL,
 opts jsonb NOT NULL,
 event jsonb NOT NULL,
 created_at TIMESTAMP WITH TIME ZONE,
 updated_at TIMESTAMP WITH TIME ZONE NULL,
 deleted_at TIMESTAMP WITH TIME ZONE NULL
);
CREATE UNIQUE INDEX opts_urid_idx ON urs( (opts->>'*urID') );
