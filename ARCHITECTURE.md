# Architecture

## Stage 1 — Manual sharding

Three standalone PostgreSQL nodes, with sharding logic living entirely in the Go application layer.

### How it worked

**ID assignment — interleaved sequences:**
Each node used a `BIGSERIAL` sequence with a different start and a fixed increment of 3, so IDs were globally unique and non-overlapping:

| Node | Sequence start | IDs generated |
|------|---------------|---------------|
| postgres0 | 1 | 1, 4, 7, 10, … |
| postgres1 | 2 | 2, 5, 8, 11, … |
| postgres2 | 3 | 3, 6, 9, 12, … |

**Write routing — round-robin:**
An atomic counter in `ShardManager` cycled through the three connections on every insert, distributing load evenly.

**Read/delete routing — deterministic formula:**
Given any order ID, the owning node was computed as `(id - 1) % 3` — no lookup table needed.

**Fan-out reads:**
`GET /orders` spawned one goroutine per node, collected results over a channel, then merged and sorted by ID before responding.

**Infrastructure:**
Three `postgres:16-alpine` containers, each exposed on a different host port (5432, 5433, 5434), each initialised with its own `init-shard-N.sql`.

### Trade-offs
- All routing logic was application code that had to be kept in sync with the database layout.
- Adding or removing nodes required changing the formula, the sequences, and rebalancing data manually.
- Fan-out reads required N parallel connections and client-side sorting.

---

## Stage 2 — Citus (current)

Citus is a PostgreSQL extension that moves sharding inside the database. The application connects to a single **coordinator** node and issues plain SQL; Citus handles routing to **worker** nodes transparently.

### How it works

**Cluster topology:**

| Node | Role | Host port |
|------|------|-----------|
| coordinator | Single app entry point, holds shard metadata | 5432 |
| worker0 | Stores assigned shards | 5433 |
| worker1 | Stores assigned shards | 5434 |

**Distributed table:**
The `orders` table is declared as a distributed table with `id` as the distribution column:

```sql
SELECT create_distributed_table('orders', 'id');
```

Citus creates 32 logical shards (by default) and places them across the workers. Each shard is a real PostgreSQL table on the worker, named `orders_<shardid>`.

**Routing:**
- Writes: the coordinator hashes the `id` value and routes the row to the correct worker automatically.
- Point reads/deletes (`WHERE id = $1`): the coordinator resolves the shard and issues the query to a single worker.
- Full scans (`GET /orders`): the coordinator fans out to all workers in parallel and merges results — all within PostgreSQL, invisible to the application.

**Go application:**
The entire `ShardManager` was removed. The application holds a single `*sql.DB` connected to the coordinator and issues standard SQL — identical to connecting to a non-distributed database.

**Useful queries (run on coordinator):**
```sql
-- See all shards and which worker holds them
SELECT shardid, nodename, nodeport
FROM pg_dist_shard_placement
ORDER BY shardid;

-- See which shard a specific row lives on
SELECT *, get_shard_id_for_distribution_column('orders', id) AS shard_id
FROM orders;
```

On a worker (5433 or 5434), you can query individual shard tables directly:
```sql
SELECT * FROM orders_102008;
```

### Trade-offs
- Routing, rebalancing, and fault tolerance are handled by Citus, not application code.
- Scaling out requires adding worker nodes via `SELECT citus_add_node(...)` — no application changes needed.
- Requires the Citus extension; the coordinator is a single point of entry (though it can be made highly available separately).
