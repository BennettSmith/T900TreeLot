// Command dbping exits 0 when a PostgreSQL URL accepts connections.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/troop900/treelot/internal/platform/postgres"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dbping <database-url>")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db, err := postgres.Open(ctx, os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = db.Close()
}
