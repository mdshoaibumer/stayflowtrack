package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// Connect to default postgres database to create stayflow_track
	adminDSN := "postgres://stayflow@localhost:5433/postgres?sslmode=disable"
	adminPool, err := pgxpool.New(ctx, adminDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect admin: %v\n", err)
		os.Exit(1)
	}

	// Drop and recreate the database
	adminPool.Exec(ctx, "DROP DATABASE IF EXISTS stayflow_track")
	_, err = adminPool.Exec(ctx, "CREATE DATABASE stayflow_track")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create db: %v\n", err)
		os.Exit(1)
	}
	adminPool.Close()
	fmt.Println("✓ Database stayflow_track created")

	// Connect to stayflow_track and run migrations
	dsn := "postgres://stayflow@localhost:5433/stayflow_track?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrationsDir := filepath.Join(".", "migrations")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read migrations: %v\n", err)
		os.Exit(1)
	}

	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, filepath.Join(migrationsDir, f.Name()))
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		sql, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", f, err)
			os.Exit(1)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			fmt.Fprintf(os.Stderr, "exec %s: %v\n", filepath.Base(f), err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ %s\n", filepath.Base(f))
	}

	fmt.Println("✓ All migrations applied")
}
