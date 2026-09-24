CREATE TABLE "user_entities" (
  "user_name" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT (now()),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "user_entities" ("user_name") DEFERRABLE INITIALLY IMMEDIATE;
-- CREATE UNIQUE INDEX ON "accounts" ("owner", "currency");
ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE ("owner", "currency");

ALTER TABLE "accounts" ADD FOREIGN KEY ("country_code") REFERENCES "countries" ("code") DEFERRABLE INITIALLY IMMEDIATE;