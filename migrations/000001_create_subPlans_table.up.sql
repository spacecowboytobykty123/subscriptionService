CREATE TABLE IF NOT EXISTS subscription_plans (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(50) NOT NULL,
    rental_limit INT NOT NULL check ( rental_limit > 0 ),
    price INT NOT NULL check ( price > 0 ),
    duration_months INT NOT NULL check ( duration_months > 0 ),
    created_at TIMESTAMP DEFAULT NOW()
    status VARCHAR(20) NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'cancelled', 'expired', 'paused'))
);