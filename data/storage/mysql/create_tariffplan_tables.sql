--
-- Table structure for table `tp_timings`
--
DROP TABLE IF EXISTS `tp_timings`;
CREATE TABLE `tp_timings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `years` varchar(255) NOT NULL,
  `months` varchar(255) NOT NULL,
  `month_days` varchar(255) NOT NULL,
  `week_days` varchar(255) NOT NULL,
  `time` varchar(32) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_tmid` (`tpid`,`tag`),
  UNIQUE KEY `tpid_tag` (`tpid`,`tag`)
);

--
-- Table structure for table `tp_destinations`
--

DROP TABLE IF EXISTS `tp_destinations`;
CREATE TABLE `tp_destinations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `prefix` varchar(24) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_dstid` (`tpid`,`tag`),
  UNIQUE KEY `tpid_dest_prefix` (`tpid`,`tag`,`prefix`)
);

--
-- Table structure for table `tp_rates`
--

DROP TABLE IF EXISTS `tp_rates`;
CREATE TABLE `tp_rates` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `connect_fee` decimal(7,4) NOT NULL,
  `rate` decimal(10,4) NOT NULL,
  `rate_unit` varchar(16) NOT NULL,
  `rate_increment` varchar(16) NOT NULL,
  `group_interval_start` varchar(16) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_tprate` (`tpid`,`tag`,`group_interval_start`),
  KEY `tpid` (`tpid`),
  KEY `tpid_rtid` (`tpid`,`tag`)
);

--
-- Table structure for table `destination_rates`
--

DROP TABLE IF EXISTS `tp_destination_rates`;
CREATE TABLE `tp_destination_rates` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `destinations_tag` varchar(64) NOT NULL,
  `rates_tag` varchar(64) NOT NULL,
  `rounding_method` varchar(255) NOT NULL,
  `rounding_decimals` tinyint(4) NOT NULL,
  `max_cost` decimal(7,4) NOT NULL,
  `max_cost_strategy` varchar(16) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_drid` (`tpid`,`tag`),
  UNIQUE KEY `tpid_drid_dstid` (`tpid`,`tag`,`destinations_tag`)
);

--
-- Table structure for table `tp_rating_plans`
--

DROP TABLE IF EXISTS `tp_rating_plans`;
CREATE TABLE `tp_rating_plans` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `destrates_tag` varchar(64) NOT NULL,
  `timing_tag` varchar(64) NOT NULL,
  `weight` DECIMAL(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  KEY `tpid_rpl` (`tpid`,`tag`),
  UNIQUE KEY `tpid_rplid_destrates_timings_weight` (`tpid`,`tag`,`destrates_tag`,`timing_tag`)
);

--
-- Table structure for table `tp_rate_profiles`
--

DROP TABLE IF EXISTS `tp_rating_profiles`;
CREATE TABLE `tp_rating_profiles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `loadid` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `category` varchar(32) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `activation_time` varchar(24) NOT NULL,
  `rating_plan_tag` varchar(64) NOT NULL,
  `fallback_subjects` varchar(64),
  `cdr_stat_queue_ids` varchar(64),
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
   KEY `tpid` (`tpid`),
  KEY `tpid_loadid` (`tpid`, `loadid`),
  UNIQUE KEY `tpid_loadid_tenant_category_dir_subj_atime` (`tpid`,`loadid`, `tenant`,`category`,`direction`,`subject`,`activation_time`)
);

--
-- Table structure for table `tp_shared_groups`
--

DROP TABLE IF EXISTS `tp_shared_groups`;
CREATE TABLE `tp_shared_groups` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `account` varchar(64) NOT NULL,
  `strategy` varchar(24) NOT NULL,
  `rating_subject` varchar(24) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_shared_group` (`tpid`,`tag`,`account`,`strategy`,`rating_subject`)
);

--
-- Table structure for table `tp_actions`
--

DROP TABLE IF EXISTS `tp_actions`;
CREATE TABLE `tp_actions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `action` varchar(24) NOT NULL,
  `balance_tag` varchar(64) NOT NULL,
  `balance_type` varchar(24) NOT NULL,
  `directions` varchar(8) NOT NULL,
  `units` varchar(256) NOT NULL,
  `expiry_time` varchar(24) NOT NULL,
  `timing_tags` varchar(128) NOT NULL,
  `destination_tags` varchar(64) NOT NULL,
  `rating_subject` varchar(64) NOT NULL,
  `categories` varchar(32) NOT NULL,
  `shared_groups` varchar(64) NOT NULL,
  `balance_weight` varchar(10) NOT NULL,
  `balance_blocker` varchar(5) NOT NULL,
  `balance_disabled` varchar(24) NOT NULL,
  `extra_parameters` varchar(256) NOT NULL,
  `filter` varchar(256) NOT NULL,
  `weight` DECIMAL(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_action` (`tpid`,`tag`,`action`,`balance_tag`,`balance_type`,`directions`,`expiry_time`,`timing_tags`,`destination_tags`,`shared_groups`,`balance_weight`,`weight`)
);

