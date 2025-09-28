package models

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func newTestDB(t *testing.T) *sql.DB {
	_ = godotenv.Load("../../.env")

	dbPass, ok := os.LookupEnv("MYSQL_PASSWORD_TEST")
	if !ok {
		t.Fatal("MYSQL_PASSWORD_TEST environment variable not set")
	}

	dsn := fmt.Sprintf("test_web:%s@/test_snippetbox?parseTime=true&multiStatements=true", dbPass)

	t.Log(os.Getenv("MYSQL_PASSWORD_TEST"))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
	})

	return db
}
