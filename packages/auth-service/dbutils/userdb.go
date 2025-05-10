package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DBPATH = "user.sqlite"

type UserDB struct {
	conn *sql.DB
}

func Connect() (*UserDB, error) {
	conn, err := sql.Open("sqlite3", DBPATH)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %v", err)
	}

	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS user(
		id INTEGER PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`)
	if err != nil {
		return nil, err
	}

	return &UserDB{
		conn: conn,
	}, nil
}

func (s *UserDB) Ping() error {
	return s.conn.Ping()
}

func (s *UserDB) Close() {
	s.conn.Close()
}

func (s *UserDB) InsertUser(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %v", err)
	}

	_, err = tx.Exec(`INSERT INTO user(username, password) VALUES(?,?)`, username, HashString(password))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to exec: %v, failed to rollback tx: %v", err, rbErr)
		}
		return fmt.Errorf("failed to exec: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %v", err)
	}
	return nil
}

func (s *UserDB) FindUser(username string) (string, bool) {
	var user, pass string
	err := s.conn.QueryRow(`SELECT username,password FROM user WHERE username=?`, username).Scan(&user, &pass)
	if err != nil {
		return "", false
	}

	return pass, true
}
