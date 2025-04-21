CREATE TABLE IF NOT EXIST subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    plan_id INT REFERENCES subscription_plans(id) ON DELETE CASCADE,
    remaining_limit INT NOT NULL CHECK (remaining_limit > 0),
    expires_at TIMESTAMP NOT NULL CHECK (expires_at > NOW())
)