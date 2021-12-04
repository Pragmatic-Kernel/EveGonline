# KillmailsGetter

## Solar Systems Table

CREATE TABLE IF NOT EXISTS `solar_systems` (`id` integer,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` text, `security_status` real, `region_id` integer,PRIMARY KEY (`id`));
CREATE INDEX `idx_solar_systems_deleted_at` ON `solar_systems`(`deleted_at`);

Using https://www.fuzzwork.co.uk/dump/latest/, (mapSolarSystems) rearrange CSV to match schema and import to DB

After import (gorm filter on deleted_at = NULL)
UPDATE solar_systems SET deleted_at = NULL;

## Inventory Type Table

CREATE TABLE IF NOT EXISTS "mappings" ("id" integer PRIMARY KEY, "name" TEXT, `created_at` datetime, `updated_at` datetime, `deleted_at` datetime, `category` text);
CREATE INDEX `idx_mappings_deleted_at` ON `mappings`(`deleted_at`);

Using https://www.fuzzwork.co.uk/dump/latest/, (invTypes.csv) rearrange CSV to match schema, add category == "inventory_type" and import to DB

After import (gorm filter on deleted_at = NULL)
UPDATE mappings SET deleted_at = NULL;
