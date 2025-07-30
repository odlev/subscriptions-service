CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name TEXT NOT NULL,
    price DECIMAL NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    start_date DATE NOT NULL,
    end_date DATE,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
