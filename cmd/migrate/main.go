package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"

	"voco/internal/app"
	migrator "voco/internal/pkg/migrate"
)

func main() {
	releaseMode := os.Getenv("VOCO_RELEASE_MODE")
	if releaseMode == "" {
		releaseMode = "prod"
	}

	cfg, err := app.MustLoadCfg(releaseMode)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.Pg.DSN() == "" {
		log.Fatal("VOCO_PG_URL (or VOCO_PG_HOST/...) is required")
	}

	migrationsPath := flag.String("path", "", "optional path to SQL migrations (default: embedded)")
	flag.Parse()

	dsn := cfg.Pg.MigrateDSN()
	var m *migrate.Migrate
	if *migrationsPath != "" {
		m, err = migrator.New(dsn, *migrationsPath)
	} else {
		m, err = migrator.NewFromFS(dsn)
	}
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("migrate source close: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("migrate db close: %v", dbErr)
		}
	}()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("usage: migrate [-path migrations] <up|down|steps N|version|force V>")
	}

	switch args[0] {
	case "up":
		if err := migrator.Up(m); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrate: up ok")
	case "down":
		if err := migrator.Down(m); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("migrate: down ok")
	case "steps":
		if len(args) < 2 {
			log.Fatal("usage: migrate steps <N>")
		}
		n, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("steps: invalid N: %v", err)
		}
		if err := migrator.Steps(m, n); err != nil {
			log.Fatalf("migrate steps: %v", err)
		}
		log.Printf("migrate: steps %d ok", n)
	case "version":
		v, dirty, err := migrator.Version(m)
		if err != nil {
			log.Fatalf("migrate version: %v", err)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
	case "force":
		if len(args) < 2 {
			log.Fatal("usage: migrate force <version>")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("force: invalid version: %v", err)
		}
		if err := migrator.Force(m, v); err != nil {
			log.Fatalf("migrate force: %v", err)
		}
		log.Printf("migrate: forced version %d", v)
	default:
		log.Fatalf("unknown command %q", args[0])
	}
}
