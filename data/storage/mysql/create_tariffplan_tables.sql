--
-- Table structure for table `tp_timings`
--
CREATE TABLE `tp_timings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `years` varchar(255) NOT NULL,
  `months` varchar(255) NOT NULL,
  `month_days` varchar(255) NOT NULL,
  `week_days` varchar(255) NOT NULL,
  `time` varchar(16) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_tag` (`tpid`,`tag`),
  UNIQUE KEY `tpid_tmid` (`tpid`,`tag`)
);

--
-- Table structure for table `tp_destinations`
--

CREATE TABLE `tp_destinations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `prefix` varchar(24) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_tag` (`tpid`,`tag`),
  UNIQUE KEY `tpid_dest_prefix` (`tpid`,`tag`,`prefix`)
);

--
-- Table structure for table `tp_rates`
--

CREATE TABLE `tp_rates` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `connect_fee` decimal(5,4) NOT NULL,
  `rate` DECIMAL(5,4) NOT NULL,
  `rate_unit` int(11) NOT NULL,
  `rate_increment` int(11) NOT NULL,
  `group_interval_start` int(11) NOT NULL,
  `rounding_method` varchar(255) NOT NULL,
  `rounding_decimals` tinyint(4) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_tprate` (`tpid`,`tag`,`group_interval_start`),
  KEY `tpid` (`tpid`),
  KEY `tpid_tag` (`tpid`,`tag`)
);

--
-- Table structure for table `destination_rates`
--

CREATE TABLE `tp_destination_rates` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `destinations_tag` varchar(64) NOT NULL,
  `rates_tag` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_tag` (`tpid`,`tag`),
  UNIQUE KEY `tpid_tag_dst_rates` (`tpid`,`tag`,`destinations_tag`)
);

--
-- Table structure for table `tp_rating_plans`
--

CREATE TABLE `tp_rating_plans` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `destrates_tag` varchar(64) NOT NULL,
  `timing_tag` varchar(64) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_tag` (`tpid`,`tag`),
  UNIQUE KEY `tpid_tag_destrates_timings_weight` (`tpid`,`tag`,`destrates_tag`,`timing_tag`,`weight`)
);

--
-- Table structure for table `tp_rate_profiles`
--

CREATE TABLE `tp_rating_profiles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `tor` varchar(16) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `activation_time` varchar(24) NOT NULL,
  `rating_plan_tag` varchar(64) NOT NULL,
  `rates_fallback_subject` varchar(64),
  PRIMARY KEY (`id`),
  KEY `tpid_tag` (`tpid`, `tag`),
  UNIQUE KEY `tpid_tag_tenant_tor_dir_subj_atime` (`tpid`,`tag`, `tenant`,`tor`,`direction`,`subject`,`activation_time`)
);

--
-- Table structure for table `tp_actions`
--

CREATE TABLE `tp_actions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `action` varchar(24) NOT NULL,
  `balance_type` varchar(24) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `units` double(20,4) NOT NULL,
  `expiry_time` varchar(24) NOT NULL,
  `destination_tag` varchar(64) NOT NULL,
  `rating_subject` varchar(64) NOT NULL,
  `balance_weight` double(8,2) NOT NULL,
  `extra_parameters` varchar(256) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_action` (`tpid`,`tag`,`action`,`balance_type`,`direction`,`expiry_time`,`destination_tag`,`balance_weight`,`weight`)
);

--
-- Table structure for table `tp_action_timings`
--

CREATE TABLE `tp_action_timings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `actions_tag` varchar(64) NOT NULL,
  `timing_tag` varchar(64) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_action_schedule` (`tpid`,`tag`,`actions_tag`)
);

--
-- Table structure for table `tp_action_triggers`
--

CREATE TABLE `tp_action_triggers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `balance_type` varchar(24) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `threshold_type` char(12) NOT NULL,
  `threshold_value` double(20,4) NOT NULL,
  `destination_tag` varchar(64) NOT NULL,
  `actions_tag` varchar(64) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_trigger_definition` (`tpid`,`tag`,`balance_type`,`direction`,`threshold_type`,`threshold_value`,`destination_tag`,`actions_tag`)
);

--
-- Table structure for table `tp_account_actions`
--

CREATE TABLE `tp_account_actions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `account` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `action_timings_tag` varchar(64),
  `action_triggers_tag` varchar(64),
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_account` (`tpid`,`tag`,`tenant`,`account`,`direction`)
);
