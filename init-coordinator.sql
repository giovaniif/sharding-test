CREATE EXTENSION IF NOT EXISTS citus;
SELECT citus_set_coordinator_host('coordinator', 5432);
SELECT citus_add_node('worker0', 5432);
SELECT citus_add_node('worker1', 5432);

CREATE TABLE IF NOT EXISTS orders (
    id         BIGSERIAL      PRIMARY KEY,
    product    VARCHAR(255)   NOT NULL,
    quantity   INT            NOT NULL,
    price      DECIMAL(10,2)  NOT NULL,
    created_at TIMESTAMP      NOT NULL DEFAULT NOW()
);

SELECT create_distributed_table('orders', 'id');
