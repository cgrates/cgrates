
--
-- Table structure for table `rater_cdrs`
--
CREATE TABLE `rated_cdrs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cgrid` char(40) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `cost` double(20,4) default NULL,
  `source` char(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `costid` (`cgrid`,`subject`)
);
