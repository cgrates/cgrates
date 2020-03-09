
--
-- Sample db and users creation. Replace here with your own details
--

DROP DATABASE IF EXISTS cgrates;
CREATE DATABASE cgrates;

GRANT ALL on cgrates.* TO 'cgrates'@'localhost' IDENTIFIED BY 'CGRateS.org';
