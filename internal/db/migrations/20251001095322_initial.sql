-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id varchar PRIMARY KEY,
    email varchar UNIQUE NOT NULL,
    role varchar NOT NULL,
    hashed_password text,
    last_logged_at bigint,
    created_at bigint,
    social_account boolean DEFAULT false,
    social_id varchar,
    social_provider varchar
);

CREATE TABLE user_logs (
    id varchar PRIMARY KEY,
    -- user_id varchar REFERENCES users(id)
    user_email varchar NOT NULL,
    user_agent text NOT NULL,
    ip varchar NOT NULL,
    logged_at bigint NOT NULL
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TABLE user_logs;
-- +goose StatementEnd
