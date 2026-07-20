//go:build ignore

// Run from the Inventory service root:
//   go run ./scripts/migration.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/term"
)

// ── SQL ───────────────────────────────────────────────────────────────────────

// Ensure txn_log exists with the expected shape.
const sqlCreateTxnLog = `
CREATE TABLE IF NOT EXISTS txn_log (
	auto_id     BIGSERIAL   NOT NULL,
	trans_date  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	description VARCHAR     NOT NULL DEFAULT '',
	txn_id      VARCHAR     NOT NULL,
	location_id BIGINT      NOT NULL DEFAULT 0,
	item_code   VARCHAR     NOT NULL,
	qty_in      FLOAT       NOT NULL DEFAULT 0,
	qty_out     FLOAT       NOT NULL DEFAULT 0,
	CONSTRAINT pk_txn_log      PRIMARY KEY (description, txn_id, item_code),
	CONSTRAINT unique_txn_log  UNIQUE (auto_id)
);`

// Disable DELETE on txn_log via a rule — any DELETE becomes a no-op.
// Rules fire before the statement so no rows are ever removed.
const sqlDisableDelete = `
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_rules
		WHERE tablename = 'txn_log' AND rulename = 'no_delete_txn_log'
	) THEN
		EXECUTE $r$
			CREATE RULE no_delete_txn_log AS
				ON DELETE TO txn_log
				DO INSTEAD NOTHING
		$r$;
	END IF;
END;
$$;`

// Indexes for fast lookups on the most common query predicates.
const sqlIndexes = `
CREATE INDEX IF NOT EXISTS idx_txn_log_item_code
	ON txn_log (item_code);

CREATE INDEX IF NOT EXISTS idx_txn_log_location_id
	ON txn_log (location_id);

CREATE INDEX IF NOT EXISTS idx_txn_log_item_location
	ON txn_log (item_code, location_id);

CREATE INDEX IF NOT EXISTS idx_txn_log_trans_date
	ON txn_log (trans_date DESC);

CREATE INDEX IF NOT EXISTS idx_txn_log_item_date
	ON txn_log (item_code, trans_date DESC);`

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load(".env")

	host   := envOr("DB_HOST", "localhost")
	port   := envOr("DB_PORT", "5432")
	dbName := envOr("DB_NAME", "inventory")

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Inventory Service — txn_log Migration")
	fmt.Printf("  DB : %s:%s/%s\n", host, port, dbName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	dbUser := prompt("Database user: ")
	dbPass := promptPassword("Database password: ")
	fmt.Println()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, host, port, dbName)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close(ctx)
	fmt.Println("Connected.")
	fmt.Println()

	steps := []struct {
		name string
		sql  string
	}{
		{"Create txn_log table (if not exists)", sqlCreateTxnLog},
		{"Disable DELETE on txn_log via rule", sqlDisableDelete},
		{"Create indexes on txn_log", sqlIndexes},
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %s … ", i+1, len(steps), step.name)
		if _, err := conn.Exec(ctx, step.sql); err != nil {
			fmt.Println("FAILED")
			log.Fatalf("error: %v", err)
		}
		fmt.Println("OK")
	}

	fmt.Println()
	fmt.Println("Migration complete.")
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func prompt(label string) string {
	fmt.Print(label)
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	return strings.TrimSpace(sc.Text())
}

func promptPassword(label string) string {
	fmt.Print(label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatalf("failed to read password: %v", err)
	}
	return string(b)
}
