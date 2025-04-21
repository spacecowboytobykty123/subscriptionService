DROP INDEX IF EXIST idx_subscriptions_user_id on subscriptions(user_id);
DROP INDEX IF idx_subscriptions_expires_at on subscriptions(expires_at);
