CREATE TABLE IF NOT EXISTS items (
    item_id BIGSERIAL PRIMARY KEY,
    item_name TEXT NOT NULL,
    item_amount INTEGER NOT NULL,
    item_price BIGINT NOT NULL,
    item_description TEXT,
    times_bought INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_data (
    login TEXT UNIQUE NOT NULL,
    pwd_hash TEXT NOT NULL,
    user_id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS orders (
    "time" TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    order_id BIGSERIAL PRIMARY KEY,
    items text NOT NULL,
    delivery_addr TEXT NOT NULL,
    user_id BIGINT
);