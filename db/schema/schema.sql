CREATE TABLE "accounts" (
  "id" bigint PRIMARY KEY,
  "balance" numeric(20,5) CHECK (balance >= 0) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "transactions" (
  "id" bigserial PRIMARY KEY,
  "source_account_id" bigint NOT NULL,
  "destination_account_id" bigint NOT NULL,
  "amount" numeric(20,5) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "accounts" ("id");

CREATE INDEX ON "accounts" ("created_at");

CREATE INDEX ON "transactions" ("id");

CREATE INDEX ON "transactions" ("created_at");

CREATE INDEX ON "transactions" ("source_account_id");

CREATE INDEX ON "transactions" ("destination_account_id");

CREATE INDEX ON "transactions" ("source_account_id", "destination_account_id");

CREATE INDEX ON "transactions" ("source_account_id", "destination_account_id", "amount");

COMMENT ON COLUMN "accounts"."balance" IS 'positive';

COMMENT ON COLUMN "transactions"."amount" IS 'positive';

ALTER TABLE "transactions" ADD FOREIGN KEY ("source_account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transactions" ADD FOREIGN KEY ("destination_account_id") REFERENCES "accounts" ("id");
