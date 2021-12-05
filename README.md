# EveGonline

## Solar Systems Table

```sql
CREATE TABLE IF NOT EXISTS `solar_systems` (`id` integer,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` text, `security_status` real, `region_id` integer,PRIMARY KEY (`id`));
CREATE INDEX `idx_solar_systems_deleted_at` ON `solar_systems`(`deleted_at`);
```

Using https://www.fuzzwork.co.uk/dump/latest/, (mapSolarSystems) rearrange CSV to match schema and import to DB

After import (gorm filter on deleted_at = NULL)
```sql
UPDATE solar_systems SET deleted_at = NULL;
```

## Inventory Type Table

```sql
CREATE TABLE IF NOT EXISTS "mappings" ("id" integer PRIMARY KEY, "name" TEXT, `created_at` datetime, `updated_at` datetime, `deleted_at` datetime, `category` text);
CREATE INDEX `idx_mappings_deleted_at` ON `mappings`(`deleted_at`);
```

Using https://www.fuzzwork.co.uk/dump/latest/, (invTypes.csv) rearrange CSV to match schema, add category == "inventory_type" and import to DB

After import (gorm filter on deleted_at = NULL)
```sql
UPDATE mappings SET deleted_at = NULL;
```

## About icons and rendered images

All renders should be in the static export (render), matching on ship type ID.
All items should be in the static export (types), matching on item type ID, with two sizes, 32 and 64 px.

## Tokens Table

```sql
CREATE TABLE IF NOT EXISTS "tokens" (`id` integer,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`access_token` text,`refresh_token` text,`char_id` integer, `exp` integer, `corp_
id` integer,PRIMARY KEY (`id`));
CREATE INDEX `idx_tokens_deleted_at` ON `tokens`(`deleted_at`);
```

## Assets table

```sql
CREATE TABLE IF NOT EXISTS "assets" (`id` integer,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`etag` text, `size` integer,PRIMARY KEY (`id`, `size`));
CREATE INDEX `idx_assets_deleted_at` ON "assets"(`deleted_at`);
```
