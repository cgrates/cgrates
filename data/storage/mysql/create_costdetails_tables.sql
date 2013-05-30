
--
-- Table structure for table `cost_details`
--
CREATE TABLE `cost_details` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cgrid` char(40) NOT NULL,
  `accid` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `tor` varchar(16) NOT NULL,
  `account` varchar(64) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `destination` varchar(64) NOT NULL,
  `time_start` datetime NOT NULL,
  `cost` DECIMAL(5,4) NOT NULL,
  `connect_fee` DECIMAL(5,4) NOT NULL,
  `timespans` text,
  `source` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `costid` (`cgrid`,`subject`)
);

