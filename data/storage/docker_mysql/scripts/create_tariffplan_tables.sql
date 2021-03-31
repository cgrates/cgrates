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
  `min_items` int(11) NOT NULL,
  `metric_ids` varchar(128) NOT NULL,
  `metric_filter_ids` varchar(64) NOT NULL,
  `stored` BOOLEAN NOT NULL,
  `blocker` BOOLEAN NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `threshold_ids` varchar(64) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_stats` (`tpid`,  `tenant`, `id`, `filter_ids`,`metric_ids`)
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
  `type` varchar(16) NOT NULL,
  `element` varchar(64) NOT NULL,
  `values` varchar(256) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_filters` (`tpid`,`tenant`, `id`, `type`, `element`)
);

--
-- Table structure for table `tp_routes`
--


DROP TABLE IF EXISTS tp_routes;
CREATE TABLE tp_routes (
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
  UNIQUE KEY `unique_tp_routes` (`tpid`,`tenant`,
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
  `attribute_filter_ids` varchar(64) NOT NULL,
  `path` varchar(64) NOT NULL,
  `type` varchar(64) NOT NULL,
  `value` varchar(64) NOT NULL,
  `blocker` BOOLEAN NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_attributes` (`tpid`,`tenant`,
    `id`,`filter_ids`,`path`,`value` )
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
  `weight` decimal(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_chargers` (`tpid`,`tenant`,
    `id`,`filter_ids`,`run_id`,`attribute_ids`)
);

--
-- Table structure for table `tp_dispatchers`
--

DROP TABLE IF EXISTS tp_dispatcher_profiles;
CREATE TABLE tp_dispatcher_profiles (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `subsystems` varchar(64) NOT NULL,
  `filter_ids` varchar(64) NOT NULL,
  `activation_interval` varchar(64) NOT NULL,
  `strategy` varchar(64) NOT NULL,
  `strategy_parameters` varchar(64) NOT NULL,
  `conn_id` varchar(64) NOT NULL,
  `conn_filter_ids` varchar(64) NOT NULL,
  `conn_weight` decimal(8,2) NOT NULL,
  `conn_blocker` BOOLEAN NOT NULL,
  `conn_parameters` varchar(64) NOT NULL,
  `weight` decimal(8,2) NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_dispatcher_profiles` (`tpid`,`tenant`,
    `id`,`filter_ids`,`strategy`,`conn_id`,`conn_filter_ids`)
);

--
-- Table structure for table `tp_dispatchers`
--

DROP TABLE IF EXISTS tp_dispatcher_hosts;
CREATE TABLE tp_dispatcher_hosts (
  `pk` int(11) NOT NULL AUTO_INCREMENT,
  `tpid` varchar(64) NOT NULL,
  `tenant` varchar(64) NOT NULL,
  `id` varchar(64) NOT NULL,
  `address` varchar(64) NOT NULL,
  `transport` varchar(64) NOT NULL,
  `tls` BOOLEAN NOT NULL,
  `created_at` TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `tpid` (`tpid`),
  UNIQUE KEY `unique_tp_dispatchers_hosts` (`tpid`,`tenant`,
    `id`,`address`)
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
