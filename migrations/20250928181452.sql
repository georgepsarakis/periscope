-- Modify "alerts" table
ALTER TABLE "public"."alerts" ADD COLUMN "title" text NOT NULL, ADD COLUMN "description" text NULL;
-- Create index "idx_project_id" to table: "project_ingestion_api_keys"
CREATE INDEX "idx_project_id" ON "public"."project_ingestion_api_keys" ("project_id");
