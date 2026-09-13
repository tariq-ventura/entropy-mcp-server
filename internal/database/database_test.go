package database

import (
	"context"
	"testing"
)

func TestOpenRejectsEmptyConnectionString(t *testing.T) {
	database, err := Open(context.Background(), Config{})
	if err == nil {
		t.Fatal("expected an error for an empty connection string")
	}
	if database != nil {
		t.Fatal("expected no database when configuration is invalid")
	}
}

func TestNilDatabaseCannotPingAndCanClose(t *testing.T) {
	var database *Database
	if err := database.Ping(context.Background()); err == nil {
		t.Fatal("expected ping to fail for an uninitialized database")
	}
	if err := database.Close(); err != nil {
		t.Fatalf("nil close should be safe: %v", err)
	}
}
