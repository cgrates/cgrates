--
-- Table structure for table `accounts`
--

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id VARCHAR(64) NOT NULL,
  account JSONB NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX accounts_idx ON accounts ("id");
