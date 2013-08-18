
--
-- Table structure for table `cost_details`
--
CREATE TABLE `cost_details` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cgrid` char(40) NOT NULL,
  `accid` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(128) NOT NULL,
  `tor` varchar(32) NOT NULL,
  `account` varchar(128) NOT NULL,
  `subject` varchar(128) NOT NULL,
  `destination` varchar(128) NOT NULL,
  `cost` DECIMAL(5,4) NOT NULL,
  `connect_fee` DECIMAL(5,4) NOT NULL,
  `timespans` text,
  `source` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `costid` (`cgrid`,`subject`)
);

