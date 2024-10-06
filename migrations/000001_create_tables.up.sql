CREATE TABLE IF NOT EXISTS items (
    it_id BIGSERIAL PRIMARY KEY,
    it_name TEXT NOT NULL,
    it_amount INTEGER NOT NULL,
    it_price BIGINT NOT NULL,
    it_desc TEXT,
    it_times_bought INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_data (
    ad_login TEXT UNIQUE NOT NULL,
    ad_pwd_hash TEXT NOT NULL,
    ad_user_id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS orders (
    or_id BIGSERIAL PRIMARY KEY,
    or_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    or_items text NOT NULL,
    or_delivery_addr TEXT NOT NULL,
    or_user_id BIGINT
);