-- 1. Remove the unique constraint from accounts
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";

-- 2. Remove the foreign key constraints from accounts
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "accounts_country_code_fkey";
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";

-- 3. Drop the user_entities table
DROP TABLE IF EXISTS "user_entities";