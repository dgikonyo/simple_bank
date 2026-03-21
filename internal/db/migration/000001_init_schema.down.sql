-- Drop constraints first (foreign keys)
ALTER TABLE IF EXISTS "order_items"
DROP CONSTRAINT IF EXISTS "order_items_product_id_fkey";

ALTER TABLE IF EXISTS "products"
DROP CONSTRAINT IF EXISTS "products_merchant_id_fkey";

ALTER TABLE IF EXISTS "merchants"
DROP CONSTRAINT IF EXISTS "merchants_country_code_fkey";

ALTER TABLE IF EXISTS "merchants"
DROP CONSTRAINT IF EXISTS "merchants_admin_id_fkey";

ALTER TABLE IF EXISTS "transfers"
DROP CONSTRAINT IF EXISTS "transfers_from_account_id_fkey";

ALTER TABLE IF EXISTS "transfers"
DROP CONSTRAINT IF EXISTS "transfers_to_account_id_fkey";

transfers
ALTER TABLE IF EXISTS "entries"
DROP CONSTRAINT IF EXISTS "entries_account_id_fkey";

-- Drop indexes
DROP INDEX IF EXISTS "entries_account_id_idx";

DROP INDEX IF EXISTS "transfers_from_account_id_idx";

DROP INDEX IF EXISTS "transfers_to_account_id_idx";

DROP INDEX IF EXISTS "transfers_from_account_id_to_account_id_idx";

-- Drop tables in reverse order of creation (due to foreign key dependencies)
DROP TABLE IF EXISTS "order_items";

DROP TABLE IF EXISTS "orders";

DROP TABLE IF EXISTS "products";

DROP TABLE IF EXISTS "merchants";

DROP TABLE IF EXISTS "countries";

DROP TABLE IF EXISTS "transfers";

DROP TABLE IF EXISTS "entries";

DROP TABLE IF EXISTS "accounts";

DROP TABLE IF EXISTS "user_entities";

DROP TABLE IF EXISTS "schema_migrations";

-- Drop custom types
DROP TYPE IF EXISTS "Currency";