package test_helpers

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"strings"
)
const CleanupFilePath = "../../database_data/init_db/clean.sql"
const InitDbFilePath = "../../database_data/init_db/test_init.sql"

func cleanUPDatabase(ctx context.Context, db *sqlx.DB) error {
	b, err := ioutil.ReadFile(CleanupFilePath)
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

func populateDatabase(ctx context.Context, db *sqlx.DB) error {
	b, err := ioutil.ReadFile(InitDbFilePath)
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

func PrepateDB(ctx context.Context, db *sqlx.DB) error {
	err := cleanUPDatabase(ctx, db)
	if err != nil {
		return fmt.Errorf("cleanup err: %v", err.Error())
	}

	err = populateDatabase(ctx, db)
	if err != nil {
		return fmt.Errorf("populate err: %v", err.Error())
	}
	return nil
}
