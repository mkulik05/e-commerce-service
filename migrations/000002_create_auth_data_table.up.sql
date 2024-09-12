CREATE TABLE IF NOT EXISTS auth_data (
    login text NOT NULL,
    pwd_hash text NOT NULL,
    user_id bigint NOT NULL
);