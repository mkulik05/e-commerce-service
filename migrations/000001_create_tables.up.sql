CREATE TABLE IF NOT EXISTS items (
    item_id bigint NOT NULL,
    item_name text NOT NULL,
    item_amount integer NOT NULL,
    item_price bigint NOT NULL,
    item_description text,
    times_bought integer NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_data (
    login text NOT NULL,
    pwd_hash text NOT NULL,
    user_id bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    "time" timestamp without time zone NOT NULL,
    order_id bigint NOT NULL,
    items_id bigint[] NOT NULL,
    delivery_addr text NOT NULL,
    user_id bigint
);