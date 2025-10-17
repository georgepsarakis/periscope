-- Modify "alert_destination_notifications" table
ALTER TABLE "public"."alert_destination_notifications" ADD COLUMN "last_error" text NULL, ADD COLUMN "attempted_at" timestamptz NULL;