--
-- Table structure for table `tp_action_timings`
--

DROP TABLE IF EXISTS `tp_action_plans`;
CREATE TABLE `tp_action_plans` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `actions_tag` varchar(64) NOT NULL,
  `timing_tag` varchar(64) NOT NULL,
  `weight` DECIMAL(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_action_schedule` (`tpid`,`tag`,`actions_tag`)
);

--
-- Table structure for table `tp_action_triggers`
--

DROP TABLE IF EXISTS `tp_action_triggers`;
CREATE TABLE `tp_action_triggers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tag` varchar(64) NOT NULL,
  `unique_id` varchar(64) NOT NULL,
  `threshold_type` char(64) NOT NULL,
  `threshold_value` DECIMAL(20,4) NOT NULL,
  `recurrent` BOOLEAN NOT NULL,
  `min_sleep` varchar(16) NOT NULL,
  `expiry_time` varchar(24) NOT NULL,
  `activation_time` varchar(24) NOT NULL,
  `balance_tag` varchar(64) NOT NULL,
  `balance_type` varchar(24) NOT NULL,
  `balance_directions` varchar(8) NOT NULL,
  `balance_categories` varchar(32) NOT NULL,
  `balance_destination_tags` varchar(64) NOT NULL,
  `balance_rating_subject` varchar(64) NOT NULL,
  `balance_shared_groups` varchar(64) NOT NULL,
  `balance_expiry_time` varchar(24) NOT NULL,
  `balance_timing_tags` varchar(128) NOT NULL,
  `balance_weight` varchar(10) NOT NULL,
  `balance_blocker` varchar(5) NOT NULL,
  `balance_disabled` varchar(5) NOT NULL,
  `min_queued_items` int(11) NOT NULL,
  `actions_tag` varchar(64) NOT NULL,
  `weight` DECIMAL(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_trigger_definition` (`tpid`,`tag`,`balance_tag`,`balance_type`,`balance_directions`,`threshold_type`,`threshold_value`,`balance_destination_tags`,`actions_tag`)
);

--
-- Table structure for table `tp_account_actions`
--

DROP TABLE IF EXISTS `tp_account_actions`;
CREATE TABLE `tp_account_actions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `loadid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `account` varchar(64) NOT NULL,
  `action_plan_tag` varchar(64),
  `action_triggers_tag` varchar(64),
  `allow_negative` BOOLEAN NOT NULL,
  `disabled` BOOLEAN NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_account` (`tpid`,`loadid`,`tenant`,`account`)
);

--
-- Table structure for table `tp_derived_chargers`
--

DROP TABLE IF EXISTS tp_derived_chargers;
CREATE TABLE tp_derived_chargers (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `loadid` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `category` varchar(32) NOT NULL,
  `account` varchar(64) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `destination_ids` varchar(64) NOT NULL,
  `runid`  varchar(24) NOT NULL,
  `run_filters`  varchar(256) NOT NULL,
  `req_type_field`  varchar(64) NOT NULL,
  `direction_field`  varchar(64) NOT NULL,
  `tenant_field`  varchar(64) NOT NULL,
  `category_field`  varchar(64) NOT NULL,
  `account_field`  varchar(64) NOT NULL,
  `subject_field`  varchar(64) NOT NULL,
  `destination_field`  varchar(64) NOT NULL,
  `setup_time_field`  varchar(64) NOT NULL,
  `pdd_field`  varchar(64) NOT NULL,
  `answer_time_field`  varchar(64) NOT NULL,
  `usage_field`  varchar(64) NOT NULL,
  `supplier_field`  varchar(64) NOT NULL,
  `disconnect_cause_field`  varchar(64) NOT NULL,
  `rated_field`  varchar(64) NOT NULL,
  `cost_field`  varchar(64) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_users`
--

