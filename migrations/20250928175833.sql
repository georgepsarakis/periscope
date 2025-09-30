-- Modify "events" table
ALTER TABLE "public"."events" ALTER COLUMN "title" TYPE text, ALTER COLUMN "title" SET NOT NULL;
-- Modify "project_alert_destinations" table
ALTER TABLE "public"."project_alert_destinations" ADD CONSTRAINT "fk_project_alert_destinations_alert_destination_type" FOREIGN KEY ("alert_destination_type_id") REFERENCES "public"."alert_destination_types" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create "project_ingestion_api_keys" table
CREATE TABLE "public"."project_ingestion_api_keys" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "key" text NOT NULL,
  "project_id" bigint NOT NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_projects_ingestion_api_keys" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- Create index "idx_project_ingestion_api_keys_deleted_at" to table: "project_ingestion_api_keys"
CREATE INDEX "idx_project_ingestion_api_keys_deleted_at" ON "public"."project_ingestion_api_keys" ("deleted_at");
-- Drop "ingestion_api_keys" table
DROP TABLE "public"."ingestion_api_keys";
