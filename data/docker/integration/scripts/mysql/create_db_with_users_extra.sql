
--
-- Sample db and users creation. Replace here with your own details
--

CREATE DATABASE cgrates2;

GRANT ALL on cgrates.* TO 'cgrates'@'127.0.0.1' IDENTIFIED BY 'CGRateS.org';
GRANT ALL on cgrates2.* TO 'cgrates'@'127.0.0.1' IDENTIFIED BY 'CGRateS.org';
