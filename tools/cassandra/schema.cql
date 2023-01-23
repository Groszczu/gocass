DROP KEYSPACE IF EXISTS shop;

CREATE KEYSPACE shop WITH REPLICATION = {
    'class': 'SimpleStrategy',
    'replication_factor': 3
};

USE shop;

CREATE TABLE users (
    id uuid PRIMARY KEY,
    name text
);

CREATE TABLE carts (
    user_id uuid,
    cart_id uuid,
    discount_code text,
    discount_percent int,
    PRIMARY KEY(user_id, cart_id)
);

CREATE TABLE cart_products (
    cart_id uuid,
    product_id uuid,
    name text,
    description text,
    price_in_cents int,
    quantity int,
    PRIMARY KEY(cart_id, product_id)
);

CREATE TABLE discount_codes (
    code text PRIMARY KEY,
    discount_percent int,
    usage_limit int,
    used_by set<uuid>
);

CREATE TABLE discount_code_usages (
    code text PRIMARY KEY,
    usage_count counter
);

CREATE TYPE order_product (
    product_id uuid,
    name text,
    description text,
    price_in_cents int,
    quantity int,
);

CREATE TABLE orders (
    cart_id uuid PRIMARY KEY,
    total_price_in_cents int,
    products map<uuid, frozen<order_product>>,
    status text
);

INSERT INTO discount_codes(code, discount_percent, usage_limit, used_by)
    VALUES ('abc', 10, 250, {});

INSERT INTO discount_codes(code, discount_percent, usage_limit, used_by)
    VALUES ('def', 15, 100, {});

