-- +goose Up
-- create "agent_addresses" table
CREATE TABLE "public"."agent_addresses" (
  "chain_id" bigint NOT NULL,
  "collection_address" text NOT NULL,
  "token_id" numeric(78) NOT NULL,
  "type" text NOT NULL,
  "value" text NOT NULL,
  PRIMARY KEY ("chain_id", "collection_address", "token_id", "type", "value")
);
-- create index "idx_agent_addresses_type_value" to table: "agent_addresses"
CREATE INDEX "idx_agent_addresses_type_value" ON "public"."agent_addresses" ("type", "value");
-- create "agent_collections" table
CREATE TABLE "public"."agent_collections" (
  "chain_id" bigint NOT NULL,
  "address" text NOT NULL,
  "name" text NOT NULL,
  "symbol" text NOT NULL,
  "description" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("chain_id", "address")
);
-- create "agent_ens_records" table
CREATE TABLE "public"."agent_ens_records" (
  "chain_id" bigint NOT NULL,
  "collection_address" text NOT NULL,
  "token_id" numeric(78) NOT NULL,
  "key" text NOT NULL,
  "value" text NOT NULL,
  PRIMARY KEY ("chain_id", "collection_address", "token_id", "key")
);
-- create "agents" table
CREATE TABLE "public"."agents" (
  "chain_id" bigint NOT NULL,
  "collection_address" text NOT NULL,
  "token_id" numeric(78) NOT NULL,
  "name" text NOT NULL,
  "ens_name" text NOT NULL,
  "creator" text NOT NULL,
  "owner" text NOT NULL,
  "clone_of" numeric(78) NULL,
  "images" jsonb NOT NULL,
  "attributes" jsonb NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("chain_id", "collection_address", "token_id")
);
-- create index "idx_agents_ens_name" to table: "agents"
CREATE INDEX "idx_agents_ens_name" ON "public"."agents" ("ens_name");
-- create index "idx_agents_owner" to table: "agents"
CREATE INDEX "idx_agents_owner" ON "public"."agents" ("owner");
-- create "sync_cursors" table
CREATE TABLE "public"."sync_cursors" (
  "chain_id" bigint NOT NULL,
  "block" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("chain_id")
);

-- +goose Down
-- reverse: create "sync_cursors" table
DROP TABLE "public"."sync_cursors";
-- reverse: create index "idx_agents_owner" to table: "agents"
DROP INDEX "public"."idx_agents_owner";
-- reverse: create index "idx_agents_ens_name" to table: "agents"
DROP INDEX "public"."idx_agents_ens_name";
-- reverse: create "agents" table
DROP TABLE "public"."agents";
-- reverse: create "agent_ens_records" table
DROP TABLE "public"."agent_ens_records";
-- reverse: create "agent_collections" table
DROP TABLE "public"."agent_collections";
-- reverse: create index "idx_agent_addresses_type_value" to table: "agent_addresses"
DROP INDEX "public"."idx_agent_addresses_type_value";
-- reverse: create "agent_addresses" table
DROP TABLE "public"."agent_addresses";
