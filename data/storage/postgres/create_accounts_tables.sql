--
-- Table structure for table `accounts`
--

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
  pk SERIAL PRIMARY KEY,
  tenant VARCHAR(40) NOT NULL,
  id varchar(64) NOT NULL,
  account jsonb NOT NULL,
  UNIQUE (tenant, id)
);
CREATE UNIQUE INDEX accounts_unique ON accounts ("id");
