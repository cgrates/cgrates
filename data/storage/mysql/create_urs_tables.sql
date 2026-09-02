--
-- Table structure for table `urs`
--

DROP TABLE IF EXISTS urs;
CREATE TABLE urs (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `tenant` VARCHAR(40) NOT NULL,
 `opts` JSON NOT NULL,
 `event` JSON NOT NULL,
 `created_at` TIMESTAMP NULL,
 `updated_at` TIMESTAMP NULL,
 `deleted_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`)
);
ALTER TABLE urs ADD COLUMN urid VARCHAR(40) GENERATED ALWAYS AS ( JSON_VALUE(opts, '$."*urID"') );
CREATE UNIQUE INDEX opts_urid_idx ON urs (urid);