
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


CREATE TABLE users (
                       id text PRIMARY KEY NOT NULL,
                       full_name VARCHAR(100) NOT NULL,
                       username VARCHAR(50) UNIQUE NOT NULL,
                       email VARCHAR(255) NOT NULL,
                       password VARCHAR(255),
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

);

CREATE UNIQUE INDEX index_user_email
    ON users(email);

CREATE UNIQUE INDEX index_user_id
    ON users(id);


