package server

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Autumn-27/artex/db"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestMain acquires a PostgreSQL advisory lock (7337741002) for the entire
// server test suite so cross-package DELETE cleanup races with db/agent
// packages are avoided when running `go test ./...`.
func TestMain(m *testing.M) {
	dsn, _, err := db.DSN()
	if err != nil {
		os.Exit(m.Run())
	}
	conn, err := sql.Open("pgx", dsn)
	if err != nil || conn.Ping() != nil {
		os.Exit(m.Run())
	}
	defer conn.Close()
	if _, err := conn.Exec(`SELECT pg_advisory_lock(7337741002)`); err != nil {
		os.Exit(m.Run())
	}
	defer conn.Exec(`SELECT pg_advisory_unlock(7337741002)`) //nolint:errcheck
	// Prior package runs (and crashed tests) leave non-terminal tasks in the shared
	// ARTEX_PG_DSN database. New() restores them and starts workers into the next
	// test's TempDir — mark them terminal first so suites start from a quiet DB.
	_, _ = conn.Exec(`UPDATE tasks SET status='done', paused=true, queued=false WHERE status IS NULL OR status NOT IN ('done','failed','timeout')`)
	os.Exit(m.Run())
}
