--
-- Table structure for table `tp_timings`
--
DROP TABLE IF EXISTS `tp_timings`;
CREATE TABLE `tp_timings` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `years` varchar(255) NOT NULL,
  `months` varchar(255) NOT NULL,
  `month_days` varchar(255) NOT NULL,
  `week_days` varchar(255) NOT NULL,
  `time` varchar(16) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  KEY `tpid_tmid` (`tpid`,`id`),
  UNIQUE KEY `tpid_id` (`tpid`,`id`)
);

--
-- Table structure for table `tp_destinations`
--

DROP TABLE IF EXISTS `tp_destinations`;
CREATE TABLE `tp_destinations` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `prefix` varchar(24) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  KEY `tpid_dstid` (`tpid`,`id`),
  UNIQUE KEY `tpid_dest_prefix` (`tpid`,`id`,`prefix`)
);

--
-- Table structure for table `tp_rates`
--

DROP TABLE IF EXISTS `tp_rates`;
CREATE TABLE `tp_rates` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `connect_fee` decimal(7,4) NOT NULL,
  `rate` decimal(7,4) NOT NULL,
  `rate_unit` varchar(16) NOT NULL,
  `rate_increment` varchar(16) NOT NULL,
  `group_interval_start` varchar(16) NOT NULL,
  PRIMARY KEY (`tbid`),
  UNIQUE KEY `unique_tprate` (`tpid`,`id`,`group_interval_start`),
  KEY `tpid` (`tpid`),
  KEY `tpid_rtid` (`tpid`,`id`)
);

--
-- Table structure for table `destination_rates`
--

DROP TABLE IF EXISTS `tp_destination_rates`;
CREATE TABLE `tp_destination_rates` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `destinations_id` varchar(64) NOT NULL,
  `rates_id` varchar(64) NOT NULL,
  `rounding_method` varchar(255) NOT NULL,
  `rounding_decimals` tinyint(4) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  KEY `tpid_drid` (`tpid`,`id`),
  UNIQUE KEY `tpid_drid_dstid` (`tpid`,`id`,`destinations_id`)
);

--
-- Table structure for table `tp_rating_plans`
--

DROP TABLE IF EXISTS `tp_rating_plans`;
CREATE TABLE `tp_rating_plans` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `destrates_id` varchar(64) NOT NULL,
  `timing_id` varchar(64) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  KEY `tpid_rpl` (`tpid`,`id`),
  UNIQUE KEY `tpid_rplid_destrates_timings_weight` (`tpid`,`id`,`destrates_id`,`timing_id`)
);

--
-- Table structure for table `tp_rate_profiles`
--

DROP TABLE IF EXISTS `tp_rating_profiles`;
CREATE TABLE `tp_rating_profiles` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `loadid` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `category` varchar(16) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `activation_time` varchar(24) NOT NULL,
  `rating_plan_id` varchar(64) NOT NULL,
  `fallback_subjects` varchar(64),
  PRIMARY KEY (`tbid`),
  KEY `tpid_loadid` (`tpid`, `loadid`),
  UNIQUE KEY `tpid_loadid_tenant_category_dir_subj_atime` (`tpid`,`loadid`, `tenant`,`category`,`direction`,`subject`,`activation_time`)
);

--
-- Table structure for table `tp_shared_groups`
--

DROP TABLE IF EXISTS `tp_shared_groups`;
CREATE TABLE `tp_shared_groups` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `account` varchar(24) NOT NULL,
  `strategy` varchar(24) NOT NULL,
  `rating_subject` varchar(24) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_shared_group` (`tpid`,`id`,`account`,`strategy`,`rating_subject`)
);

--
-- Table structure for table `tp_actions`
--

DROP TABLE IF EXISTS `tp_actions`;
CREATE TABLE `tp_actions` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `action` varchar(24) NOT NULL,
  `balance_type` varchar(24) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `units` double(20,4) NOT NULL,
  `expiry_time` varchar(24) NOT NULL,
  `destination_id` varchar(64) NOT NULL,
  `rating_subject` varchar(64) NOT NULL,
  `category` varchar(16) NOT NULL,
  `shared_group` varchar(64) NOT NULL,
  `balance_weight` double(8,2) NOT NULL,
  `extra_parameters` varchar(256) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_action` (`tpid`,`id`,`action`,`balance_type`,`direction`,`expiry_time`,`destination_id`,`shared_group`,`balance_weight`,`weight`)
);

