--
-- Table structure for table `tp_timings`
--

CREATE TABLE `tp_timings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tag` varchar(24) NOT NULL,
  `years` varchar(255) NOT NULL,
  `months` varchar(255) NOT NULL,
  `month_days` varchar(255) NOT NULL,
  `week_days` varchar(255) NOT NULL,
  `time` varchar(16) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_destinations`
--

CREATE TABLE `tp_destinatins` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tag` varchar(24) NOT NULL,
  `prefix` varchar(24) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_rates`
--

CREATE TABLE `tp_rates` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tag` varchar(24) NOT NULL,
  `destinations_tag` varchar(24) NOT NULL,
  `connect_fee` DECIMAL(5,4) NOT NULL,
  `rate` DECIMAL(5,4) NOT NULL,
  `rate_increments` INT(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_rate_timings`
--

CREATE TABLE `tp_rate_timings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tag` varchar(24) NOT NULL,
  `rates_tag` varchar(24) NOT NULL,
  `timings_tag` varchar(24) NOT NULL,
  `weight` smallint(5) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_rate_profiles`
--

CREATE TABLE `tp_rate_profiles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `tor` varchar(16) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `rates_fallback_subject` varchar(64),
  `rates_timing_tag` varchar(24) NOT NULL,
  `activation_time` char(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_actions`
--

CREATE TABLE `tp_actions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tag` varchar(24) NOT NULL,
  `action` varchar(24) NOT NULL,
  `balances_tag` varchar(24) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `units` int(11) NOT NULL,
  `destinations_tag` varchar(24) NOT NULL,
  `rate_type` varchar(8) NOT NULL,
  `rate` DECIMAL(5,4) NOT NULL,
  `minutes_weight` smallint(5) NOT NULL,
  `weight` smallint(5) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_action_timings`
--

CREATE TABLE `tp_action_timings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tag` varchar(24) NOT NULL,
  `actions_tag` varchar(24) NOT NULL,
  `timings_tag` varchar(24) NOT NULL,
  `weight` smallint(5) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_action_triggers`
--

CREATE TABLE `tp_action_triggers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tag` varchar(24) NOT NULL,
  `balances_tag` varchar(24) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `threshold` int(11) NOT NULL,
  `destinations_tag` varchar(24) NOT NULL,
  `actions_tag` varchar(24) NOT NULL,
  `weight` smallint(5) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_account_actions`
--

CREATE TABLE `tp_account_actions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` char(40) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `account` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `action_timings_tag` varchar(24),
  `action_triggers_tag` varchar(24),
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);
