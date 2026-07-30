CREATE TABLE
    IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        email VARCHAR(320) UNIQUE NOT NULL,
        display_name VARCHAR(50) NOT NULL,
        password_hash TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now ()
    );