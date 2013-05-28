
--
-- Table structure for table `callcosts`
--
CREATE TABLE `callcosts` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uuid` varchar(80),
  `source` varchar(32) NOT NULL,
  `direction` varchar(32) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `tor` varchar(8) NOT NULL,
  `account` varchar(64) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `destination` varchar(64) NOT NULL,
  `cost` double(20,4) default NULL,
  `connect_fee` double(20,4) default NULL,
  `timespans` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `cgrid` (`uuid`)
);

