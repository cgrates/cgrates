/*
This script will migrate CDRs from the old CGRateS tables to the new cdrs table
but it only migrate CDRs where the duration is > 0.
If you need CDRs also with duration is = 0 you can make the appropriate change in the line beginning WHERE cdrs_primary.usage
Also the script will process 10,000 CDRs before committing to save system resources 
especially in systems where they are millions of CDRs to be migrated
You can increase or lower the value of step in the line after BEGIN below.

You have to use 'CALL cgrates.migration();' to execute the script. If named other then default use that database name.
*/


DELIMITER //

CREATE PROCEDURE `migration`()
BEGIN
        /* DECLARE variables */
        DECLARE max_cdrs bigint;
        DECLARE start_id bigint;
        DECLARE end_id bigint;
        DECLARE step bigint;
        /* Optimize table for performance */
        ALTER TABLE cdrs DISABLE KEYS;
        SET autocommit=0;
        SET unique_checks=0;
        SET foreign_key_checks=0;
        /* You must change the step var to commit every step rows inserted */
        SET step := 10000;
        SET start_id := 0;
        SET end_id := start_id + step;
        SET max_cdrs = (select max(id) from rated_cdrs);
        WHILE (start_id <= max_cdrs) DO
                INSERT INTO
			cdrs(cgrid,run_id,origin_host,source,origin_id,tor,request_type,tenant,category,account,subject,destination,setup_time,pdd,answer_time,`usage`,supplier,disconnect_cause,extra_fields,cost_source,cost,cost_details,extra_info, created_at, updated_at, deleted_at) 
			SELECT cdrs_primary.cgrid,rated_cdrs.runid as run_id,cdrs_primary.cdrhost as origin_host,cdrs_primary.cdrsource as source,cdrs_primary.accid as origin_id, cdrs_primary.tor,rated_cdrs.reqtype as request_type, rated_cdrs.tenant,rated_cdrs.category, rated_cdrs.account, rated_cdrs.subject, rated_cdrs.destination,rated_cdrs.setup_time,rated_cdrs.pdd,rated_cdrs.answer_time,rated_cdrs.`usage`,rated_cdrs.supplier,rated_cdrs.disconnect_cause,cdrs_extra.extra_fields,cost_details.cost_source,rated_cdrs.cost,cost_details.timespans as cost_details,rated_cdrs.extra_info,rated_cdrs.created_at,rated_cdrs.updated_at, null 
                        FROM rated_cdrs
                        INNER JOIN cdrs_primary ON rated_cdrs.cgrid = cdrs_primary.cgrid
                        INNER JOIN cdrs_extra ON rated_cdrs.cgrid = cdrs_extra.cgrid
                        INNER JOIN cost_details ON rated_cdrs.cgrid = cost_details.cgrid
                        WHERE cdrs_primary.`usage` > '0'
                     	AND not exists (select 1 from cdrs where cdrs.cgrid = cdrs_primary.cgrid AND cdrs.run_id=rated_cdrs.runid)
						AND rated_cdrs.id >= start_id
						AND rated_cdrs.id < end_id
                    	GROUP BY cgrid, run_id, origin_id;
                SET start_id = start_id + step;
                SET end_id = end_id + step;
        END WHILE;
        /* SET Table for live usage */
       SET autocommit=1;
        SET unique_checks=1;
        SET foreign_key_checks=1;
        ALTER TABLE cdrs ENABLE KEYS;
        OPTIMIZE TABLE cdrs;
END //

DELIMITER ;

CALL cgrates.migration();