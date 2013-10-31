
--
-- Table structure for table `rater_cdrs`
--
CREATE TABLE `rated_cdrs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cgrid` char(40) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `cost` DECIMAL(20,4) DEFAULT NULL,
  `extra_info` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `costid` (`cgrid`,`subject`)
);
