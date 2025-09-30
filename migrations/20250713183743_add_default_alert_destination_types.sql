INSERT INTO alert_destination_types(title, key, created_at)
VALUES('Internal Logger', 'internal.logger.error', CURRENT_TIMESTAMP),
('Webhook', 'external.webhook.post', CURRENT_TIMESTAMP);
