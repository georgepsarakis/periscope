-- Modify "project_alert_destinations" table
ALTER TABLE "public"."project_alert_destinations" DROP COLUMN "configuration";
-- Create "alert_destination_notification_webhook_configurations" table
CREATE TABLE "public"."alert_destination_notification_webhook_configurations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "project_alert_destination_id" bigint NOT NULL,
  "url" text NOT NULL,
  "http_method" text NOT NULL,
  "headers" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_alert_destination_notification_webhook_configuratio26157943" to table: "alert_destination_notification_webhook_configurations"
CREATE INDEX "idx_alert_destination_notification_webhook_configuratio26157943" ON "public"."alert_destination_notification_webhook_configurations" ("deleted_at");
-- Modify "project_ingestion_api_keys" table
ALTER TABLE "public"."project_ingestion_api_keys" DROP CONSTRAINT "fk_projects_ingestion_api_keys", ADD CONSTRAINT "fk_projects_project_ingestion_api_keys" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
