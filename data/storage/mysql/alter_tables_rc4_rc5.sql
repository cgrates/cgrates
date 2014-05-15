ALTER TABLE cdrs_primary
	CHANGE COLUMN tor category varchar(16) NOT NULL,
	CHANGE COLUMN duration `usage` bigint(20) NOT NULL,
	ADD COLUMN tor varchar(16) NOT NULL AFTER cgrid;

UPDATE cdrs_primary SET tor="*voice";

ALTER TABLE cost_details
	DROP COLUMN accid,
	MODIFY COLUMN cost_time datetime NOT NULL AFTER tbid,
	CHANGE COLUMN `source` cost_source varchar(64) NOT NULL AFTER cost_time,
	MODIFY COLUMN runid varchar(64) NOT NULL AFTER cgrid,
	CHANGE COLUMN tor category varchar(32) NOT NULL AFTER tenant,
	ADD COLUMN tor varchar(16) NOT NULL after runid,
	MODIFY COLUMN direction varchar(8) NOT NULL AFTER tor;

UPDATE cost_details SET tor="*voice";

ALTER TABLE rated_cdrs
	MODIFY COLUMN mediation_time datetime NOT NULL AFTER tbid,
	MODIFY COLUMN subject varchar(128) NOT NULL,
	ADD COLUMN reqtype varchar(24) NOT NULL AFTER runid,
	ADD COLUMN direction varchar(8) NOT NULL AFTER reqtype,
	ADD COLUMN tenant varchar(64) NOT NULL AFTER direction,
	ADD COLUMN category varchar(16) NOT NULL AFTER tenant,	
	ADD COLUMN account varchar(128) NOT NULL AFTER category,	
	ADD COLUMN destination varchar(128) NOT NULL AFTER subject,
	ADD COLUMN setup_time datetime NOT NULL AFTER destination,
	ADD COLUMN answer_time datetime NOT NULL AFTER setup_time,
	ADD COLUMN `usage` bigint(20) NOT NULL AFTER answer_time;

ALTER TABLE tp_rates
	DROP COLUMN rounding_method,
	DROP COLUMN rounding_decimals;

ALTER TABLE tp_destination_rates
	ADD COLUMN rounding_method varchar(255) NOT NULL,
	ADD COLUMN rounding_decimals tinyint(4) NOT NULL;

ALTER TABLE tp_rating_profiles
	DROP KEY tpid_loadid_tenant_tor_dir_subj_atime,
	CHANGE COLUMN tor category varchar(16) NOT NULL,
	ADD UNIQUE KEY `tpid_loadid_tenant_category_dir_subj_atime` (`tpid`,`loadid`,`tenant`,`category`,`direction`,`subject`,`activation_time`);





