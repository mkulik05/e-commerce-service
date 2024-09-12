CREATE TABLE IF NOT EXISTS orders (
    "time" timestamp without time zone NOT NULL,
    order_id bigint NOT NULL,
    items_id bigint[] NOT NULL,
    delivery_addr text NOT NULL,
    user_id bigint
);