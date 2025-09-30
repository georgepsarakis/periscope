-- Modify "alerts" table
ALTER TABLE "public"."alerts" ADD COLUMN "project_id" bigint NOT NULL;
-- Create index "idx_alert_project_id" to table: "alerts"
CREATE INDEX "idx_alert_project_id" ON "public"."alerts" ("project_id");
