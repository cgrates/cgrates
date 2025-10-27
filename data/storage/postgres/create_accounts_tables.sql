--
-- Table structure for table `accounts`
--

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
  pk SERIAL PRIMARY KEY,
  id varchar(64) NOT NULL,
  account bytea NOT NULL
);
CREATE UNIQUE INDEX accounts_unique ON accounts ("id");
