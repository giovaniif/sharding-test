package db

import (
	"database/sql"
	"fmt"
	"sync/atomic"

	_ "github.com/lib/pq"
)

const NumShards = 3

type ShardManager struct {
	shards    [NumShards]*sql.DB
	insertIdx uint64
}

func ConnectShards(hosts [NumShards]string) (*ShardManager, error) {
	sm := &ShardManager{}
	for i, host := range hosts {
		dsn := fmt.Sprintf(
			"host=%s port=5432 user=postgres password=postgres dbname=orders sslmode=disable",
			host,
		)
		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("shard %d: %w", i, err)
		}
		if err := conn.Ping(); err != nil {
			return nil, fmt.Errorf("shard %d unreachable: %w", i, err)
		}
		sm.shards[i] = conn
	}
	return sm, nil
}

// ShardFor returns the DB connection that owns the given ID.
// Ownership is determined by interleaved sequences: (id-1) % NumShards == shard index.
func (sm *ShardManager) ShardFor(id int) *sql.DB {
	return sm.shards[(id-1)%NumShards]
}


// NextShard picks the next shard for an insert via round-robin.
func (sm *ShardManager) NextShard() (int, *sql.DB) {
	idx := int(atomic.AddUint64(&sm.insertIdx, 1)-1) % NumShards
	return idx, sm.shards[idx]
}

func (sm *ShardManager) All() [NumShards]*sql.DB {
	return sm.shards
}

func (sm *ShardManager) Close() {
	for _, s := range sm.shards {
		s.Close()
	}
}
