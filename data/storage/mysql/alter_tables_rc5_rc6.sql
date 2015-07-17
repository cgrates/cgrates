USE `cgrates`;

ALTER TABLE `cdrs_primary` 
	CHANGE COLUMN tbid `id` int(11) NOT NULL auto_increment first ,
	CHANGE `cgrid` `cgrid` char(40) NOT NULL after `id` ,
	ADD COLUMN `pdd` decimal(12,9) NOT NULL after `setup_time` ,
	CHANGE `answer_time` `answer_time` datetime   NOT NULL after `pdd` ,
	ADD COLUMN `supplier` varchar(128) NOT NULL after `usage` , 
	ADD COLUMN `disconnect_cause` varchar(64) NOT NULL after `supplier` , 
	ADD COLUMN `created_at` timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP  on update CURRENT_TIMESTAMP after `disconnect_cause` , 
	ADD COLUMN `deleted_at` timestamp   NOT NULL DEFAULT '0000-00-00 00:00:00' after `created_at` ,
	ADD KEY `answer_time_idx`(`answer_time`) , 
	ADD KEY `deleted_at_idx`(`deleted_at`) , 
	DROP KEY `PRIMARY`, ADD PRIMARY KEY(`id`) ;

ALTER TABLE `cdrs_extra` 
	CHANGE COLUMN tbid `id` int(11) NOT NULL auto_increment first ,
	CHANGE `cgrid` `cgrid` char(40) NOT NULL after `id` , 
	ADD COLUMN `created_at` timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP  on update CURRENT_TIMESTAMP after `extra_fields` , 
	ADD COLUMN `deleted_at` timestamp   NOT NULL DEFAULT '0000-00-00 00:00:00' after `created_at`,
	ADD UNIQUE KEY `cgrid`(`cgrid`) , 
	ADD KEY `deleted_at_idx`(`deleted_at`) , 
	DROP KEY `PRIMARY`, ADD PRIMARY KEY(`id`) ;

ALTER TABLE `cost_details`
	CHANGE COLUMN tbid `id` int(11) NOT NULL auto_increment first ,
	CHANGE `cost_source` `cost_source` varchar(64) NOT NULL after `timespans` , 
	ADD COLUMN `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP after `cost_source` , 
	ADD COLUMN `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' after `created_at` , 
	ADD COLUMN `deleted_at` timestamp   NOT NULL DEFAULT '0000-00-00 00:00:00' after `updated_at` , 
	DROP COLUMN `cost_time` , 
	ADD KEY `deleted_at_idx`(`deleted_at`) , 
	DROP KEY `PRIMARY`, ADD PRIMARY KEY(`id`) ;

ALTER TABLE `rated_cdrs` 
	CHANGE COLUMN tbid `id` int(11) NOT NULL auto_increment first ,
	CHANGE `cgrid` `cgrid` char(40) NOT NULL after `id` , 
	CHANGE `category` `category` varchar(32) NOT NULL after `tenant` , 
	ADD COLUMN `pdd` decimal(12,9)   NOT NULL after `setup_time` , 
	CHANGE `answer_time` `answer_time` datetime   NOT NULL after `pdd` , 
	ADD COLUMN `supplier` varchar(128) NOT NULL after `usage` , 
	ADD COLUMN `disconnect_cause` varchar(64) NOT NULL after `supplier` , 
	CHANGE `cost` `cost` decimal(20,4)   NULL after `disconnect_cause` , 
	ADD COLUMN `created_at` timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP  on update CURRENT_TIMESTAMP after `extra_info` , 
	ADD COLUMN `updated_at` timestamp   NOT NULL DEFAULT '0000-00-00 00:00:00' after `created_at` , 
	ADD COLUMN `deleted_at` timestamp   NOT NULL DEFAULT '0000-00-00 00:00:00' after `updated_at` , 
	DROP COLUMN `mediation_time` , 
	ADD KEY `deleted_at_idx`(`deleted_at`) , 
	DROP KEY `PRIMARY`, ADD PRIMARY KEY(`id`) ;