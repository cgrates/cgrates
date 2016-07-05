/*
This script will migrate CDRs from the old CGRateS tables to the new cdrs table
but it only migrate CDRs where the duration is > 0.
If you need CDRs also with duration is = 0 you can make the appropriate change in the line beginning WHERE cdrs_primary.usage

Also the script will process 10,000 CDRs before committing to save system resources 
especially in systems where they are millions of CDRs to be migrated
You can increase or lower the value of step in the line after BEGIN below.
*/

DO $$
DECLARE
    count_cdrs integer;
    start_id integer;
    end_id integer;
    step integer;
BEGIN
	step := 10000;
	start_id := 0;
	end_id := start_id + step;
	select count(*) INTO count_cdrs from rated_cdrs;
	WHILE start_id < count_cdrs
	LOOP
		INSERT INTO 
			cdrs(cgrid,run_id,origin_host,source,origin_id,tor,request_type,direction,tenant,category,account,subject,destination,setup_time,pdd,answer_time,usage,supplier,disconnect_cause,extra_fields,cost_source,cost,cost_details,extra_info, created_at, updated_at, deleted_at) 
			SELECT cdrs_primary.cgrid,rated_cdrs.runid as run_id,cdrs_primary.cdrhost as origin_host,cdrs_primary.cdrsource as source,cdrs_primary.accid as origin_id, cdrs_primary.tor,rated_cdrs.reqtype as request_type,rated_cdrs.direction, rated_cdrs.tenant,rated_cdrs.category, rated_cdrs.account, rated_cdrs.subject, rated_cdrs.destination,rated_cdrs.setup_time,rated_cdrs.pdd,rated_cdrs.answer_time,rated_cdrs.usage,rated_cdrs.supplier,rated_cdrs.disconnect_cause,cdrs_extra.extra_fields,cost_details.cost_source,rated_cdrs.cost,cost_details.timespans as cost_details,rated_cdrs.extra_info,rated_cdrs.created_at,rated_cdrs.updated_at, rated_cdrs.deleted_at 
			FROM rated_cdrs 
			INNER JOIN cdrs_primary ON rated_cdrs.cgrid = cdrs_primary.cgrid 
			INNER JOIN cdrs_extra ON rated_cdrs.cgrid = cdrs_extra.cgrid 
			INNER JOIN cost_details ON rated_cdrs.cgrid = cost_details.cgrid 
			WHERE cdrs_primary.usage > '0'
			AND not exists (select 1 from cdrs c where c.cgrid = cdrs_primary.cgrid)
			;
		start_id = start_id + step;
		end_id = end_id + step;
	END LOOP;
END 
$$;
