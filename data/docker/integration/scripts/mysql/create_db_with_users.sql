
--
-- Sample db and users creation. Replace here with your own details
--

DROP DATABASE IF EXISTS cgrates;
CREATE DATABASE cgrates;
CREATE USER IF NOT EXISTS 'cgrates'@'127.0.0.1' IDENTIFIED BY 'CGRateS.org';
GRANT ALL PRIVILEGES ON cgrates.* TO 'cgrates'@'127.0.0.1' WITH GRANT OPTION;
FLUSH PRIVILEGES;