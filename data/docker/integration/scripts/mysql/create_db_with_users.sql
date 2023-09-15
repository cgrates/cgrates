
--
-- Sample db and users creation. Replace here with your own details
--

DROP DATABASE IF EXISTS cgrates;
CREATE DATABASE cgrates;

GRANT ALL on cgrates.* TO 'cgrates'@'127.0.0.1' IDENTIFIED BY 'CGRateS.org';
