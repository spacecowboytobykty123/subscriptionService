CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL CHECK (user_id > 0),
    plan_id INT REFERENCES subscription_plans(id) ON DELETE CASCADE,
    remaining_limit INT NOT NULL CHECK (remaining_limit > 0),
    expires_at TIMESTAMP NOT NULL CHECK (expires_at > NOW())
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'cancelled', 'expired', 'paused'))
)