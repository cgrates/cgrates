
--
-- Table structure for table `rater_cdrs`
--
CREATE TABLE `rated_cdrs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cgrid` char(40) NOT NULL,
  `cost` double(20,4) default NULL,
  `cgrcostid` int(11) NOT NULL,
  `cdrsrc` char(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `cgrid` (`cgrid`)
);
