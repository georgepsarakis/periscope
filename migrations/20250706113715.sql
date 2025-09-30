-- Modify "alert_destination_notifications" table
ALTER TABLE "public"."alert_destination_notifications" ADD COLUMN "completed_at" timestamptz NULL, ADD COLUMN "total_attempts" bigint NOT NULL;
-- Create index "idx_alert_destination_notifications_completed_at" to table: "alert_destination_notifications"
CREATE INDEX "idx_alert_destination_notifications_completed_at" ON "public"."alert_destination_notifications" ("completed_at");
-- Modify "projects" table
ALTER TABLE "public"."projects" DROP COLUMN "api_key_event_forwarding";
-- Create "ingestion_api_keys" table
CREATE TABLE "public"."ingestion_api_keys" (
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
-- Create index "idx_ingestion_api_keys_deleted_at" to table: "ingestion_api_keys"
CREATE INDEX "idx_ingestion_api_keys_deleted_at" ON "public"."ingestion_api_keys" ("deleted_at");
-- Create index "idx_project_id" to table: "ingestion_api_keys"
CREATE INDEX "idx_project_id" ON "public"."ingestion_api_keys" ("project_id");
