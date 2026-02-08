CREATE TYPE "Currency" AS ENUM('USD', 'EUR');

CREATE TABLE "accounts" (
	"id" BIGSERIAL PRIMARY KEY NOT NULL,
	"owner" varchar NOT NULL,
	"balance" bigint NOT NULL,
	"currency" varchar NOT NULL,
	"country_code" int NOT NULL,
	"created_at" timestamptz NOT NULL DEFAULT (now()),
	"updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "entries" ("id" bigserial PRIMARY KEY NOT NULL, "account_id" bigint NOT NULL, "amount" bigint NOT NULL, "created_at" timestamptz NOT NULL DEFAULT (now()));

CREATE TABLE "transfers" ("id" bigserial PRIMARY KEY NOT NULL, "from_account_id" bigint NOT NULL, "to_account_id" bigint NOT NULL, "amount" bigint NOT NULL CHECK (amount > 0), "created_at" timestamptz NOT NULL DEFAULT (now()));

CREATE TABLE "merchants" ("id" bigserial PRIMARY KEY NOT NULL, "merchant_name" varchar NOT NULL, "country_code" int NOT NULL, "created_at" timestamptz NOT NULL DEFAULT (now()), "admin_id" int NOT NULL);

CREATE TABLE "countries" ("code" int PRIMARY KEY NOT NULL, "name" varchar, "continent_name" varchar);

CREATE TABLE "products" ("id" int PRIMARY KEY NOT NULL, "name" varchar NOT NULL, "merchant_id" int NOT NULL, "price" int, "status" varchar, "created_at" timestamptz NOT NULL DEFAULT (now()));
CREATE TABLE "order_items" ("order_id" int, "product_id" int NOT NULL, "quantity" int);

CREATE TABLE "orders" ("id" int PRIMARY KEY NOT NULL, "user_id" int, "status" varchar, "created_at" timestamptz NOT NULL DEFAULT (now()));

CREATE INDEX ON "accounts" ("owner");

CREATE INDEX ON "entries" ("account_id");

CREATE INDEX ON "transfers" ("from_account_id");

CREATE INDEX ON "transfers" ("to_account_id");

CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");

COMMENT ON COLUMN "entries"."amount" IS 'can be negative';

COMMENT ON COLUMN "transfers"."amount" IS 'must be positive';

ALTER TABLE "entries"
ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

ALTER TABLE "transfers"
ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

ALTER TABLE "transfers"
ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

ALTER TABLE "merchants"
ADD FOREIGN KEY ("country_code") REFERENCES "countries" ("code") ON DELETE CASCADE;

ALTER TABLE "merchants"
ADD FOREIGN KEY ("admin_id") REFERENCES "accounts" ("id") ON DELETE CASCADE;

ALTER TABLE "products"
ADD FOREIGN KEY ("merchant_id") REFERENCES "merchants" ("id") ON DELETE CASCADE;

ALTER TABLE "order_items"
ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE;

ALTER TABLE "order_items"
ADD FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE CASCADE;