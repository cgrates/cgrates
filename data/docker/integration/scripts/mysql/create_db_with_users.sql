
--
-- Sample db and users creation. Replace here with your own details
--

DROP DATABASE IF EXISTS cgrates;
CREATE DATABASE cgrates;
CREATE USER IF NOT EXISTS 'cgrates'@'localhost' IDENTIFIED BY 'CGRateS.org';
GRANT ALL PRIVILEGES ON cgrates.* TO 'cgrates'@'localhost' WITH GRANT OPTION;
FLUSH PRIVILEGES;
