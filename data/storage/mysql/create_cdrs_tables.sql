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
ALTER TABLE cdrs ADD COLUMN urid VARCHAR(40) GENERATED ALWAYS AS ( JSON_VALUE(opts, '$."*urID"') );
CREATE UNIQUE INDEX opts_urid_idx ON cdrs (urid);