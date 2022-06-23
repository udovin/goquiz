package migrations_test

import (
	"os"
	"testing"

	"github.com/udovin/goquiz/config"
	"github.com/udovin/goquiz/core"
	"github.com/udovin/goquiz/migrations"
)

func TestMigrations(t *testing.T) {
	cfg := config.Config{
		DB: config.DB{
			Options: config.SQLiteOptions{Path: ":memory:"},
		},
		Security: &config.Security{
			PasswordSalt: "qwerty123",
		},
	}
	c, err := core.NewCore(cfg)
	if err != nil {
		t.Fatal("Error:", err)
	}
	c.SetupAllStores()
	if err := migrations.Apply(c); err != nil {
		t.Fatal("Error:", err)
	}
	if err := migrations.Unapply(c, true); err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPostgresMigrations(t *testing.T) {
	pgHost, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		t.Skip()
	}
	pgPort, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok {
		t.Skip()
	}
	cfg := config.Config{
		DB: config.DB{
			Options: config.PostgresOptions{
				Hosts:    []string{pgHost + ":" + pgPort},
				User:     "postgres",
				Password: "postgres",
				Name:     "postgres",
				SSLMode:  "disable",
			},
		},
		Security: &config.Security{
			PasswordSalt: "qwerty123",
		},
	}
	c, err := core.NewCore(cfg)
	if err != nil {
		t.Fatal("Error:", err)
	}
	c.SetupAllStores()
	if err := migrations.Apply(c); err != nil {
		t.Fatal("Error:", err)
	}
	if err := migrations.Unapply(c, true); err != nil {
		t.Fatal("Error:", err)
	}
}
