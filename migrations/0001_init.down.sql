-- +migrate Down
DROP TABLE IF EXISTS clicks;
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS custom_domains;
DROP TABLE IF EXISTS link_revisions;
DROP TABLE IF EXISTS notifications;

DROP INDEX IF EXISTS idx_clicks_org_name;
DROP INDEX IF EXISTS idx_notifications_user_id;