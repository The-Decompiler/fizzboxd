package main

import (
	"sort"
	"testing"
)

func TestDB(t *testing.T) {
	db, err := OpenSQLDB("sqlite3", "test.db")
	if err != nil {
		t.Fatalf("failed to open database: %v\n", err)
	}

    if err := db.Follow("username1", "channel1", "guild1"); err != nil {
        t.Fatalf("failed to insert test follow values: %v", err)
    }

    if err := db.Follow("username1", "channel1", "guild1"); err == nil {
        t.Error("unique constraint did not work")
    }

    if err := db.Follow("username2", "channel1", "guild1"); err != nil {
        t.Fatalf("failed to insert test follow values: %v", err)
    }

	// Same guild, different channel
    if err := db.Follow("username3", "channel2", "guild1"); err != nil {
        t.Errorf("failed to insert test follow values: %v", err)
    }

	// Different guild, different channel
    if err := db.Follow("username4", "channel3", "guild2"); err != nil {
        t.Errorf("failed to insert test follow values: %v", err)
    }

    following, err := db.Following("channel1")
    if err != nil {
        t.Errorf("failed to get list of usernames: %v", err)
    }

    sort.Strings(following)
    if len(following) != 2 || following[0] != "username1" || following[1] != "username2" {
	    t.Errorf("list of usernames is wrong, expected [username1 username2] got %v", following)
    }

    if err := db.Unfollow("username1", "channel1"); err != nil {
        t.Fatalf("failed to unfollow: %v", err)
    }

    following, err = db.Following("channel1")
    if err != nil {
        t.Errorf("failed to get list of usernames: %v", err)
    }

    if len(following) != 1 || following[0] != "username2" {
	    t.Errorf("list of usernames is wrong, expected [username2] got %v", following)
    }
}
