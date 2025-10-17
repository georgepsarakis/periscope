UPDATE alert_destination_types
SET title = 'Generic Webhook', key='external.webhook.generic'
WHERE key='external.webhook.post';

INSERT INTO alert_destination_types(title, key, created_at)
VALUES('Slack Webhook', 'external.webhook.slack', CURRENT_TIMESTAMP);

