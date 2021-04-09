package main

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
PRAGMA foreign_keys;

CREATE TABLE IF NOT EXISTS Usernames (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS Guilds (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	guild TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS Channels (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	channel TEXT NOT NULL UNIQUE,
	guild_id INTEGER NOT NULL,
	FOREIGN KEY (guild_id) REFERENCES Guilds(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Follows (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username_id INTEGER NOT NULL,
	channel_id INTEGER NOT NULL,
	history TEXT NOT NULL DEFAULT '',
	UNIQUE(username_id, channel_id),
	FOREIGN KEY (username_id) REFERENCES Usernames(id) ON DELETE CASCADE,
	FOREIGN KEY (channel_id) REFERENCES Channels(id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS CleanGuilds
AFTER DELETE ON Channels
WHEN (SELECT COUNT(*) FROM Channels WHERE guild_id = OLD.guild_id) = 0
BEGIN
	DELETE FROM Guilds WHERE id = OLD.guild_id;
END;

CREATE TRIGGER IF NOT EXISTS CleanChannels
AFTER DELETE ON Follows
WHEN (SELECT COUNT(*) FROM Follows WHERE channel_id = OLD.channel_id) = 0
BEGIN
	DELETE FROM Channels WHERE id = OLD.channel_id;
END;

CREATE TRIGGER IF NOT EXISTS CleanUsernames
AFTER DELETE ON Follows
WHEN (SELECT COUNT(*) FROM Follows WHERE username_id = OLD.username_id) = 0
BEGIN
	DELETE FROM Usernames WHERE id = OLD.username_id;
END;
`

type DB struct {
	lock sync.RWMutex
	db   *sql.DB
}

func OpenSQLDB(driver, source string) (*DB, error) {
	sqlDB, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}

	db := &DB{db: sqlDB}
	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.db.Close()
}

func (db *DB) init() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	return tx.Commit()
}

func (db *DB) Follow(username, channel, guild string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT OR IGNORE INTO Usernames(username) VALUES (?)", username)
	if err != nil {
		return fmt.Errorf("failed to add username '%s': %v", username, err)
	}

	_, err = tx.Exec("INSERT OR IGNORE INTO Guilds(guild) VALUES (?)", guild)
	if err != nil {
		return fmt.Errorf("failed to add guild '%s': %v", guild, err)
	}

	_, err = tx.Exec(`INSERT OR IGNORE INTO Channels(channel, guild_id)
		VALUES (?, (SELECT id FROM Guilds WHERE guild = ?))`, channel, guild)
	if err != nil {
		return fmt.Errorf("failed to add channel '%s': %v", channel, err)
	}

	_, err = tx.Exec(`INSERT INTO Follows(username_id, channel_id)
		VALUES (
			(SELECT id FROM Usernames WHERE username = ?),
			(SELECT id FROM Channels WHERE channel = ?)
		)`, username, channel)
	if err != nil {
		return fmt.Errorf("failed to follow username '%s' in channel '%s': %v", username, channel, err)
	}

	return tx.Commit()
}

func (db *DB) Unfollow(username, channel string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	_, err := db.db.Exec(`DELETE FROM Follows WHERE
		username_id = (SELECT id FROM Usernames WHERE username = ?)
		and
		channel_id = (SELECT id FROM Channels WHERE channel = ?)`, username, channel)

	if err != nil {
		return fmt.Errorf("failed to unfollow username '%s' in channel '%s': %v", username, channel, err)
	}

	return nil
}

func (db *DB) Following(channel string) ([]string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var following []string

	rows, err := db.db.Query(`SELECT u.username
		FROM Follows f INNER JOIN Usernames u INNER JOIN Channels c
		ON f.username_id = u.id and f.channel_id = c.id
		WHERE c.channel = ?`, channel)
	if err != nil {
		return following, fmt.Errorf("failed to get list of usernames for channel '%s': %v", channel, err)
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		following = append(following, username)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return following, nil
}
