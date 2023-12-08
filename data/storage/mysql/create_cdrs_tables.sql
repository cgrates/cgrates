--
-- Table structure for table `cdrs`
--

DROP TABLE IF EXISTS cdrs;
CREATE TABLE cdrs (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `opts` JSON NOT NULL,
 `event` JSON NOT NULL,
 `created_at` TIMESTAMP NULL,
 `updated_at` TIMESTAMP NULL,
 `deleted_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`)
);
ALTER TABLE cdrs ADD COLUMN cdrid VARCHAR(40) GENERATED ALWAYS AS ( JSON_VALUE(opts, '$."*cdrID"') );
CREATE UNIQUE INDEX opts_cdrid_idx ON cdrs (cdrid);