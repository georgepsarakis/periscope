-- Create "alert_destination_notifications" table
CREATE TABLE "public"."alert_destination_notifications" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "alert_id" bigint NOT NULL,
  "project_alert_destination_id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_alert_destination_notifications_deleted_at" to table: "alert_destination_notifications"
CREATE INDEX "idx_alert_destination_notifications_deleted_at" ON "public"."alert_destination_notifications" ("deleted_at");
-- Create index "idx_alert_destinations_alert_id" to table: "alert_destination_notifications"
CREATE INDEX "idx_alert_destinations_alert_id" ON "public"."alert_destination_notifications" ("alert_id");
-- Create "alert_destination_types" table
CREATE TABLE "public"."alert_destination_types" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "title" text NOT NULL,
  "key" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_alert_destination_types_deleted_at" to table: "alert_destination_types"
CREATE INDEX "idx_alert_destination_types_deleted_at" ON "public"."alert_destination_types" ("deleted_at");
-- Create "alerts" table
CREATE TABLE "public"."alerts" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "event_group_id" bigint NOT NULL,
  "triggered_at" timestamptz NOT NULL,
  "escalated_at" timestamptz NULL,
  "acknowledged_at" timestamptz NULL,
  "notified_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_alert_event_grp_key" to table: "alerts"
CREATE INDEX "idx_alert_event_grp_key" ON "public"."alerts" ("event_group_id");
-- Create index "idx_alert_triggered_at_key" to table: "alerts"
CREATE INDEX "idx_alert_triggered_at_key" ON "public"."alerts" ("triggered_at");
-- Create index "idx_alerts_deleted_at" to table: "alerts"
CREATE INDEX "idx_alerts_deleted_at" ON "public"."alerts" ("deleted_at");
-- Create "event_groups" table
CREATE TABLE "public"."event_groups" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "total_count" bigint NOT NULL,
  "event_received_at" timestamptz NOT NULL,
  "project_id" bigint NOT NULL,
  "aggregation_key" text NOT NULL,
  "alert_triggered_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_event_groups_deleted_at" to table: "event_groups"
CREATE INDEX "idx_event_groups_deleted_at" ON "public"."event_groups" ("deleted_at");
-- Create index "idx_proj_aggr_key" to table: "event_groups"
CREATE INDEX "idx_proj_aggr_key" ON "public"."event_groups" ("project_id", "aggregation_key");
-- Create "events" table
CREATE TABLE "public"."events" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "event_id" text NOT NULL,
  "fingerprint" text NOT NULL,
  "stack_trace" json NULL,
  "event_group_id" bigint NOT NULL,
  "project_id" bigint NOT NULL,
  "emitted_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_event_group_id" to table: "events"
CREATE INDEX "idx_event_group_id" ON "public"."events" ("event_group_id");
-- Create index "idx_events_deleted_at" to table: "events"
CREATE INDEX "idx_events_deleted_at" ON "public"."events" ("deleted_at");
-- Create index "idx_project_id_emitted_at" to table: "events"
CREATE INDEX "idx_project_id_emitted_at" ON "public"."events" ("project_id", "emitted_at");
-- Create "project_alert_destinations" table
CREATE TABLE "public"."project_alert_destinations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "project_id" bigint NOT NULL,
  "alert_destination_type_id" bigint NOT NULL,
  "configuration" json NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_project_alert_destinations_deleted_at" to table: "project_alert_destinations"
CREATE INDEX "idx_project_alert_destinations_deleted_at" ON "public"."project_alert_destinations" ("deleted_at");
-- Create "projects" table
CREATE TABLE "public"."projects" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "api_key_event_forwarding" text NOT NULL,
  "public_id" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_projects_deleted_at" to table: "projects"
CREATE INDEX "idx_projects_deleted_at" ON "public"."projects" ("deleted_at");
-- Create index "uq_project_name" to table: "projects"
CREATE UNIQUE INDEX "uq_project_name" ON "public"."projects" ("name");
-- Create index "uq_project_public_id" to table: "projects"
CREATE UNIQUE INDEX "uq_project_public_id" ON "public"."projects" ("public_id");
