
ALTER TABLE cdrs CHANGE COLUMN `usage` `usage_old` DECIMAL(30,9);
ALTER TABLE cdrs ADD `usage` BIGINT;
UPDATE cdrs SET `usage` = `usage_old` * 1000000000 WHERE usage_old IS NOT NULL;
ALTER TABLE cdrs DROP COLUMN usage_old;
