--
-- Table structure for table `cdrs_primary`
--
DROP TABLE IF EXISTS `cdrs_primary`;
CREATE TABLE `cdrs_primary` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cgrid` char(40) NOT NULL,
  `accid` varchar(64) NOT NULL,
  `direction` enum('0','1','2') NOT NULL DEFAULT '1',
  `tenant` varchar(64) NOT NULL,
  `tor` varchar(8) NOT NULL,
  `account` varchar(64) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `destination` varchar(64) NOT NULL,
  `time_start` datetime NOT NULL,
  `duration` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `cgrid` (`cgrid`)
);

--
-- Table structure for table cdrs_extra
--
DROP TABLE IF EXISTS `cdrs_extra`;
CREATE TABLE `cdrs_extra` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cgrid` char(40) NOT NULL,
  `extra_fields` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `cgrid` (`cgrid`)
);
