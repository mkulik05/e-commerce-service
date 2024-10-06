CREATE TABLE IF NOT EXISTS items (
    it_id SERIAL PRIMARY KEY,
    it_name TEXT NOT NULL,
    it_amount INTEGER NOT NULL,
    it_price INTEGER NOT NULL,
    it_desc TEXT,
    it_times_bought INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_data (
    ad_user_id SERIAL PRIMARY KEY,
    ad_login varchar(50) UNIQUE NOT NULL,
    ad_pwd_hash varchar(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    or_id SERIAL PRIMARY KEY,
    or_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    or_delivery_addr varchar(250) NOT NULL,
    or_user_id INTEGER REFERENCES auth_data (ad_user_id)
);

CREATE TABLE IF NOT EXISTS m2m_order_items (
    oi_order_id INTEGER REFERENCES orders (or_id),
    oi_item_id INTEGER REFERENCES items (it_id),
    oi_item_amount INTEGER NOT NULL,
    PRIMARY KEY (oi_order_id, oi_item_id)
);