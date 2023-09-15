
--
-- extra DB for ees and ers
DROP DATABASE IF EXISTS cgrates2;
CREATE DATABASE cgrates2;

GRANT ALL on cgrates2.* TO 'cgrates'@'127.0.0.1' IDENTIFIED BY 'CGRateS.org';