--
-- Table structure for table `tp_action_timings`
--

DROP TABLE IF EXISTS `tp_action_plans`;
CREATE TABLE `tp_action_plans` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `actions_id` varchar(64) NOT NULL,
  `timing_id` varchar(64) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_action_schedule` (`tpid`,`id`,`actions_id`)
);

--
-- Table structure for table `tp_action_triggers`
--

DROP TABLE IF EXISTS `tp_action_triggers`;
CREATE TABLE `tp_action_triggers` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `balance_type` varchar(24) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `threshold_type` char(12) NOT NULL,
  `threshold_value` double(20,4) NOT NULL,
  `recurrent` bool NOT NULL,
  `min_sleep` int(11) NOT NULL,
  `destination_id` varchar(64) NOT NULL,
  `balance_weight` double(8,2) NOT NULL, 
  `balance_expiry_time` varchar(24) NOT NULL, 
  `balance_rating_subject` varchar(64) NOT NULL,
  `balance_category` varchar(16) NOT NULL,
  `balance_shared_group` varchar(64) NOT NULL,
  `min_queued_items` int(11) NOT NULL,
  `actions_id` varchar(64) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_trigger_definition` (`tpid`,`id`,`balance_type`,`direction`,`threshold_type`,`threshold_value`,`destination_id`,`actions_id`)
);

--
-- Table structure for table `tp_account_actions`
--

DROP TABLE IF EXISTS `tp_account_actions`;
CREATE TABLE `tp_account_actions` (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `loadid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `account` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `action_plan_id` varchar(64),
  `action_triggers_id` varchar(64),
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_account` (`tpid`,`loadid`,`tenant`,`account`,`direction`)
);

--
-- Table structure for table `tp_lcr_rules`
--

DROP TABLE IF EXISTS tp_lcr_rules;
CREATE TABLE tp_lcr_rules (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `direction`	varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `customer` varchar(64) NOT NULL,
  `destination_id` varchar(64) NOT NULL,
  `category` varchar(16) NOT NULL,
  `strategy` varchar(16) NOT NULL,
  `suppliers`	varchar(64) NOT NULL,
  `activation_time` varchar(24) NOT NULL,
  `weight` double(8,2) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_derived_chargers`
--

DROP TABLE IF EXISTS tp_derived_chargers;
CREATE TABLE tp_derived_chargers (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `loadid` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `category` varchar(16) NOT NULL,
  `account` varchar(24) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `run_id`  varchar(24) NOT NULL,
  `run_filter`  varchar(24) NOT NULL,
  `reqtype_field`  varchar(24) NOT NULL,
  `direction_field`  varchar(24) NOT NULL,
  `tenant_field`  varchar(24) NOT NULL,
  `category_field`  varchar(24) NOT NULL,
  `account_field`  varchar(24) NOT NULL,
  `subject_field`  varchar(24) NOT NULL,
  `destination_field`  varchar(24) NOT NULL,
  `setup_time_field`  varchar(24) NOT NULL,
  `answer_time_field`  varchar(24) NOT NULL,
  `duration_field`  varchar(24) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`)
);


--
-- Table structure for table `tp_cdr_stats`
--

DROP TABLE IF EXISTS tp_cdr_stats;
CREATE TABLE tp_cdr_stats (
  `tbid` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `queue_length` int(11) NOT NULL,
  `time_window` int(11) NOT NULL,
  `metrics` varchar(64) NOT NULL,
  `setup_interval` varchar(64) NOT NULL,
  `tor` varchar(64) NOT NULL,
  `cdr_host` varchar(64) NOT NULL,
  `cdr_source` varchar(64) NOT NULL,
  `req_type` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `category` varchar(16) NOT NULL,
  `account` varchar(24) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `destination_prefix` varchar(64) NOT NULL,
  `usage_interval` varchar(64) NOT NULL,
  `mediation_run_ids` varchar(64) NOT NULL,
  `rated_account` varchar(64) NOT NULL,
  `rated_subject` varchar(64) NOT NULL,
  `cost_interval` varchar(24) NOT NULL,
  `action_triggers` varchar(64) NOT NULL,
  PRIMARY KEY (`tbid`),
  KEY `tpid` (`tpid`)
);
