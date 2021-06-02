
--
-- Sample db and users creation. Replace here with your own details
--

DROP DATABASE IF EXISTS cgrates;
CREATE DATABASE cgrates;

GRANT ALL on cgrates.* TO 'cgrates'@'localhost' IDENTIFIED BY 'CGRateS.org';

-- extra DB for ers
DROP DATABASE IF EXISTS cgrates2;
CREATE DATABASE cgrates2;

GRANT ALL on cgrates2.* TO 'cgrates'@'localhost' IDENTIFIED BY 'CGRateS.org';
