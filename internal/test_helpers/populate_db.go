package test_helpers

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"strings"
)

type Config struct {
	InitFilePath    string
	CleanUpFilePath string
}

func cleanUPDatabase(ctx context.Context, db *sqlx.DB, filePath string) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open sql file %s", err.Error())
	}

	statements := strings.Split(string(b), ";")
	for _, q := range statements {
		_, err := db.ExecContext(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to execute %s", err.Error())
		}
	}

	return nil
}

func populateDatabase(ctx context.Context, db *sqlx.DB, filePath string) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open sql file %s", err.Error())
	}

	statements := strings.Split(string(b), ";")
	for _, q := range statements {
		_, err := db.ExecContext(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to execute %s", err.Error())
		}
	}

	return nil
}

func PrepareDB(ctx context.Context, db *sqlx.DB, cfg Config) error {
	if cfg.CleanUpFilePath == "" || cfg.InitFilePath == "" {
		return fmt.Errorf("got empty path to init_test.sql or cleanup.sql file")
	}

	err := cleanUPDatabase(ctx, db, cfg.CleanUpFilePath)
	if err != nil {
		return fmt.Errorf("cleanup err: %v", err.Error())
	}

	err = populateDatabase(ctx, db, cfg.InitFilePath)
	if err != nil {
		return fmt.Errorf("populate err: %v", err.Error())
	}

	return nil
}
