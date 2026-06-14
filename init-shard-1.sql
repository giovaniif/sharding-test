CREATE SEQUENCE orders_id_seq START 2 INCREMENT 3;
CREATE TABLE IF NOT EXISTS orders (
    id         BIGINT         DEFAULT nextval('orders_id_seq') PRIMARY KEY,
    product    VARCHAR(255)   NOT NULL,
    quantity   INT            NOT NULL,
    price      DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP      NOT NULL DEFAULT NOW()
);