DROP TABLE IF EXISTS tp_users;
CREATE TABLE tp_users (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `user_name` varchar(64) NOT NULL,
  `masked` BOOLEAN NOT NULL,
  `attribute_name` varchar(64) NOT NULL,
  `attribute_value` varchar(64) NOT NULL,
  `weight` DECIMAL(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`)
);

--
-- Table structure for table `tp_aliases`
--

DROP TABLE IF EXISTS tp_aliases;
CREATE TABLE tp_aliases (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `category` varchar(64) NOT NULL,
  `account` varchar(64) NOT NULL,
  `subject` varchar(64) NOT NULL,
  `destination_id` varchar(64) NOT NULL,
  `context` varchar(64) NOT NULL,
  `target` varchar(64) NOT NULL,
  `original` varchar(64) NOT NULL,
  `alias` varchar(64) NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_aliases` (`tpid`,`direction`,`tenant`,`category`,`account`,`subject`,`context`, `target`)
);

--
-- Table structure for table `tp_resources`
--

DROP TABLE IF EXISTS tp_resources;
CREATE TABLE tp_resources (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `filter_ids` varchar(64) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `usage_ttl` varchar(32) NOT NULL,
  `limit` varchar(64) NOT NULL,
  `allocation_message` varchar(64) NOT NULL,
  `blocker` BOOLEAN NOT NULL,
  `stored` BOOLEAN NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `threshold_ids` varchar(64) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_resource` (`tpid`,`tenant`, `id`,`filter_ids` )
);

--
-- Table structure for table `tp_stats`
--

DROP TABLE IF EXISTS tp_stats;
CREATE TABLE tp_stats (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `filter_ids` varchar(64) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `queue_length` int(11) NOT NULL,
  `ttl` varchar(32) NOT NULL,
  `metrics` varchar(128) NOT NULL,
  `parameters` varchar(64) NOT NULL,
  `blocker` BOOLEAN NOT NULL,
  `stored` BOOLEAN NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `min_items` int(11) NOT NULL,
  `threshold_ids` varchar(64) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_stats` (`tpid`,  `tenant`, `id`, `filter_ids`)
);

--
-- Table structure for table `tp_threshold_cfgs`
--

DROP TABLE IF EXISTS tp_thresholds;
CREATE TABLE tp_thresholds (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `filter_ids` varchar(64) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `max_hits` int(11) NOT NULL,
  `min_hits` int(11) NOT NULL,
  `min_sleep` varchar(16) NOT NULL,
  `blocker` BOOLEAN NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `action_ids` varchar(64) NOT NULL,
  `async` BOOLEAN NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_thresholds` (`tpid`,`tenant`, `id`,`filter_ids`,`action_ids`)
);

--
-- Table structure for table `tp_filter`
--

DROP TABLE IF EXISTS tp_filters;
CREATE TABLE tp_filters (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `filter_type` varchar(16) NOT NULL,
  `filter_field_name` varchar(64) NOT NULL,
  `filter_field_values` varchar(256) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_filters` (`tpid`,`tenant`, `id`, `filter_type`, `filter_field_name`)
);

--
-- Table structure for table `tp_suppliers`
--


DROP TABLE IF EXISTS tp_suppliers;
CREATE TABLE tp_suppliers (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `filter_ids` varchar(64) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `sorting` varchar(32) NOT NULL,
  `sorting_parameters` varchar(64) NOT NULL,
  `supplier_id` varchar(32) NOT NULL,
  `supplier_filter_ids` varchar(64) NOT NULL,
  `supplier_account_ids` varchar(64) NOT NULL,
  `supplier_ratingplan_ids` varchar(64) NOT NULL,
  `supplier_resource_ids` varchar(64) NOT NULL,
  `supplier_stat_ids` varchar(64) NOT NULL,
  `supplier_weight` decimal(8,2) NOT NULL,
  `supplier_blocker` BOOLEAN NOT NULL,
  `supplier_parameters` varchar(64) NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_suppliers` (`tpid`,`tenant`,
    `id`,`filter_ids`,`supplier_id`,`supplier_filter_ids`,`supplier_account_ids`,
    `supplier_ratingplan_ids`,`supplier_resource_ids`,`supplier_stat_ids` )
);

--
-- Table structure for table `tp_attributes`
--

DROP TABLE IF EXISTS tp_attributes;
CREATE TABLE tp_attributes (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `contexts` varchar(64) NOT NULL,
  `filter_ids` varchar(64) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `field_name` varchar(64) NOT NULL,
  `initial` varchar(64) NOT NULL,
  `substitute` varchar(64) NOT NULL,
  `append` BOOLEAN NOT NULL,
  `blocker` BOOLEAN NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_attributes` (`tpid`,`tenant`,
    `id`,`filter_ids`,`field_name`,`initial`,`substitute` )
);

--
-- Table structure for table `tp_chargers`
--

DROP TABLE IF EXISTS tp_chargers;
CREATE TABLE tp_chargers (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `filter_ids` varchar(64) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `run_id` varchar(64) NOT NULL,
  `attribute_ids` varchar(64) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_chargers` (`tpid`,`tenant`,
    `id`,`filter_ids`,`run_id`,`attribute_ids`)
);

--
-- Table structure for table `versions`
--

DROP TABLE IF EXISTS versions;
CREATE TABLE versions (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `item` varchar(64) NOT NULL,
  `version` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_item` (`id`,`item`)
);
